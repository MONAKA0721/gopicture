package main

import (
  "fmt"
  "net/http"
  "os"
  "io"
  "crypto/rand"
  "strings"
  "encoding/base32"
  "net/url"
  "errors"

  "github.com/gorilla/sessions"
  "github.com/gofrs/uuid"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/google"
)
// loginHandler initiates an OAuth flow to authenticate the user.
func loginHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := uuid.Must(uuid.NewV4()).String()

	oauthFlowSession, err := SessionStore.New(r, sessionID)
	if err != nil {
		fmt.Println(err)
	}
	oauthFlowSession.Options.MaxAge = 10 * 60 // 10 minutes

	redirectURL, err := validateRedirectURL(r.FormValue("redirect"))
	if err != nil {
		fmt.Println(err)
	}
	oauthFlowSession.Values[oauthFlowRedirectKey] = redirectURL

	if err := oauthFlowSession.Save(r, w); err != nil {
		fmt.Println(err)
	}

	// Use the session ID for the "state" parameter.
	// This protects against CSRF (cross-site request forgery).
	// See https://godoc.org/golang.org/x/oauth2#Config.AuthCodeURL for more detail.
	url := OAuthConfig.AuthCodeURL(sessionID, oauth2.ApprovalForce,
		oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusFound)
}

func configureOAuthClient(clientID, clientSecret string) *oauth2.Config {
	redirectURL := os.Getenv("OAUTH2_CALLBACK")
	if redirectURL == "" {
		redirectURL = "http://localhost:8888/oauth2callback"
	}
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func init() {
	OAuthConfig = configureOAuthClient(os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"))
	// Configure storage method for session-wide information.
	// Update "something-very-secret" with a hard to guess string or byte sequence.
	// 乱数生成
	b := make([]byte, 48)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		panic(err)
	}
	str := strings.TrimRight(base32.StdEncoding.EncodeToString(b), "=")
	cookieStore := sessions.NewCookieStore([]byte(str))
	cookieStore.Options = &sessions.Options{
		HttpOnly: true,
	}
	SessionStore = cookieStore
}

// validateRedirectURL checks that the URL provided is valid.
// If the URL is missing, redirect the user to the application's root.
// The URL must not be absolute (i.e., the URL must refer to a path within this
// application).
func validateRedirectURL(path string) (string, error) {
	if path == "" {
		return "/", nil
	}

	// Ensure redirect URL is valid and not pointing to a different server.
	parsedURL, err := url.Parse(path)
	if err != nil {
		return "/", err
	}
	if parsedURL.IsAbs() {
		return "/", errors.New("URL must not be absolute")
	}
	return path, nil
}
