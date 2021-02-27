// given a slice of FQDNs to resolve and the expected minimum number of
// addresses, thorw wrros if the resolution fails or the count is 0.
// log a target specific alarm if the number of addresses it smaller than
// expected
// if an nslookupfails completly, error the lambda, and log an event
// if it succeeds but the count is wrong log an alarm event
// turn on debug loggging with "DEBUG"= "TRUE"
// turn on random errors wiht "RANDOM_FAILURES" = TRUE
// override lookup request with LOOKUPS = [valid json object]
//ex. job = `[{"Target": "www.google.com"}, {"ExpectedResponses": 1}]`

// Event reserved metadata
// imprivata_event_type: TEST_RESULT
// devops maintains the test AND the resource under test
// imprivata_event_audience: DEVOPS
// imprivata_event_severity:
//  - 1: wake someone up
//  - 2: get this in front of someone next business day
//  - 3: informational: lives and dies in a log archive
//  - 4: debug/diag: debug logging or the output of a diagnostic request

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

var client = lambda.New(session.New())

type lookupRequest struct {
	Target            string `json:"target"`
	ExpectedResponses int    `json:"expectedResponses"`
}
type lookupResult struct {
	NumberOfAddresses int      `json:"numberOfAddresses"`
	Responses         []string `json:"responses"`
}

// validate data n the LOOKUPS env var
// handle problems as appropriate
func lookups() ([]lookupRequest, string, error) {
	var r []lookupRequest
	defaultJob := `[{"Target": "www.google.com", "ExpectedResponses": 1}]`

	lv := os.Getenv("LOOKUPS")
	//getenv returns "" for empty AND unset
	if lv == "" {
		alarmBadLookupVar()
		json.Unmarshal([]byte(defaultJob), &r)
		msg := fmt.Sprintf("required LOOKUPS var is unset")
		return r, msg, errors.New(msg)
	}

	if err := json.Unmarshal([]byte(lv), &r); err != nil {
		json.Unmarshal([]byte(defaultJob), &r)
		msg, _ := alarmBadLookupVar()
		return r, msg, err
	}
	return r, "Got valid env LOOKUPS variable setting", nil
}

// called when the value of LOOKUPS can't be parsed as JSON
// logs the correct data
// returns proper out and error
func alarmBadLookupVar() (string, error) {

	alarmData := make(map[string]string)
	//set audience
	alarmData["imprivata_event_audience"] = "DEVOPS"
	// set severity to 2. It's a blindspot on a test  that shouldn't fail often
	// so if it waits until the next day, we're ok
	alarmData["imprivata_event_severity"] = "2"
	// alarm detail
	msg := fmt.Sprintf("Bad LOOKUPS value. Unable to parse JSON")
	alarmData["imprivata_event_message"] = msg

	emitStructuredEvent(alarmData, 1)
	return msg, errors.New(msg)
}

// called when the required var LOOKUPS is not set
// logs the correct data
// returns proper out and error
func alarmUnsetLookupVar() (string, error) {

	alarmData := make(map[string]string)
	//set audience
	alarmData["imprivata_event_audience"] = "DEVOPS"
	// set severity to 2. It's a blindspot on a test  that shouldn't fail often
	// so if it waits until the next day, we're ok
	alarmData["imprivata_event_severity"] = "2"
	// alarm detail
	msg := fmt.Sprintf("Bad LOOKUPS value. The required varaible is unset")
	alarmData["imprivata_event_message"] = msg

	emitStructuredEvent(alarmData, 1)
	return msg, errors.New(msg)
}

// called for each host lookup failure
// logs the correct data
// returns proper out and error
func alarmHostLookupFailed(h string) (string, error) {

	alarmData := make(map[string]string)
	//set audience
	alarmData["imprivata_event_audience"] = "DEVOPS"
	// set severity to 2. It's a blindspot on a test  that shouldn't fail often
	// so if it waits until the next day, we're ok
	alarmData["imprivata_event_severity"] = "2"
	// alarm detail
	msg := fmt.Sprintf("Lookup failed for: %s", h)
	alarmData["imprivata_event_message"] = msg

	emitStructuredEvent(alarmData, 1)
	return msg, errors.New(msg)
}

// Called when the lookup works, but we don't get enough addresses back
// This failure is the reason the monitor exists.
// It breaks customers wiht FQDN based ACLs causing intermittent cloud
// connecton failures
func alarmTooFewAddresses(addr string, expected int, actual int) (string, error) {

	alarmData := make(map[string]string)
	//set audience
	alarmData["imprivata_event_audience"] = "DEVOPS"
	// set severity to 1. this is the reason the monitor was created
	alarmData["imprivata_event_severity"] = "1"
	// alarm detail
	msg := fmt.Sprintf("Too few addresses for %s. expected %d. got %d", addr, 4, 1)
	alarmData["imprivata_event_message"] = msg

	emitStructuredEvent(alarmData, 1)
	return msg, errors.New(msg)
}

// Called to log a successful lookup test
func goodLookupResult(t string, e int, a int) {

	successEvent := map[string]string{}
	successEvent["target"] = t
	successEvent["expectedAddressCount"] = fmt.Sprint(e)
	successEvent["actualAddressCount"] = fmt.Sprint(a)
	// log the success
	emitStructuredEvent(successEvent, 3)
}

// Randomly select from all of the valid outcomes and log accordingly
// This is critical for generating all the possible output data in order
// to tune the metrics log matching
func failRandomly(req []lookupRequest) (string, error) {
	failures := []string{
		"lookup_error",
		"too_few_addresses",
		"real_execution",
		"unset_lookup_var",
		"bad_json_lookup_var",
	}
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator

	switch failure := failures[rand.Intn(len(failures))]; failure {
	case "unset_lookup_var":
		return alarmUnsetLookupVar()
	case "bad_json_lookup_var":
		return alarmBadLookupVar()
	case "lookup_error":
		return alarmHostLookupFailed("fake_random_failure")
	case "too_few_addresses":
		alarmTooFewAddresses("fake_random_failure", 4, 1)
		errMsg := "Too few addresses for fake_random_failure. expected 4. got 1"
		return errMsg, nil
	case "real_execution":
		return executeLookups(req)
	default:
		return executeLookups(req)
	}
}

func emitStructuredEvent(eventData map[string]string, severity int) {
	eventData["imprivata_event_type"] = "TEST_RESULT"
	eventData["imprivata_event_audience"] = "DEVOPS"
	eventData["imprivata_event_severity"] = fmt.Sprint(severity)
	logMsg, _ := json.Marshal(eventData)
	log.Println(string(logMsg))
}

func executeLookups(req []lookupRequest) (string, error) {
	// Use testErrors to track status without breaking out on a failure.  If we
	// break on a bad lookup, we miss an opportunity to catch a severity 1 in
	//future loop iterations
	var testErrors error = nil
	var testOut string = "job status: success"

	// iterate on targets
	for _, target := range req {
		//log target message before attempt
		logData := map[string]string{"target": target.Target}
		logMsg, _ := json.Marshal(logData)
		log.Println(string(logMsg))

		// attempt lookup
		addresses, err := net.LookupHost(target.Target)
		if err != nil {
			alarmHostLookupFailed(target.Target)
			testOut = fmt.Sprintf("One or more lookups failed. see logs for details")
			testErrors = err
			continue
		}
		// evaluate lookup response
		res := lookupResult{Responses: addresses, NumberOfAddresses: len(addresses)}
		// bad result, throw alarm
		if res.NumberOfAddresses < target.ExpectedResponses {
			return alarmTooFewAddresses(target.Target, target.ExpectedResponses, res.NumberOfAddresses)
		}
		//if debug mode, dump the result
		if debugMode() {
			jsonString, _ := json.Marshal(res)
			log.Println(string(jsonString))
		}
		// Log the success summary
		goodLookupResult(target.Target, target.ExpectedResponses, res.NumberOfAddresses)
	}

	return testOut, testErrors
}

// read all environment variables into a map
// makes it easy to acces and to dump to json log
func environmentMap() map[string]string {
	items := make(map[string]string)
	for _, item := range os.Environ() {
		splits := strings.Split(item, "=")
		items[splits[0]] = splits[1]
	}
	return items
}

// Called when debug logging is enable to dump context data
func debugLogging(ctx context.Context, event events.CloudWatchEvent) {
	// log event
	eventJSON, _ := json.MarshalIndent(event, "", "  ")
	log.Printf("EVENT: %s", eventJSON)
	// log environment variables
	emitStructuredEvent(environmentMap(), 4)

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
	emitStructuredEvent(ctxData, 4)
}

// return true if in debug mode
func debugMode() bool {
	res := os.Getenv("DEBUG")
	return strings.EqualFold("true", res)
}

// return true if in random fail mode
func failMode() bool {
	res := os.Getenv("RANDOM_FAILURES")
	return strings.EqualFold("true", res)
}

// write the funciton verson on execution
func logVersion() {
	version := "0.1.2"
	e := make(map[string]string)
	e["version"] = version
	emitStructuredEvent(e, 3)
}

// custom handler is the entry point for the function
func handleRequest(ctx context.Context, event events.CloudWatchEvent) (string, error) {

	// log the version
	logVersion()
	// if debug mode, log all the context info
	if debugMode() {
		debugLogging(ctx, event)
	}
	// validate the LOOKUPS env var data
	l, out, err := lookups()
	if err != nil {
		return out, err
	}
	// if running in fail-mode, fail randomly
	if failMode() {
		return failRandomly(l)
	}
	return executeLookups(l)

}

func main() {
	runtime.Start(handleRequest)
}
