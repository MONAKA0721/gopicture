package main

import(
  "fmt"
  "bytes"
  "io"
  "path"
  "log"
  "io/ioutil"
  "net/http"
  "golang.org/x/net/context"
  "cloud.google.com/go/storage"
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
