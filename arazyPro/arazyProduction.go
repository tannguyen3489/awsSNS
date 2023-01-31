package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/go-redis/redis"
	"github.com/go-resty/resty/v2"
	"log"
	"os"
	"os/exec"
	"time"
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
const RestartServerCommand string = "/home/ec2-user/restartBE.sh"

func main() {

	// setup loge
	file, err := os.OpenFile("./go.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	logger := log.New(file, "Arazy Log", log.LstdFlags)

	logger.Println("Start checking...")

	// for testing
	//notifyWhenServerDown(logger)

	// Create a Resty Client
	client := resty.New()

	client.SetTimeout(5 * time.Second)

	checkingUrl := "https://app.arazygroup.com/be/checkhealth"

	resp, err := client.R().
		EnableTrace().
		Get(checkingUrl)

	isCantConnect := false
	if err != nil {
		logger.Println("Server Not Response", err)
		resp1, err := client.R().
			EnableTrace().
			Get(checkingUrl)
		resp = resp1
		if err != nil {
			resp2, err := client.R().
				EnableTrace().
				Get(checkingUrl)
			resp = resp2
			if err != nil {
				isCantConnect = true
			}
		}
	}

	if isCantConnect == true {
		logger.Println("Server Not Response", err)
		notifyWhenServerDown(logger)

	} else {
		responseTime := resp.Time().Seconds()

		logger.Println("  Time       :", responseTime)

		if responseTime > 3 {
			notifyWhenServerSlow(logger)
		}
	}

	logger.Println("Everything is ok")

	//redisClient := rClient()
	//pingRedis(redisClient)

	//out, err1 := exec.Command("redis-cli PING").Output()
	//
	//if err1 != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Println(string(out))

}

func rClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	return client
}

func pingRedis(client *redis.Client) (string, error) {
	pong, err := client.Ping().Result()
	if err != nil {
		return "", err
	}
	fmt.Println(pong, err)
	return pong, err
}

func notifyWhenServerSlow(logger *log.Logger) {
	msg := "Server is slow for now"
	logger.Println("Server is slow for now")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	//
	//cfg, err := config.LoadDefaultConfig(context.TODO(),
	//	config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("awsid", "awssecret", "")),
	//)

	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := sns.NewFromConfig(cfg)

	subject := "The Arazy Production Server Is Slowly Now"

	arnChannel := ARNChannel

	var input = &sns.PublishInput{
		Subject:  &subject,
		Message:  &msg,
		TopicArn: &arnChannel,
	}
	result, err := PublishMessage(context.TODO(), client, input)
	if err != nil {
		logger.Println("Got an error publishing the message:")
		logger.Println(err)
		return
	}

	logger.Println("Message ID: " + *result.MessageId)
}

func notifyWhenServerDown(logger *log.Logger) {
	logger.Println("exec restart BE command ")

	msg := "Server is down, restarting now."

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client := sns.NewFromConfig(cfg)

	subject := "The Arazy Production Server is down. Restarting..."

	arnChannel := ARNChannel

	var input = &sns.PublishInput{
		Subject:  &subject,
		Message:  &msg,
		TopicArn: &arnChannel,
	}
	result, err := PublishMessage(context.TODO(), client, input)

	cmd := exec.Command("sudo /bin/sh", RestartServerCommand)
	_, restartBeError := cmd.Output()

	logger.Println("error when exec restart BE command ", restartBeError)

	if err != nil {
		logger.Println("Got an error publishing the message:")
		logger.Println(err)
		return
	}

	logger.Println("Message ID: " + *result.MessageId)
}
