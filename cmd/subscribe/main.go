package main

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/outcrawl/backend/newsletter"
)

func subscribe(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sr := newsletter.SubscribeRequest{
		Email:     req.QueryStringParameters["email"],
		Recaptcha: req.QueryStringParameters["recaptcha"],
	}
	if err := newsletter.HandleSubscribe(sr); err != nil {
		return errorResponse(http.StatusBadRequest, err.Error()), nil
	}
	return successResponse(), nil
}

func successResponse() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type":                     "application/json",
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "true",
		},
	}
}

func errorResponse(statusCode int, message string) events.APIGatewayProxyResponse {
	type Response struct {
		Message string `json:"message"`
	}
	body, _ := json.Marshal(Response{Message: message})
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                     "application/json",
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "true",
		},
		Body: string(body),
	}
}

func main() {
	lambda.Start(subscribe)
}
