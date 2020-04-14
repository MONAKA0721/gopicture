package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"gopicture/config"
	"gopicture/database"
	"gopicture/models"

	_ "github.com/go-sql-driver/mysql"
)

var templates = template.Must(template.ParseFiles("templates/index.html", "templates/show.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	if err := templates.ExecuteTemplate(w, tmpl+".html", data); err != nil {
		log.Fatalln("Unable to execute template.")
	}
}

var bucket, storeClient = config.FirebaseInit(os.Getenv("GO_ENV"))

func main() {
	database.Init(false, models.User{}, models.Album{}, models.Picture{})
	defer database.Close()

	http.Handle("/statics/", http.StripPrefix("/statics/", http.FileServer(http.Dir("statics/"))))
	http.HandleFunc("/", Index)

	port := os.Getenv("PORT")
	http.ListenAndServe(":"+port, nil)
}

func Index(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	if strings.HasPrefix(p, "/login") {
		LoginHandler(w, r)
	} else if strings.HasPrefix(p, "/logout") {
		LogoutHandler(w, r)
	} else if strings.HasPrefix(p, "/upload") {
		UploadHandler(w, r)
	} else if strings.HasPrefix(p, "/show/") {
		ShowHandler(w, r)
	} else if strings.HasPrefix(p, "/add/") {
		AddPictureHandler(w, r)
	} else if strings.HasPrefix(p, "/favorite") {
		FavoriteHandler(w, r)
	} else if strings.HasPrefix(p, "/oauth2callback") {
		OAuthCallbackHandler(w, r)
	} else if strings.HasPrefix(p, "/favicon.ico") {
		FaviconHandler(w, r)
	} else {
		IndexHandler(w, r)
	}
}
