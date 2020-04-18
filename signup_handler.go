package main

import (
	"encoding/json"
	"fmt"
	"gopicture/models"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	var error models.Error

	// r.body に何が帰ってくるか確認
	fmt.Println(r.Body)

	// https://golang.org/pkg/encoding/json/#NewDecoder
	json.NewDecoder(r.Body).Decode(&user)

	if user.Email == "" {
		error.Message = "Email は必須です。"
		errorInResponse(w, http.StatusBadRequest, error)
		return
	}

	if user.Password == "" {
		error.Message = "パスワードは必須です。"
		errorInResponse(w, http.StatusBadRequest, error)
		return
	}

	// user に何が格納されているのか
	fmt.Println(user)

	// dump も出せる
	fmt.Println("---------------------")

	// パスワードのハッシュを生成
	// https://godoc.org/golang.org/x/crypto/bcrypt#GenerateFromPassword
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("パスワード: ", user.Password)
	fmt.Println("ハッシュ化されたパスワード", hash)

	user.Password = string(hash)
	fmt.Println("コンバート後のパスワード: ", user.Password)

	err = user.Create()
	if err != nil {
		error.Message = "サーバーエラー"
		errorInResponse(w, http.StatusInternalServerError, error)
		return
	}

	// DB に登録できたらパスワードをからにしておく
	user.Password = ""
	w.Header().Set("Content-Type", "application/json")

	// JSON 形式で結果を返却
	responseByJSON(w, user)
}

// レスポンスにエラーを突っ込んで、返却するメソッド
func errorInResponse(w http.ResponseWriter, status int, error models.Error) {
	w.WriteHeader(status) // 400 とか 500 などの HTTP status コードが入る
	json.NewEncoder(w).Encode(error)
	return
}

// JSON 形式で結果を返却
// data interface{} とすると、どのような変数の型でも引数として受け取ることができる
func responseByJSON(w http.ResponseWriter, data interface{}) {
	json.NewEncoder(w).Encode(data)
	return
}
