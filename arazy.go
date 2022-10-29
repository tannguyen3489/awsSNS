package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"net/http"
)

// SNSPublishAPI defines the interface for the Publish function.
// We use this interface to test the function using a mocked service.
type SNSPublishAPI interface {
	Publish(ctx context.Context,
		params *sns.PublishInput,
		optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

func PublishMessage(c context.Context, api SNSPublishAPI, input *sns.PublishInput) (*sns.PublishOutput, error) {
	return api.Publish(c, input)
}

const ARNChannel string = "arn:aws:sns:us-east-1:412320870653:arazy"

func main() {
	resp, err := http.Get("http://example.com/")
	if err != nil {
		fmt.Println("Server Down")
	} else {
	}

}

func notifyWhenServerSlow() {
	msg := "Server is slow for now"

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := sns.NewFromConfig(cfg)

	subject := "tantest"

	arnChannel := ARNChannel

	var input = &sns.PublishInput{
		Subject:  &subject,
		Message:  &msg,
		TopicArn: &arnChannel,
	}
	result, err := PublishMessage(context.TODO(), client, input)
	if err != nil {
		fmt.Println("Got an error publishing the message:")
		fmt.Println(err)
		return
	}

	fmt.Println("Message ID: " + *result.MessageId)
}
