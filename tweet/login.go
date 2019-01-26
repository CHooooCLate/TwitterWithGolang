package main

import (
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/gomodule/oauth1/oauth"
    "github.com/joho/godotenv"

    "log"
    "os"

    // デバッグ用
    // spew.Dump(value)
    //"github.com/davecgh/go-spew/spew"
    // reflect.TypeOf(value)
    //"reflect"
)

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

    authorizeUrl := oauthClient.AuthorizationURL(tempCredentials, nil)

    return authorizeUrl, nil
}

func main() {
    lambda.Start(Handler)
}
