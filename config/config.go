package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/net/context"
	"google.golang.org/api/option"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"

	"firebase.google.com/go"
)

func GetDBConfig() (string, string) {
	DBMS := "mysql"
	if os.Getenv("GO_ENV") == "dev" {
		USER := os.Getenv("MYSQL_USER")
		PASS := os.Getenv("MYSQL_PASSWORD")
		PROTOCOL := "tcp(mysql:3306)"
		DBNAME := os.Getenv("MYSQL_DATABASE")
		CONNECT := USER + ":" + PASS + "@" + PROTOCOL + "/" + DBNAME + "?parseTime=true"
		return DBMS, CONNECT
	}
	CONNECT := strings.Replace(os.Getenv("CLEARDB_DATABASE_URL"), "mysql://", "", 1) + "&parseTime=true"
	return DBMS, CONNECT
}

func FirebaseInit(env string) (bkthdl *storage.BucketHandle, sc *firestore.Client) {
	if env == "dev" {
		err := godotenv.Load("envfiles/dev.env")
		if err != nil {
			log.Fatalln(err)
		}
	}
	config := &firebase.Config{
		StorageBucket: "go-pictures.appspot.com",
	}
	ctx := context.Background()
	jsonStr := fmt.Sprintf(`{
	  "type": "service_account",
	  "project_id": "go-pictures",
	  "private_key_id": "%s",
	  "private_key": "%s",
	  "client_email": "firebase-adminsdk-of90d@go-pictures.iam.gserviceaccount.com",
	  "client_id": "%s",
	  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
	  "token_uri": "https://oauth2.googleapis.com/token",
	  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
	  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/firebase-adminsdk-of90d%%40go-pictures.iam.gserviceaccount.com"
	}`, os.Getenv("PRIVATE_KEY_ID"), os.Getenv("PRIVATE_KEY"), os.Getenv("CLIENT_ID"))
	opt := option.WithCredentialsJSON([]byte(jsonStr))

	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Storage(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	bucket, err := client.DefaultBucket()
	if err != nil {
		log.Fatalln(err)
	}

	storeClient, err := app.Firestore(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	return bucket, storeClient
}
