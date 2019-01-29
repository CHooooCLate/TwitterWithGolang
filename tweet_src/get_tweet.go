package main

import (
    "encoding/json"
    "fmt"

    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/gomodule/oauth1/oauth"
    "github.com/guregu/dynamo"
    "github.com/joho/godotenv"

    "io/ioutil"
    "log"
    "os"

    // デバッグ用
    // spew.Dump(value)
    "github.com/davecgh/go-spew/spew"
    // reflect.TypeOf(value)
    //"reflect"
)

type  Session struct {
  Id string `dynamo:"id"`
  AccessToken string `dynamo:"access_token"`
  SecretToken string `dynamo:"secret_token"`
  RegisterDate string `dynamo:"register_date"`
}

// Account アカウント
type Account struct {
	ID              string `json:"id_str"`
	ScreenName      string `json:"screen_name"`
	ProfileImageURL string `json:"profile_image_url"`
	Email           string `json:"email"`
}

func loadEnv() {
    err := godotenv.Load("/var/task/tweet/.env")
    //err := godotenv.Load(".env")
    if err != nil {
        log.Fatal("Error loading .env file")
    }
}

func Handler() (string, error) {
    loadEnv()

    // OAuthの設定
    oauthClient := &oauth.Client{
        Credentials: oauth.Credentials{
            Token:  os.Getenv("CONSUMER_KEY"),
            Secret: os.Getenv("CONSUMER_SECRET"),
        },
        TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
        ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authenticate",
        TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
    }

    // DynamoDBへ接続
    db := dynamo.New(session.New(), &aws.Config{
        Region: aws.String("us-east-2"), // "ap-northeast-1"等
    })

    sessionTable := db.Table("Session")

    // DBからOAuthトークンの取得
    var session []Session
    err := sessionTable.Scan().All(&session)
    if err != nil {
        fmt.Println("err")
        panic(err.Error())
    }

spew.Dump(session)

    tokenCard := &oauth.Credentials{
        Token: session[0].AccessToken,
        Secret: session[0].SecretToken,
    }
spew.Dump(tokenCard)
    var user Account
    response, err := oauthClient.Get(nil, tokenCard, "https://api.twitter.com/1.1/account/verify_credentials.json", nil)
    if err != nil {
		panic(err)
	}
    body, err := ioutil.ReadAll(response.Body)
    err = json.Unmarshal(body,&user)
spew.Dump(user)
    return "", nil
}

func main() {
    lambda.Start(Handler)
}
