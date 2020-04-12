package main

import(
  "strings"
  "net/http"
  "encoding/gob"
  "fmt"
  "golang.org/x/net/context"
  "google.golang.org/api/iterator"
  "golang.org/x/oauth2"
  "github.com/gorilla/sessions"
  oauthapi "google.golang.org/api/oauth2/v2"
  "gopicture/database"
)

var (
	OAuthConfig  *oauth2.Config
	SessionStore sessions.Store
)

func init() {
	// Gob encoding for gorilla/sessions
	gob.Register(&oauth2.Token{})
	gob.Register(&oauthapi.Userinfoplus{})
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
  // var pURL *string = &redirectURL
  forwardSession, err := SessionStore.Get(r, forwardSessionID)
  if err != nil {
		fmt.Println(err)
  }else{
    print("test")
    // if !ok {
    //   fmt.Println("session error")
    // }
  }
  _, ok := forwardSession.Values[forwardSessionKey]
  if !ok {
		print("error")
	}
	d := struct {
		AuthEnabled bool
		UserInfo     *oauthapi.Userinfoplus
		LoginURL    string
		LogoutURL   string
	}{
		AuthEnabled: OAuthConfig != nil,
		LoginURL:    "/login?redirect=" + "",
		LogoutURL:   "/logout?redirect=" + r.URL.RequestURI(),
	}
	if d.AuthEnabled {
		// Ignore any errors.
		d.UserInfo = profileFromSession(r)
	}
	ctx := context.Background()
	it := bucket.Objects(ctx, nil)
	folders := []string{}
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return
		}
		end := strings.Index(attrs.Name, "/")
		folders = append(folders, attrs.Name[:end])
	}
	m := make(map[string]bool)
	uniqFolders := []string{}

	for _, ele := range folders {
		if !m[ele] {
			m[ele] = true
			uniqFolders = append(uniqFolders, ele)
		}
	}
  type Folder struct {
    Name string
    Hash string
    TopPicName string
  }
  db := database.GetDB()
  rows, err := db.Raw(`SELECT albums.name, albums.hash, albums.id
    FROM albums INNER JOIN user_albums ON albums.id = user_albums.album_id
    WHERE user_albums.user_id = ?`, 1).Rows()
  if err != nil{
    fmt.Println(err)
  }
  defer rows.Close()
  var indexFolders []Folder
  for rows.Next() {
    var name string
    var hash string
    var aid int
    rows.Scan(&name, &hash, &aid)
    row := db.Raw(`SELECT temp.pname FROM
      (SELECT p.name pname, count(*) cnt
      FROM (albums a INNER JOIN pictures p on a.id = p.album_id)
      INNER JOIN user_fav_pictures f
      ON p.id = f.picture_id where a.id = ? GROUP BY p.name) temp
      WHERE temp.cnt = (SELECT max(cnt2)
      FROM(SELECT p.name pname, count(*) cnt2
      FROM (albums a INNER JOIN pictures p on a.id = p.album_id)
      INNER JOIN user_fav_pictures f ON p.id = f.picture_id where a.id = ?
      GROUP BY p.name) num)`, aid, aid).Row()
    var pictureName string
    row.Scan(&pictureName)
    indexFolders = append(indexFolders, Folder{Name:name, Hash:hash, TopPicName:pictureName})
  }
	data := map[string]interface{}{
    "Title": "index",
    "folders": indexFolders,
    "userinfo": d.UserInfo,
    "LoginURL": d.LoginURL,
    "LogoutURL": d.LogoutURL}
	renderTemplate(w, "index", data)
}

// profileFromSession retreives the Google+ profile from the default session.
// Returns nil if the profile cannot be retreived (e.g. user is logged out).
func profileFromSession(r *http.Request) *oauthapi.Userinfoplus {
	session, err := SessionStore.Get(r, defaultSessionID)
	if err != nil {
		return nil
	}
	tok, ok := session.Values[oauthTokenSessionKey].(*oauth2.Token)
	if !ok || !tok.Valid() {
		return nil
	}
	ui, ok := session.Values[googleProfileSessionKey].(*oauthapi.Userinfoplus)
	if !ok {
		return nil
	}
	return ui
}
