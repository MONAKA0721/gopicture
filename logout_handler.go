package main

import(
  "net/http"
  "fmt"
)

// LogoutHandler clears the default session.
func LogoutHandler(w http.ResponseWriter, r *http.Request){
	session, err := SessionStore.New(r, defaultSessionID)
	if err != nil {
		fmt.Println(err)
	}
	session.Options.MaxAge = -1 // Clear session.
	if err := session.Save(r, w); err != nil {
    fmt.Println(err)
	}
  forwardSession, err := SessionStore.New(r, forwardSessionID)
	if err != nil {
		fmt.Println(err)
	}
	forwardSession.Options.MaxAge = -1 // Clear session.
	if err := forwardSession.Save(r, w); err != nil {
    fmt.Println(err)
	}
	redirectURL := r.FormValue("redirect")
	if redirectURL == "" {
		redirectURL = "/"
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
