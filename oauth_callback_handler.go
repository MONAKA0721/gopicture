package main

import(
  "net/http"
  "fmt"
  "golang.org/x/net/context"
  oauthapi "google.golang.org/api/oauth2/v2"
  "gopicture/models"
)

const (
	defaultSessionID        = "default"
	oauthFlowRedirectKey    = "redirect"
	oauthTokenSessionKey    = "oauth_token"
	googleProfileSessionKey = "google_profile"
  forwardSessionID = "forward"
  forwardSessionKey = "normal"
  userIDSessionKey = "user_id"
)

func OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	oauthFlowSession, err := SessionStore.Get(r, r.FormValue("state"))
	if err != nil {
		fmt.Println(err)
	}

	redirectURL, ok := oauthFlowSession.Values[oauthFlowRedirectKey].(string)
	// Validate this callback request came from the app.
	if !ok {
		fmt.Println(err)
	}

	session, err := SessionStore.New(r, defaultSessionID)
	if err != nil {
		fmt.Println(err)
	}

	//パラメータからアクセスコードを読み取り
	code := r.URL.Query()["code"]
	if code == nil || len(code) == 0 {
		fmt.Fprint(w, "Invalid Parameter")
	}
	//いろいろライブラリが頑張って
	ctx := context.Background()
	tok, err := OAuthConfig.Exchange(ctx, code[0])
	if err != nil {
		fmt.Fprintf(w, "OAuth Error:%v", err)
	}
	//APIクライアントができて
	client := OAuthConfig.Client(ctx, tok)
	//Userinfo APIをGetしてDoして
	svr, _ := oauthapi.New(client)
	ui, err := svr.Userinfo.Get().Do()
	if err != nil {
		fmt.Fprintf(w, "OAuth Error:%v", err)
	}
	session.Values[oauthTokenSessionKey] = tok
	// Strip the profile to only the fields we need. Otherwise the struct is too big.
	session.Values[googleProfileSessionKey] = ui
  user := new(models.User)
  err = user.FirstOrCreate(ui.Email, ui.Name)
  if err != nil {
      print(err)
  }
  session.Values[userIDSessionKey] = user.ID
	if err := session.Save(r, w); err != nil {
		fmt.Println(err)
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
