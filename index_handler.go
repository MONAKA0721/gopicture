package main

import(
  "strings"
  "net/http"
  "fmt"
  "encoding/gob"
  "golang.org/x/net/context"
  "google.golang.org/api/iterator"
  "golang.org/x/oauth2"
  "github.com/gorilla/sessions"
  oauthapi "google.golang.org/api/oauth2/v2"
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
	d := struct {
		AuthEnabled bool
		UserInfo     *oauthapi.Userinfoplus
		LoginURL    string
		LogoutURL   string
	}{
		AuthEnabled: OAuthConfig != nil,
		LoginURL:    "/login?redirect=" + r.URL.RequestURI(),
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

	data := map[string]interface{}{"Title": "index", "folders": uniqFolders, "userinfo": d.UserInfo, "LogoutURL": d.LogoutURL}
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
