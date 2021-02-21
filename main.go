package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

var client = lambda.New(session.New())

type lookupResult struct {
	NumberOfAddresses int      `json:"numberOfAddresses"`
	Responses         []string `json:"responses"`
}

func emitStructuredEvent(eventData map[string]string) {
	logMsg, _ := json.Marshal(eventData)
	log.Println(string(logMsg))
}

func executeLookups(targets []string) (string, error) {
	for _, target := range targets {
		logData := map[string]string{"target": target}
		logMsg, _ := json.Marshal(logData)
		log.Println(string(logMsg))

		addresses, err := net.LookupHost(target)
		if err != nil {
			errmsg := fmt.Sprintf("job status: failed %s", target)
			return errmsg, err
		}
		res := lookupResult{Responses: addresses, NumberOfAddresses: len(addresses)}
		jsonString, _ := json.Marshal(res)
		log.Println(string(jsonString))
	}
	return "job status: success", nil
}

func environmentMap() map[string]string {
	items := make(map[string]string)
	for _, item := range os.Environ() {
		splits := strings.Split(item, "=")
		items[splits[0]] = splits[1]
	}
	return items
}

func debugLogging(ctx context.Context, event events.CloudWatchEvent) {
	// log event
	eventJSON, _ := json.MarshalIndent(event, "", "  ")
	log.Printf("EVENT: %s", eventJSON)
	// log environment variables
	emitStructuredEvent(environmentMap())

	ctxData := make(map[string]string)

	// request context
	lc, _ := lambdacontext.FromContext(ctx)
	//request context
	ctxData["REQUEST ID"] = lc.AwsRequestID
	// global variable
	ctxData["FUNCTION NAME"] = lambdacontext.FunctionName
	// context method
	deadline, _ := ctx.Deadline()
	ctxData["DEADLINE"] = deadline.String()
	//log some context attributes
	emitStructuredEvent(ctxData)
}

func handleRequest(ctx context.Context, event events.CloudWatchEvent) (string, error) {
	debugLogging(ctx, event)
	targets := []string{"rpapi.cts.imprivata.com"}
	executeLookups(targets)
	return "FunctionCount", nil
}

func main() {
	runtime.Start(handleRequest)
}
