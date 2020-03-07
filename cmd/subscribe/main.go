package main

import (
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/outcrawl/backend/newsletter"
)

func main() {
	lambda.Start(newsletter.HandleSubscribe)
}
