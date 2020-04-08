package main

import(
  "net/http"
  "log"
  "golang.org/x/net/context"
  "google.golang.org/api/iterator"
  "strconv"
)

func FavoriteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.ParseForm()
	path := r.FormValue("path")
	_, _, err := storeClient.Collection("favorites").Add(ctx, map[string]interface{}{
		"uid":      55,
		"filepath": path,
	})
	if err != nil {
		log.Fatalf("Failed adding alovelace: %v", err)
	}
	count := CountFavorite(path)
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
