package main

import(
  "fmt"
  "bytes"
  "io"
  "path"
  "log"
  "os"
  "strings"
  "io/ioutil"
  "net/http"
  "math/rand"
  "time"
  "encoding/json"
  "golang.org/x/net/context"
  "cloud.google.com/go/storage"
  "github.com/dgrijalva/jwt-go"

  "gopicture/models"
)
// UploadHandler は Cloud Storage for Firebase にアップロードするための関数です
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Allowed POST method only", http.StatusMethodNotAllowed)
		return
	}
	r.ParseMultipartForm(32 << 20)
	fhs := r.MultipartForm.File["upload-firebase"]
	ctx := context.Background()
  remoteFolderName := RandString(32)
  inputFolderName := r.FormValue("album")
  var pictures []models.Picture
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
		remoteFilename := RandString(32)
    pictures = append(pictures, models.Picture{ Name:remoteFilename })
		contentType := ""
		fileData, err := ioutil.ReadAll(buf)
		if err != nil {
			contentType = "application/octet-stream"
		} else {
			contentType = http.DetectContentType(fileData)
		}
		remotePath := path.Join(remoteFolderName, remoteFilename)
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
  album := models.Album{ Name: inputFolderName, Hash: remoteFolderName, Pictures: pictures }
  err := album.Create()
  if err != nil{
    fmt.Println(err)
  }
  user := new(models.User)
  ui, _ := profileFromSession(r)
  err = user.FirstOrCreate(ui.Email, ui.Name)
  if err != nil {
      print(err)
  }
  err = user.AppendUserAlbums(album)
  if err != nil{
    fmt.Println(err)
  }
	http.Redirect(w, r, "/", http.StatusFound)
}

const rsLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// RandString creates random n-length letters
func RandString(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = rsLetters[rand.Intn(len(rsLetters))]
    }
    return string(b)
}

func init() {
    rand.Seed(time.Now().UnixNano())
}

var ApiUpload = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  fmt.Println(r)
  authHeader := r.Header.Get("Authorization")
  bearerToken := strings.Split(authHeader, " ")

  tokenString := bearerToken[1]
  token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(os.Getenv("JWT_SIGNINKEY")), nil
  })
  claims, ok := token.Claims.(jwt.MapClaims);
  if !ok || !token.Valid {
    fmt.Println(err)
  }
  if r.Method != "POST" {
		http.Error(w, "Allowed POST method only", http.StatusMethodNotAllowed)
		return
	}
	r.ParseMultipartForm(32 << 20)
	fhs := r.MultipartForm.File["upload-firebase"]
	ctx := context.Background()
  remoteFolderName := RandString(32)
  inputFolderName := r.FormValue("album")
  var pictures []models.Picture
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
		remoteFilename := RandString(32)
    pictures = append(pictures, models.Picture{ Name:remoteFilename })
		contentType := ""
		fileData, err := ioutil.ReadAll(buf)
		if err != nil {
			contentType = "application/octet-stream"
		} else {
			contentType = http.DetectContentType(fileData)
		}
		remotePath := path.Join(remoteFolderName, remoteFilename)
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
  album := models.Album{ Name: inputFolderName, Hash: remoteFolderName, Pictures: pictures }
  err = album.Create()
  if err != nil{
    fmt.Println(err)
  }
  user := new(models.User)
  uid := uint(claims["uid"].(float64))
  err = user.FindByID(uid)
  if err != nil {
      print(err)
  }
  err = user.AppendUserAlbums(album)
  if err != nil{
    fmt.Println(err)
  }
  type ResponseData struct{
    Status int `json:"status"`
    Result string `json:"result"`
  }
  rd := ResponseData{http.StatusOK, "ok"}
  res, err := json.Marshal(rd)
  if err != nil{
    fmt.Println(err)
  }
  w.Header().Set("Content-Type", "application/json")
  w.Write(res)
})
