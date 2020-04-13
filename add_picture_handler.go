package main

import (
  "net/http"
  "fmt"
  "bytes"
  "io"
  "io/ioutil"
  "path"
  "log"
  "strings"

  "golang.org/x/net/context"
  "cloud.google.com/go/storage"

  "gopicture/models"
  "gopicture/database"
)

// AddPictureHandler handles adding new pictures to existing album
func AddPictureHandler(w http.ResponseWriter, r *http.Request) {
  if r.Method != "POST" {
		http.Error(w, "Allowed POST method only", http.StatusMethodNotAllowed)
		return
	}
  r.ParseMultipartForm(32 << 20)
	fhs := r.MultipartForm.File["upload-firebase"]
	ctx := context.Background()
  remoteFolderName := strings.Replace(r.URL.Path, "/add/", "", 1)
  fmt.Println(remoteFolderName)
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
  var album models.Album
  db := database.GetDB()
  defer db.Close()
  db.Where("hash = ?", remoteFolderName).First(&album)
  db.Model(&album).Association("Pictures").Append(pictures)

  user := new(models.User)
  ui := profileFromSession(r)
  err := user.FirstOrCreate(ui.Email, ui.Name)
  if err != nil {
      print(err)
  }
  err = user.AppendUserAlbums(album)
  if err != nil{
    fmt.Println(err)
  }
  http.Redirect(w, r, "/", http.StatusFound)
}
