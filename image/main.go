package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/slack-go/slack"
)

var (
	signingSecret = os.Getenv("SLACK_SIGNING_SECRET")
	//api           = slack.New(os.Getenv("SLACK_BOT_TOKEN"))
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	verifier, err := slack.NewSecretsVerifier(request.MultiValueHeaders, signingSecret)
	if err != nil {
		return handleServerError(err, "slack.NewSecretsVerifier_")
	}

	reqAccessor := &core.RequestAccessor{}
	httpReq, err := reqAccessor.ProxyEventToHTTPRequest(request)
	if err != nil {
		return handleServerError(err, "reqAccessor.ProxyEventToHTTPRequest_")
	}

	httpReq.Body = ioutil.NopCloser(io.TeeReader(httpReq.Body, &verifier))
	slashCommand, err := slack.SlashCommandParse(httpReq)
	if err != nil {
		return handleServerError(err, "slack.SlashCommandParse_")
	}

	if err = verifier.Ensure(); err != nil {
		return handleServerError(err, "verifier.Ensure_")
	}

	switch slashCommand.Command {
	case "/hiroyuki":
		texts := regexp.MustCompile(`[\s　]+`).Split(slashCommand.Text, -1)
		if len(texts) < 2 {
			return handleBadRequest("Few arguments"), nil
		}

		switch texts[0] {
		case "nandarou":
			//if _, _, err = api.PostMessage(slashCommand.ChannelID, slack.MsgOptionText(strings.Join(texts, " "), false)); err != nil {
			//	return handleServerError(err, "api.PostMessage_")
			//}
			return events.APIGatewayProxyResponse{StatusCode: 200, Body: texts[1]}, nil
		}
	}

	return handleBadRequest(fmt.Sprintf("Bad Request:%v", request)), nil
}

func handleBadRequest(message string) events.APIGatewayProxyResponse {
	log.Printf("[INFO] %s\n", message)
	return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Bad Request"}
}

func handleServerError(err error, prefix string) (events.APIGatewayProxyResponse, error) {
	log.Printf("[ERROR] %s\n", prefix+err.Error())
	return events.APIGatewayProxyResponse{StatusCode: 500, Body: prefix + err.Error()}, err
}

func main() {
	lambda.Start(handler)
}
