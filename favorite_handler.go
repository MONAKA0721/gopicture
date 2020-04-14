package main

import(
  "net/http"
  "log"
  "golang.org/x/net/context"
  "google.golang.org/api/iterator"
  "strconv"
  "fmt"
  "gopicture/database"
  "gopicture/models"
)

// FavoriteHandler handle user's favorite for pictures
func FavoriteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.ParseForm()
  albumHash := r.FormValue("albumHash")
  fileName := r.FormValue("fileName")
  db := database.GetDB()
  defer db.Close()
  row := db.Raw(`SELECT pictures.id
    FROM pictures INNER JOIN albums
    ON pictures.album_id = albums.id
    WHERE albums.hash = ? AND pictures.name = ?`, albumHash, fileName).Row()
  var pid int
  row.Scan(&pid)
  user := new(models.User)
  ui, _ := profileFromSession(r)
  err := user.FirstOrCreate(ui.Email, ui.Name)
  if err != nil {
    fmt.Println(err)
  }
  user.AppendFavPictures(pid)
  var count int
  row = db.Raw(`SELECT count(*)
  FROM user_fav_pictures
  WHERE picture_id = ?`, pid).Row()
  row.Scan(&count)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(strconv.Itoa(count)))
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
