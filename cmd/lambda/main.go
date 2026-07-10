package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/obaid/pr-guardian/internal/checker"
)

// Request is the JSON body the web dashboard sends.
type Request struct {
	PRURL string `json:"pr_url"`
	Token string `json:"token"`
}

// Response is what we send back to the browser.
type Response struct {
	Results []checker.CheckResult `json:"results"`
	Error   string                `json:"error,omitempty"`
}

// handler is the Lambda function entry point.
func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// CORS headers so the browser can call this from any origin
	headers := map[string]string{
		"Content-Type":                 "application/json",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Headers": "Content-Type",
		"Access-Control-Allow-Methods": "POST, OPTIONS",
	}

	// Handle preflight OPTIONS request (browser CORS check)
	if req.HTTPMethod == http.MethodOptions {
		return events.APIGatewayProxyResponse{StatusCode: 200, Headers: headers}, nil
	}

	// Parse request body
	var body Request
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil || body.PRURL == "" {
		return jsonResponse(400, Response{Error: "Invalid request. Send {\"pr_url\": \"...\", \"token\": \"...\"}"}, headers), nil
	}

	// Run checks
	results, err := checker.RunChecks(ctx, body.PRURL, body.Token)
	if err != nil {
		return jsonResponse(400, Response{Error: err.Error()}, headers), nil
	}

	return jsonResponse(200, Response{Results: results}, headers), nil
}

func jsonResponse(statusCode int, body interface{}, headers map[string]string) events.APIGatewayProxyResponse {
	b, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       string(b),
	}
}

func main() {
	lambda.Start(handler)
}
