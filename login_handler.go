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
  "encoding/json"

  "golang.org/x/crypto/bcrypt"
  "github.com/gorilla/sessions"
  "github.com/gofrs/uuid"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/google"
  "github.com/dgrijalva/jwt-go"
  "github.com/jinzhu/gorm"
  jwtmiddleware "github.com/auth0/go-jwt-middleware"

  "gopicture/models"
)
// LoginHandler initiates an OAuth flow to authenticate the user.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
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
	OAuthConfig = configureOAuthClient(os.Getenv("OAUTH_CLIENT_ID"), os.Getenv("OAUTH_CLIENT_SECRET"))
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

func ApiLoginHandler(w http.ResponseWriter, r *http.Request ) {
  var user models.User
  var error models.Error
  var jwt models.JWT

  json.NewDecoder(r.Body).Decode(&user)

  if user.Email == "" {
      error.Message = "Email は必須です。"
      errorInResponse(w, http.StatusBadRequest, error)
      return
  }

  if user.Password == "" {
      error.Message = "パスワードは、必須です。"
      errorInResponse(w, http.StatusBadRequest, error)
  }

  // 追加(この位置であること)
  password := user.Password
  fmt.Println("password: ", password)

  // 認証キー(Emal)のユーザー情報をDBから取得
  err := user.FindByEmail(user.Email)

  if err != nil {
    if gorm.IsRecordNotFoundError(err){
      error.Message = "ユーザが存在しません。"
      errorInResponse(w, http.StatusBadRequest, error)
    }
    fmt.Println(err)
  }

  // 追加(この位置であること)
  hasedPassword := user.Password
  fmt.Println("hasedPassword: ", hasedPassword)

  err = bcrypt.CompareHashAndPassword([]byte(hasedPassword), []byte(password))

  if err != nil {
      error.Message = "無効なパスワードです。"
      errorInResponse(w, http.StatusUnauthorized, error)
      return
  }

  token, err := createToken(user)

  if err != nil {
      fmt.Println(err)
  }

  w.WriteHeader(http.StatusOK)
  jwt.Token = token

  responseByJSON(w, jwt)
}

func createToken(user models.User) (string, error) {
  var err error

  // Token を作成
  // jwt -> JSON Web Token - JSON をセキュアにやり取りするための仕様
  // jwtの構造 -> {Base64 encoded Header}.{Base64 encoded Payload}.{Signature}
  // HS254 -> 証明生成用(https://ja.wikipedia.org/wiki/JSON_Web_Token)
  token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
      "email": user.Email,
      "uid": user.ID,
      "iss":   "__init__", // JWT の発行者が入る(文字列(__init__)は任意)
  })
  tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SIGNINKEY")))

  fmt.Println("-----------------------------")
  fmt.Println("tokenString:", tokenString)

  if err != nil {
      fmt.Println(err)
  }

  return tokenString, nil
}

// JwtMiddleware check token
var JwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
  ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
      return []byte(os.Getenv("JWT_SIGNINKEY")), nil
  },
  SigningMethod: jwt.SigningMethodHS256,
})
