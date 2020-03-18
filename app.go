package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"

	"golang.org/x/net/context"
	"google.golang.org/api/iterator"

	firebase "firebase.google.com/go"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

var templates = template.Must(template.ParseFiles("templates/index.html", "templates/show.html", "templates/CSFshow.html", "templates/upload.html"))

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"Title": "index"}
	renderTemplate(w, "index", data)
}

func CSFShowHandler(w http.ResponseWriter, r *http.Request) {
	bucket := firebaseInit()
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
		// fmt.Fprintln(w, attrs.Name)
		links = append(links, "https://storage.googleapis.com/go-pictures.appspot.com/"+attrs.Name)
	}
	data := map[string]interface{}{"Title": "CSFshow", "links": links}
	renderTemplate(w, "CSFshow", data)
	o := bucket.Object("スクリーンショット 2020-03-08 14.03.38.png")
	attrs, _ := o.Attrs(ctx)
	print(attrs.Created.String())
	return
}

// CSFUploadHandler は Cloud Storage for Firebase にアップロードするための関数です
func CSFUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := map[string]interface{}{"Title": "Upload"}
		renderTemplate(w, "upload", data)
	}
	// if r.Method = "POST" {
	// 	http.Error(w, "Allowed POST method only", http.StatusMethodNotAllowed)
	// 	return
	// }
	bucket := firebaseInit()
	ctx := context.Background()
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for {
		file, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		defer file.Close()
		// x, err := exif.Decode(file)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// } else {
		// 	time, _ := x.DateTime()
		// 	println(time.String())
		// }

		//ファイル名がない場合はスキップする
		if file.FileName() == "" {
			continue
		}
		// file, fileHeader, err := r.FormFile("uploadCSF")
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		contentType := ""
		fileData, err := ioutil.ReadAll(file)
		if err != nil {
			contentType = "application/octet-stream"
		} else {
			contentType = http.DetectContentType(fileData)
		}

		remoteFilename := file.FileName()
		writer := bucket.Object(remoteFilename).NewWriter(ctx)
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
		http.Redirect(w, r, "/uploadCSF", http.StatusFound)

	}
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Allowed POST method only", http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseMultipartForm(32 << 20) // maxMemory
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file, _, err := r.FormFile("upload")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	f, err := os.Create("/tmp/test.jpg")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	io.Copy(f, file)
	http.Redirect(w, r, "/show", http.StatusFound)
}

func ShowHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("/tmp/test.jpg")
	defer file.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	img, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeImageWithTemplate(w, "show", &img)
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
	//	w.Header().Set("Content-Type", "image/jpeg")
	//	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	//	if _, err := w.Write(buffer.Bytes()); err != nil {
	//		log.Println("unable to write image.")
	//	}
	str := base64.StdEncoding.EncodeToString(buffer.Bytes())
	data := map[string]interface{}{"Title": tmpl, "Image": str}
	renderTemplate(w, tmpl, data)
}
func firebaseInit() (bkthdl *storage.BucketHandle) {
	config := &firebase.Config{
		StorageBucket: "go-pictures.appspot.com",
	}
	opt := option.WithCredentialsFile("storage-exp-key.json")
	app, err := firebase.NewApp(context.Background(), config, opt)
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
	return bucket
}
func main() {
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/upload", UploadHandler)
	http.HandleFunc("/show", ShowHandler)
	http.HandleFunc("/uploadCSF", CSFUploadHandler)
	http.HandleFunc("/showCSF", CSFShowHandler)
	http.ListenAndServe(":8888", nil)
}
