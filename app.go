package main

import (
	"log"
	"net/http"
	"os"
	"text/template"

	"gopicture/config"
	"gopicture/database"

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
	database.Init(false)
	defer database.Close()

	http.Handle("/statics/", http.StripPrefix("/statics/", http.FileServer(http.Dir("statics/"))))
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/upload", UploadHandler)
	http.HandleFunc("/show/", ShowHandler)
	http.HandleFunc("/favorite", FavoriteHandler)
	http.HandleFunc("/oauth2callback", OAuthCallbackHandler)

	port := os.Getenv("PORT")
	http.ListenAndServe(":"+port, nil)
}
