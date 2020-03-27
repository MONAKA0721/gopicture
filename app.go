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
	"strconv"
	"strings"
	"text/template"

	"github.com/rwcarlsen/goexif/exif"

	"golang.org/x/net/context"
	"google.golang.org/api/iterator"

	firebase "firebase.google.com/go"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

var templates = template.Must(template.ParseFiles("templates/index.html", "templates/show.html"))

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"Title": "index"}
	renderTemplate(w, "index", data)
}

func ShowHandler(w http.ResponseWriter, r *http.Request) {
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
		if !strings.Contains(attrs.Name, "sequence") {
			links = append(links, "https://storage.googleapis.com/go-pictures.appspot.com/"+attrs.Name)
		}
	}
	data := map[string]interface{}{"Title": "CSFshow", "links": links}
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
			print(attrs.Name)
		}
	}
	data := map[string]interface{}{"Title": "CSFshow", "links": links}
	renderTemplate(w, "show", data)
	return
}

// UploadHandler は Cloud Storage for Firebase にアップロードするための関数です
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// if r.Method = "POST" {
	// 	http.Error(w, "Allowed POST method only", http.StatusMethodNotAllowed)
	// 	return
	// }

	defer storeClient.Close()
	ctx := context.Background()
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		defer part.Close()
		buf := bytes.NewBuffer(nil)
		buf2 := bytes.NewBuffer(nil)
		bufWriter := io.MultiWriter(buf, buf2)
		io.Copy(bufWriter, part)
		// ファイル名がない場合はスキップする
		if part.FileName() == "" {
			continue
		}

		remoteFilename := part.FileName()

		contentType := ""
		fileData, err := ioutil.ReadAll(buf)
		if err != nil {
			print("error")
			contentType = "application/octet-stream"
		} else {
			contentType = http.DetectContentType(fileData)
			print(contentType)
		}

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

		x, err := exif.Decode(buf2)
		timeStr := ""
		if err != nil {
			print("error")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			time, _ := x.DateTime()
			timeStr = time.String()
			println(time.String())
		}
		_, _, err = storeClient.Collection("links").Add(ctx, map[string]interface{}{
			"filename": remoteFilename,
			"date":     timeStr,
		})
		if err != nil {
			log.Fatalf("Failed adding alovelace: %v", err)
		}

	}
	http.Redirect(w, r, "/show", http.StatusFound)
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
func firebaseInit() (bkthdl *storage.BucketHandle, sc *firestore.Client) {
	config := &firebase.Config{
		StorageBucket: "go-pictures.appspot.com",
	}
	ctx := context.Background()
	opt := option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
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
	url := r.FormValue("url")
	_, _, err := storeClient.Collection("favorites").Add(ctx, map[string]interface{}{
		"uid":      55,
		"filepath": url,
	})
	if err != nil {
		log.Fatalf("Failed adding alovelace: %v", err)
	}
	iter := storeClient.Collection("favorites").Where("filepath", "==", url).Documents(ctx)
	count := 0
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		count++
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(strconv.Itoa(count)))
}

var bucket, storeClient = firebaseInit()

func main() {
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/upload", UploadHandler)
	http.HandleFunc("/show", ShowHandler)
	http.HandleFunc("/sequence", SequenceHandler)
	http.HandleFunc("/favorite", FavoriteHandler)
	http.ListenAndServe(":8888", nil)
}
