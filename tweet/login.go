package main

import (
    "fmt"

    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/dynamodb"
    "github.com/gomodule/oauth1/oauth"
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

    // session
    sess, err := session.NewSession()
    if err != nil {
        panic(err)
    }

    svc := dynamodb.New(sess)

    // 時間取得時のフォーマット指定
    format := "2006-01-02 15:04:05"

    // PutItem
    putParams := &dynamodb.PutItemInput{
        TableName: aws.String("Token"),
        Item: map[string]*dynamodb.AttributeValue{
            "id": {
                N: aws.String("0"),
            },
            "oauth_token": {
                S: aws.String(tempCredentials.Token),
            },
            "secret_token": {
                S: aws.String(tempCredentials.Secret),
            },
            "register_date": {
                S: aws.String(time.Now().Format(format)),
            },
        },
    }

    putItem, putErr := svc.PutItem(putParams)
    if putErr != nil {
        panic(putErr)
    }
    fmt.Println(putItem)

    authorizeUrl := oauthClient.AuthorizationURL(tempCredentials, nil)

    return authorizeUrl, nil
}

func main() {
    lambda.Start(Handler)
}
