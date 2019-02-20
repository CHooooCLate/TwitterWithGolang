package main

import (
    "encoding/base64"
    "fmt"

    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/gomodule/oauth1/oauth"
    "github.com/guregu/dynamo"
    "github.com/joho/godotenv"

    "log"
    "math/rand"
    "os"
    "strconv"
    "time"

    // デバッグ用
    // spew.Dump(value)
    //"github.com/davecgh/go-spew/spew"
    // reflect.TypeOf(value)
    //"reflect"
)

type Token struct {
  Id string `dynamo:"id"`
  OauthToken string `dynamo:"oauth_token"`
  SecretToken string `dynamo:"secret_token"`
  RegisterDate string `dynamo:"register_date"`
}

// APIGatewayへのレスポンスを定義するための構造体
type Response struct {
    Location string `json:"location"`
    Cookie string `json:"cookie"`
}

func loadEnv() {
    err := godotenv.Load("/var/task/tweet/.env")
    //err := godotenv.Load(".env")
    if err != nil {
        log.Fatal("Error loading .env file")
    }
}

func createSessionId() (string) {
    str := strconv.Itoa(rand.Intn(1000)) + time.Now().Format("2006-01-02 15:04:05")
    return base64.URLEncoding.EncodeToString([]byte(str))
}

func Handler() (Response, error) {
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

    // DynamoDBへ接続
    db := dynamo.New(session.New(), &aws.Config{
        Region: aws.String("us-east-2"), // "ap-northeast-1"等
    })

    table := db.Table("Token")

    // 時間取得時のフォーマット指定
    format := "2006-01-02 15:04:05"

    // session_idの作成
    id := createSessionId()

    t := Token{
        Id: id,
        OauthToken: tempCredentials.Token,
        SecretToken: tempCredentials.Secret,
        RegisterDate: time.Now().Format(format),
    }

    if err := table.Put(t).Run(); err != nil {
        fmt.Println("err")
        panic(err.Error())
    }

    authorizeUrl := oauthClient.AuthorizationURL(tempCredentials, nil)

    response := Response{
        Location: authorizeUrl,
        Cookie: fmt.Sprintf("id=%s;", id),
    }

    return response, nil
}

func main() {
    lambda.Start(Handler)
}
