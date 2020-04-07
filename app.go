package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	firebase "firebase.google.com/go"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	oauthapi "google.golang.org/api/oauth2/v2"
	plus "google.golang.org/api/plus/v1"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/gofrs/uuid"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

var templates = template.Must(template.ParseFiles("templates/index.html", "templates/show.html"))

var (
	OAuthConfig  *oauth2.Config
	SessionStore sessions.Store
)

const (
	defaultSessionID        = "default"
	oauthFlowRedirectKey    = "redirect"
	oauthTokenSessionKey    = "oauth_token"
	googleProfileSessionKey = "google_profile"
)

type File struct {
	Link     string
	Path     string
	FavCount int
}

type Profile struct {
	ID, DisplayName, ImageURL string
}

//Userテーブル準備
type User struct {
	gorm.Model
	Name string
}

func CountFavorite(path string) (count int) {
	iter := storeClient.Collection("favorites").Where("filepath", "==", path).Documents(context.Background())
	_count := 0
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		_count++
	}
	return _count
}

func configureOAuthClient(clientID, clientSecret string) *oauth2.Config {
	redirectURL := os.Getenv("OAUTH2_CALLBACK")
	if redirectURL == "" {
		redirectURL = "http://localhost:8888/oauth2callback"
	}
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func init() {
	OAuthConfig = configureOAuthClient(os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"))
	// Configure storage method for session-wide information.
	// Update "something-very-secret" with a hard to guess string or byte sequence.
	// 乱数生成
	b := make([]byte, 48)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		panic(err)
	}
	str := strings.TrimRight(base32.StdEncoding.EncodeToString(b), "=")
	cookieStore := sessions.NewCookieStore([]byte(str))
	cookieStore.Options = &sessions.Options{
		HttpOnly: true,
	}
	SessionStore = cookieStore
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	d := struct {
		AuthEnabled bool
		Profile     *oauthapi.Userinfoplus
		LoginURL    string
		LogoutURL   string
	}{
		AuthEnabled: OAuthConfig != nil,
		LoginURL:    "/login?redirect=" + r.URL.RequestURI(),
		LogoutURL:   "/logout?redirect=" + r.URL.RequestURI(),
	}
	if d.AuthEnabled {
		// Ignore any errors.
		d.Profile = profileFromSession(r)
	}
	fmt.Print(d.Profile)
	ctx := context.Background()
	it := bucket.Objects(ctx, nil)
	folders := []string{}
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return
		}
		end := strings.Index(attrs.Name, "/")
		folders = append(folders, attrs.Name[:end])
	}
	m := make(map[string]bool)
	uniqFolders := []string{}

	for _, ele := range folders {
		if !m[ele] {
			m[ele] = true
			uniqFolders = append(uniqFolders, ele)
		}
	}

	data := map[string]interface{}{"Title": "index", "folders": uniqFolders, "profile": d.Profile}
	renderTemplate(w, "index", data)
}

func ShowHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/show/"):]
	ctx := context.Background()
	prefix := id + "/"
	it := bucket.Objects(ctx, &storage.Query{
		Prefix: prefix,
	})
	list := []File{}
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return
		}
		if attrs.Name[len(attrs.Name)-1:] != "/" {
			path := attrs.Name
			rep := regexp.MustCompile(`\W`)
			path = rep.ReplaceAllString(path, "")
			count := CountFavorite(path)
			list = append(list, File{"https://storage.googleapis.com/go-pictures.appspot.com/" + attrs.Name, path, count})
		}
	}
	data := struct {
		List []File
	}{
		List: list,
	}
	renderTemplate(w, "show", data)
	return
}

func SequenceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	it := bucket.Objects(ctx, nil)
	links := []string{}
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return
		}
		if strings.Contains(attrs.Name, "sequence/D") {
			links = append(links, "https://storage.googleapis.com/go-pictures.appspot.com/"+attrs.Name)
		}
	}
	data := map[string]interface{}{"Title": "CSFshow", "links": links}
	renderTemplate(w, "show", data)
	return
}

// UploadHandler は Cloud Storage for Firebase にアップロードするための関数です
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Allowed POST method only", http.StatusMethodNotAllowed)
		return
	}
	r.ParseMultipartForm(32 << 20)
	fhs := r.MultipartForm.File["upload-firebase"]
	ctx := context.Background()
	for _, fh := range fhs {
		f, err := fh.Open()
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		buf := bytes.NewBuffer(nil)
		buf2 := bytes.NewBuffer(nil)
		bufWriter := io.MultiWriter(buf, buf2)
		io.Copy(bufWriter, f)

		if fh.Filename == "" {
			continue
		}
		remoteFilename := fh.Filename

		contentType := ""
		fileData, err := ioutil.ReadAll(buf)
		if err != nil {
			contentType = "application/octet-stream"
		} else {
			contentType = http.DetectContentType(fileData)
		}
		folderName := r.FormValue("album")
		remotePath := path.Join(folderName, remoteFilename)
		writer := bucket.Object(remotePath).NewWriter(ctx)
		writer.ObjectAttrs.ContentType = contentType
		writer.ObjectAttrs.CacheControl = "no-cache"
		writer.ObjectAttrs.ACL = []storage.ACLRule{
			{
				Entity: storage.AllUsers,
				Role:   storage.RoleReader,
			},
		}
		defer writer.Close()
		if _, err = writer.Write(fileData); err != nil {
			log.Fatalln(err)
		}
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	if err := templates.ExecuteTemplate(w, tmpl+".html", data); err != nil {
		log.Fatalln("Unable to execute template.")
	}
}

func writeImageWithTemplate(w http.ResponseWriter, tmpl string, img *image.Image) {
	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, *img, nil); err != nil {
		log.Fatalln("Unable to encode image.")
	}
	str := base64.StdEncoding.EncodeToString(buffer.Bytes())
	data := map[string]interface{}{"Title": tmpl, "Image": str}
	renderTemplate(w, tmpl, data)
}
func firebaseInit() (bkthdl *storage.BucketHandle, sc *firestore.Client) {
	if os.Getenv("GO_ENV") == "dev" {
		err := godotenv.Load("envfiles/dev.env")
		if err != nil {
			log.Fatalln(err)
		}
	}
	config := &firebase.Config{
		StorageBucket: "go-pictures.appspot.com",
	}
	ctx := context.Background()
	jsonStr := fmt.Sprintf(`{
	  "type": "service_account",
	  "project_id": "go-pictures",
	  "private_key_id": "%s",
	  "private_key": "%s",
	  "client_email": "firebase-adminsdk-of90d@go-pictures.iam.gserviceaccount.com",
	  "client_id": "%s",
	  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
	  "token_uri": "https://oauth2.googleapis.com/token",
	  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
	  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/firebase-adminsdk-of90d%%40go-pictures.iam.gserviceaccount.com"
	}`, os.Getenv("PRIVATE_KEY_ID"), os.Getenv("PRIVATE_KEY"), os.Getenv("CLIENT_ID"))
	opt := option.WithCredentialsJSON([]byte(jsonStr))

	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Storage(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	bucket, err := client.DefaultBucket()
	if err != nil {
		log.Fatalln(err)
	}

	storeClient, err := app.Firestore(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	return bucket, storeClient
}

func FavoriteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.ParseForm()
	path := r.FormValue("path")
	_, _, err := storeClient.Collection("favorites").Add(ctx, map[string]interface{}{
		"uid":      55,
		"filepath": path,
	})
	if err != nil {
		log.Fatalf("Failed adding alovelace: %v", err)
	}
	count := CountFavorite(path)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(strconv.Itoa(count)))
}

var bucket, storeClient = firebaseInit()

//GORM関連のコード
func GetDBConn() *gorm.DB {
	db, err := gorm.Open(GetDBConfig())
	if err != nil {
		fmt.Println(err)
	}

	db.LogMode(true)
	return db
}

func GetDBConfig() (string, string) {
	DBMS := "mysql"
	USER := os.Getenv("MYSQL_USER")
	PASS := os.Getenv("MYSQL_PASSWORD")
	PROTOCOL := "tcp(mysql:3306)"
	DBNAME := os.Getenv("MYSQL_DATABASE")
	CONNECT := USER + ":" + PASS + "@" + PROTOCOL + "/" + DBNAME
	return DBMS, CONNECT
}

// validateRedirectURL checks that the URL provided is valid.
// If the URL is missing, redirect the user to the application's root.
// The URL must not be absolute (i.e., the URL must refer to a path within this
// application).
func validateRedirectURL(path string) (string, error) {
	if path == "" {
		return "/", nil
	}

	// Ensure redirect URL is valid and not pointing to a different server.
	parsedURL, err := url.Parse(path)
	if err != nil {
		return "/", err
	}
	if parsedURL.IsAbs() {
		return "/", errors.New("URL must not be absolute")
	}
	return path, nil
}

// loginHandler initiates an OAuth flow to authenticate the user.
func loginHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := uuid.Must(uuid.NewV4()).String()

	oauthFlowSession, err := SessionStore.New(r, sessionID)
	if err != nil {
		fmt.Println(err)
	}
	oauthFlowSession.Options.MaxAge = 10 * 60 // 10 minutes

	redirectURL, err := validateRedirectURL(r.FormValue("redirect"))
	if err != nil {
		fmt.Println(err)
	}
	oauthFlowSession.Values[oauthFlowRedirectKey] = redirectURL

	if err := oauthFlowSession.Save(r, w); err != nil {
		fmt.Println(err)
	}

	// Use the session ID for the "state" parameter.
	// This protects against CSRF (cross-site request forgery).
	// See https://godoc.org/golang.org/x/oauth2#Config.AuthCodeURL for more detail.
	url := OAuthConfig.AuthCodeURL(sessionID, oauth2.ApprovalForce,
		oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusFound)
}

func OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	oauthFlowSession, err := SessionStore.Get(r, r.FormValue("state"))
	if err != nil {
		fmt.Println(err)
	}

	redirectURL, ok := oauthFlowSession.Values[oauthFlowRedirectKey].(string)
	// Validate this callback request came from the app.
	if !ok {
		fmt.Println(err)
	}

	session, err := SessionStore.New(r, defaultSessionID)
	if err != nil {
		fmt.Println(err)
	}

	//パラメータからアクセスコードを読み取り
	code := r.URL.Query()["code"]
	if code == nil || len(code) == 0 {
		fmt.Fprint(w, "Invalid Parameter")
	}
	//いろいろライブラリが頑張って
	ctx := context.Background()
	tok, err := OAuthConfig.Exchange(ctx, code[0])
	if err != nil {
		fmt.Fprintf(w, "OAuth Error:%v", err)
	}
	//APIクライアントができて
	client := OAuthConfig.Client(ctx, tok)
	//Userinfo APIをGetしてDoして
	svr, _ := oauthapi.New(client)
	ui, err := svr.Userinfo.Get().Do()
	if err != nil {
		fmt.Fprintf(w, "OAuth Error:%v", err)
	} else {
		//メールアドレス取得！
		fmt.Println(ui.Email)
	}
	session.Values[oauthTokenSessionKey] = tok
	// Strip the profile to only the fields we need. Otherwise the struct is too big.
	session.Values[googleProfileSessionKey] = ui
	if err := session.Save(r, w); err != nil {
		fmt.Println(err)
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// stripProfile returns a subset of a plus.Person.
func stripProfile(p *plus.Person) *Profile {
	return &Profile{
		ID:          p.Id,
		DisplayName: p.DisplayName,
		ImageURL:    p.Image.Url,
	}
}

// profileFromSession retreives the Google+ profile from the default session.
// Returns nil if the profile cannot be retreived (e.g. user is logged out).
func profileFromSession(r *http.Request) *oauthapi.Userinfoplus {
	session, err := SessionStore.Get(r, defaultSessionID)
	if err != nil {
		return nil
	}

	tok, ok := session.Values[oauthTokenSessionKey].(*oauth2.Token)
	if !ok || !tok.Valid() {
		return nil
	}
	ui, ok := session.Values[googleProfileSessionKey].(*oauthapi.Userinfoplus)
	if !ok {
		return nil
	}
	return ui
}

func main() {
	db := GetDBConn()
	db.AutoMigrate(&User{})

	http.Handle("/statics/", http.StripPrefix("/statics/", http.FileServer(http.Dir("statics/"))))
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/upload", UploadHandler)
	http.HandleFunc("/show/", ShowHandler)
	http.HandleFunc("/sequence", SequenceHandler)
	http.HandleFunc("/favorite", FavoriteHandler)
	http.HandleFunc("/oauth2callback", OAuthCallbackHandler)

	port := os.Getenv("PORT")
	http.ListenAndServe(":"+port, nil)
}
