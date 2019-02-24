package main

import (
    "encoding/base64"
    "encoding/json"
    "fmt"

    //"github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/gomodule/oauth1/oauth"
    "github.com/guregu/dynamo"
    "github.com/joho/godotenv"

    "io/ioutil"
    "log"
    "os"
    "regexp"
    "time"

    // デバッグ用
    // spew.Dump(value)
    //"github.com/davecgh/go-spew/spew"
    // reflect.TypeOf(value)
    //"reflect"
)

// APIGatewayからのリクエストを受け取るための構造体
type Request struct {
    OauthToken string `json:"oauth_token"`
    OauthVerifier string `json:"oauth_verifier"`
    Cookie string `json:"Cookie"`
}

// TwitterAPIから取得した一時Tokenを保存するための構造体
type Token struct {
  Id string `dynamo:"id"`
  OauthToken string `dynamo:"oauth_token"`
  SecretToken string `dynamo:"secret_token"`
  RegisterDate string `dynamo:"register_date"`
}

// TwitterAPIから取得したアクセストークンを保存するための構造体
type  Session struct {
  Id string `dynamo:"id"`
  AccessToken string `dynamo:"access_token"`
  SecretToken string `dynamo:"secret_token"`
  RegisterDate string `dynamo:"register_date"`
}

// TwitterAPIから取得したユーザー情報から、screen_nameを取り出すための構造体
type Account struct {
	ScreenName string `json:"screen_name"`
}

// APIGatewayへのレスポンスを定義するための構造体
type Response struct {
    Location string `json:"location"`
    Cookie string `json:"cookie"`
}

type Cookie struct {
    Id string `json:id`
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

func Handler(request Request) (Response, error) {
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

    // Tokenテーブル
    tokenTable := db.Table("Token")

    // session_idの取り出し
    assigned := regexp.MustCompile("id=(.*)")
	group := assigned.FindSubmatch([]byte(request.Cookie))
    sessionId := string(group[1])

    // DBからOAuthトークンの取得
    var token []Token
    err := tokenTable.Get("id", sessionId).All(&token)
    if err != nil {
        fmt.Println("err")
        panic(err.Error())
    }

    // OAuthトークンを決められた構造体にする
    tempCredentials := &oauth.Credentials{
        Token: token[0].OauthToken,
        Secret: token[0].SecretToken,
    }

    // Twitterから返されたOAuthトークンと、あらかじめlogin.goで入れておいたセッション上のものと一致するかをチェック
    if tempCredentials.Token != request.OauthToken {
        panic(tempCredentials.Token + "_" + request.OauthToken)
    }

    //アクセストークンの取得
    tokenCard, _, err := oauthClient.RequestToken(nil, tempCredentials, request.OauthVerifier)
    if err != nil {
        log.Fatal("RequestToken:", err)
        panic(err.Error())
    }

    // 時間取得時のフォーマット指定
    format := "2006-01-02 15:04:05"

    // TwitterAPIからアクセストークンの取得
    response, err := oauthClient.Get(nil, tokenCard, "https://api.twitter.com/1.1/account/verify_credentials.json", nil)
    if err != nil {
		panic(err)
	}
    body, err := ioutil.ReadAll(response.Body)

    // 取得したユーザー情報をJSONから変換する
    var user Account
    err = json.Unmarshal(body,&user)

    // Sessionテーブル
    sessionTable := db.Table("Session")

    // session_idの作成
    id := createSessionId(user.ScreenName)

    // Sessionテーブルに格納するレコードの作成
    s := Session{
        Id: id,
        AccessToken: tokenCard.Token,
        SecretToken: tokenCard.Secret,
        RegisterDate: time.Now().Format(format),
    }
    // INSERTの実行
    if err = sessionTable.Put(s).Run(); err != nil {
        fmt.Println("err")
        panic(err.Error())
    }

    // リダイレクトさせてCookieをつけたい
    redirectUrl := "https://mb8mab272h.execute-api.us-east-2.amazonaws.com/twimal/my-page"

    returnResponse := Response{
        Location: redirectUrl,
        Cookie: fmt.Sprintf("id=%s;", id),
    }

    return returnResponse, err
}

func main() {
    lambda.Start(Handler)
}
