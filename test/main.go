package main

import (
  "github.com/aws/aws-lambda-go/lambda"
)

func Handler() (string, error) {
    return "<form method=\"post\" action=\"/register\"><input type=\"text\" name=\"email\" value=\"\"/><input type=\"password\" name=\"password\" value=\"\"/><input type=\"submit\" value=\"Send\"></form>", nil
}

func main() {
    lambda.Start(Handler)
}
