package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/jinzhu/gorm"
   _ "github.com/go-sql-driver/mysql"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
	"text/template"

	firebase "firebase.google.com/go"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/joho/godotenv"
)

var templates = template.Must(template.ParseFiles("templates/index.html", "templates/show.html"))

type File struct {
	Link     string
	Path     string
	FavCount int
}

//Userテーブル準備
type User struct {
    Id int64 `gorm:"primary_key"`
    Name string `sql:"size:255"`
    CreatedAt time.Time
    UpdatedAT time.Time
    DeletedAt time.Time
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

func IndexHandler(w http.ResponseWriter, r *http.Request) {
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

	data := map[string]interface{}{"Title": "index", "folders": uniqFolders}
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
   USER := "root"
   PASS := ""
   PROTOCOL := ""
   DBNAME := "gopicture"
   OPTION := "charset=utf8&parseTime=True&loc=Local"

   CONNECT := USER + ":" + PASS + "@" + PROTOCOL + "/" + DBNAME + "?" + OPTION

   return DBMS, CONNECT
}

func main() {
	//データベース接続、テーブル作成
	db := GetDBConn()
  db.AutoMigrate(&User{})

	port := os.Getenv("PORT")
	http.Handle("/statics/", http.StripPrefix("/statics/", http.FileServer(http.Dir("statics/"))))
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/upload", UploadHandler)
	http.HandleFunc("/show/", ShowHandler)
	http.HandleFunc("/sequence", SequenceHandler)
	http.HandleFunc("/favorite", FavoriteHandler)
	http.ListenAndServe(":"+port, nil)
}
