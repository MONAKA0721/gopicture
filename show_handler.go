package main

import (
	"fmt"
	"gopicture/database"
	"net/http"
	"github.com/gorilla/sessions"
)

type File struct {
	Link     string
	FileName string
	FavCount int
}

func ShowHandler(w http.ResponseWriter, r *http.Request) {
	profile := profileFromSession(r)
	if profile == nil {
		forwardSession, err := SessionStore.New(r, forwardSessionID)
		forwardSession.Options = &sessions.Options{
			Path: "/",
    }
		if err != nil {
			fmt.Println(err)
		}
		redirectURL := r.URL.String()
		forwardSession.Values[forwardSessionKey] = redirectURL
		if err := forwardSession.Save(r, w); err != nil {
	    fmt.Println(err)
		}
		http.Redirect(w, r, "/", http.StatusFound)
	}
	albumHash := r.URL.Path[len("/show/"):]
	db := database.GetDB()
	defer db.Close()
	rows, err := db.Raw(`SELECT pictures.name, pictures.id FROM pictures
		INNER JOIN albums ON albums.id = pictures.album_id
		WHERE albums.hash = ?`, albumHash).Rows()
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()
	list := []File{}
	var pname string
	var pid int
	for rows.Next() {
		rows.Scan(&pname, &pid)
		var count int
		row := db.Raw(`SELECT count(*)
	  FROM user_fav_pictures
	  WHERE picture_id = ?`, pid).Row()
		row.Scan(&count)
		file := File{"https://storage.googleapis.com/go-pictures.appspot.com/" +
			albumHash + "/" + pname, pname, count}
		list = append(list, file)
	}
	data := struct {
		List []File
	}{
		List: list,
	}
	renderTemplate(w, "show", data)
	return
}
