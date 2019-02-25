package main

import (
    "fmt"

    //"github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/guregu/dynamo"

    "regexp"

    // デバッグ用
    // spew.Dump(value)
    //"github.com/davecgh/go-spew/spew"
    // reflect.TypeOf(value)
    //"reflect"
)

// APIGatewayからのリクエストを受け取るための構造体
type Request struct {
    Cookie string `json:"Cookie"`
}

// TwitterAPIから取得したアクセストークンを保存するための構造体
type  Session struct {
  Id string `dynamo:"id"`
  AccessToken string `dynamo:"access_token"`
  SecretToken string `dynamo:"secret_token"`
  RegisterDate string `dynamo:"register_date"`
}

// APIGatewayへのレスポンスを定義するための構造体
type Response struct {
    Location string `json:"location"`
    Cookie string `json:"cookie"`
}

type Cookie struct {
    Id string `json:id`
}

func Handler(request Request) (Response, error) {
    // DynamoDBへ接続
    db := dynamo.New(session.New(), &aws.Config{
        Region: aws.String("us-east-2"), // "ap-northeast-1"等
    })

    // Tokenテーブル
    sessionTable := db.Table("Session")

    // session_idの取り出し
    assigned := regexp.MustCompile("id=(.*)")
	group := assigned.FindSubmatch([]byte(request.Cookie))
    sessionId := string(group[1])

    // DBからOAuthトークンを削除
    err := sessionTable.Delete("id", sessionId).Run()
    if err != nil {
        fmt.Println("err")
        panic(err.Error())
    }

    // 時間取得時のフォーマット指定
    expired := "Sat,01-Jan-2000 00:00:00"

    // リダイレクトさせてCookieをつけたい
    redirectUrl := "https://mb8mab272h.execute-api.us-east-2.amazonaws.com/twimal/my-page"

    returnResponse := Response{
        Location: redirectUrl,
        Cookie: fmt.Sprintf("id=%s;%s;", sessionId, expired),
    }

    return returnResponse, err
}

func main() {
    lambda.Start(Handler)
}
