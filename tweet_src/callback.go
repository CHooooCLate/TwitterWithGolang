package main

import (
    "encoding/base64"
    "fmt"

    //"github.com/aws/aws-lambda-go/events"
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

type Request struct {
    OauthToken string `json:"oauth_token"`
    OauthVerifier string `json:"oauth_verifier"`
}

type Token struct {
  Id int `dynamo:"id"`
  OauthToken string `dynamo:"oauth_token"`
  SecretToken string `dynamo:"secret_token"`
  RegisterDate string `dynamo:"register_date"`
}

type  Session struct {
  Id string `dynamo:"id"`
  AccessToken string `dynamo:"access_token"`
  SecretToken string `dynamo:"secret_token"`
  RegisterDate string `dynamo:"register_date"`
}

func loadEnv() {
    err := godotenv.Load("/var/task/tweet/.env");
    //err := godotenv.Load(".env");
    if err != nil {
        log.Fatal("Error loading .env file")
    }
}

func createSessionId(screenName string) (string) {
    str := screenName + time.Now().Format("2006-01-02 15:04:05")
    return base64.URLEncoding.EncodeToString([]byte(str))
}

// Cookie渡せたらこっち
//func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
// Cookie渡せないのでこっち
func Handler(request Request) (string, error) {
    fmt.Println("Start Callback")

    loadEnv()

    //OAuthの設定
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

    table := db.Table("Token")

    // DBからOAuthトークンの取得
    var token []Token
    err := table.Get("id", 0).All(&token)
    if err != nil {
        fmt.Println("err")
        panic(err.Error())
    }

    var tempCredentials *oauth.Credentials
    tempCredentials.Token = token[0].OauthToken
    tempCredentials.Secret = token[0].SecretToken

    // Twitterから返されたOAuthトークンと、あらかじめlogin.goで入れておいたセッション上のものと一致するかをチェック
    //if tempCredentials.Token != request.PathParameters["oauth_token"] {
        //fmt.Println("invalid oauth_token")
    //}

    //アクセストークンの取得
    tokenCard, _, err := oauthClient.RequestToken(nil, tempCredentials, request.OauthVerifier)
    if err != nil {
        log.Fatal("RequestToken:", err)
    }

    spew.Dump(tokenCard)

    // 時間取得時のフォーマット指定
    format := "2006-01-02 15:04:05"

    screenName := "a"

    s := Session{
        Id: createSessionId(screenName),
        AccessToken: tokenCard.Token,
        SecretToken: tokenCard.Secret,
        RegisterDate: time.Now().Format(format),
    }

    if err := table.Put(s).Run(); err != nil {
        fmt.Println("err")
        panic(err.Error())
    }

    return tokenCard.Token, err


    // Cookie渡すときはこっち
    //return events.APIGatewayProxyResponse{
        //Body:       string(jsonBytes),
        //StatusCode: 200,
    //}, nil
}

func main() {
    lambda.Start(Handler)
}
