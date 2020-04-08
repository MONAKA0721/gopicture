package main

import (
	"net/http"
	"regexp"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"cloud.google.com/go/storage"
)

type File struct {
	Link     string
	Path     string
	FavCount int
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
