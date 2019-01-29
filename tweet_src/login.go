package main

import (
    "fmt"

    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/gomodule/oauth1/oauth"
    "github.com/guregu/dynamo"
    "github.com/joho/godotenv"

    "log"
    "os"
    "time"

    // デバッグ用
    // spew.Dump(value)
    "github.com/davecgh/go-spew/spew"
    // reflect.TypeOf(value)
    //"reflect"
)

type Token struct {
  Id int `dynamo:"id"`
  OauthToken string `dynamo:"oauth_token"`
  SecretToken string `dynamo:"secret_token"`
  RegisterDate string `dynamo:"register_date"`
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

    callbackUrl := os.Getenv("CALLBACK_URL")

    tempCredentials, err := oauthClient.RequestTemporaryCredentials(nil, callbackUrl, nil)
    if err != nil {
        log.Fatal("RequestTemporaryCredentials:", err)
    }

    spew.Dump(tempCredentials)

    // DynamoDBへ接続
    db := dynamo.New(session.New(), &aws.Config{
        Region: aws.String("us-east-2"), // "ap-northeast-1"等
    })

    table := db.Table("Token")

    // 時間取得時のフォーマット指定
    format := "2006-01-02 15:04:05"

    t := Token{
        Id: 0,
        OauthToken: tempCredentials.Token,
        SecretToken: tempCredentials.Secret,
        RegisterDate: time.Now().Format(format),
    }

    if err := table.Put(t).Run(); err != nil {
        fmt.Println("err")
        panic(err.Error())
    }

    // DBからOAuthトークンの取得
    var token []Token
    err = table.Get("id", 0).All(&token)
    if err != nil {
        fmt.Println("err")
        panic(err.Error())
    }
    spew.Dump(token[0].OauthToken)

    authorizeUrl := oauthClient.AuthorizationURL(tempCredentials, nil)

    return authorizeUrl, nil
}

func main() {
    lambda.Start(Handler)
}
