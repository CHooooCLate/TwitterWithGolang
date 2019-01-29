package main

import (
    "fmt"
    
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/ChimeraCoder/anaconda"
    "github.com/joho/godotenv"
    
    "io/ioutil"
    "log"
    "net/url"
    "os"
    
    // デバッグ用
    //"reflect"
    //"github.com/davecgh/go-spew/spew"
)

type Request struct {
    Animal string `json:"animal"`
}


func loadEnv() {
    err := godotenv.Load("/var/task/tweet/.env");
    if err != nil {
        log.Fatal("Error loading .env file")
    }
}

func getTwitterApi() *anaconda.TwitterApi {
    anaconda.SetConsumerKey(os.Getenv("CONSUMER_KEY"))
    anaconda.SetConsumerSecret(os.Getenv("CONSUMER_SECRET"))
    return anaconda.NewTwitterApi(os.Getenv("ACCESS_TOKEN"), os.Getenv("ACCESS_TOKEN_SECRET"))
}

func Handler(request Request) (string, error) {
    fmt.Println("Start twimal")

    // ファイルをOpenする
    file, err := os.Open("/var/task/tweet/index.html")
    if err != nil{
        fmt.Println("error")
    }
    defer file.Close()

    // HTMLファイルを読み込み
    inputHtml, err := ioutil.ReadAll(file)

    // .envの読み込み
    loadEnv();

    // Tweetを取得
    api := getTwitterApi()

    v := url.Values{}
    v.Set("count", "20")

    var searchResult anaconda.SearchResponse

    switch request.Animal {
    case "test":
        searchResult, _ = api.GetSearch("月曜の夜！まだやすみ！！ OR ジュゲム", v)
    case "dog":
        searchResult, _ = api.GetSearch("犬 OR dog filter:videos", v)
    case "cat":
        searchResult, _ = api.GetSearch("猫 OR cat filter:videos", v)
    case "fish":
        searchResult, _ = api.GetSearch("魚 OR fish filter:videos", v)
    default:
        searchResult, _ = api.GetSearch("犬 OR 猫　OR 動物 filter:videos", v)
    }

    // HTMLにTweetを埋め込み
    list := "<ul>"
    for _, tweet := range searchResult.Statuses {
        base := fmt.Sprintf("<blockquote class=\"twitter-tweet\"><a href=\"https://twitter.com/%s/status/%s\"></a></blockquote>", tweet.User.ScreenName, tweet.IdStr)
        list += "<li>" + base + "</li>"
    }
    list += "</ul>"

    outputHtml := fmt.Sprintf(string(inputHtml), list)

    // 出力確認
    //fmt.Println(string(outputHtml));

    return string(outputHtml), nil
}

func main() {
    lambda.Start(Handler)
}
