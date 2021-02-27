# lambda-dns-lookup

This lambda monitors the results of important DNS lookups to make sure that each returns the expected number of IP addresses.  Results are logged to cloudwatch using JSON so it's easy to parse the results into metrics , which are in turn used to trigger alarms.

There is currently no way to alarm directly alarm on an event. the metrics stage is required

## Monitor output
All tests run against some target under test. There is a difference between test errors (anything that prevents the test from running correctly) and target errors (the test runs and the target is failing).  The first creates blind spot, the second indicates that the resource we care about is failing.

This DNS test shouldn't fail often so having a blind spot is a severity 2. We don't need to wake anybody up for it.  It's ok to fix it the next business day

## Deployment

configure your aws client credentials with access to the account the bucket to store the 
Configure deploy.tf and run terraform init/plan/apply

## Configure

The function is configured by injecting environment variables using terraform.

NOTE:  There is a [limit](https://aws.amazon.com/premiumsupport/knowledge-center/lambda-environment-variable-size/#:~:text=The%20default%20quota%20value%20of,use%20an%20external%20data%20store.) on the total size of al environment variable data. If the test request has to exceed that, the project should be extended to get the data from parameter store

**DEBUG:** 'DEBUG' variable enables debug logging in the lambda.  It is enabled by setting the value to 'true' The check is case-insensitive. 

**RANDOM_FAILURES:** 'RANDOM_FAILURES' variable causes the function to log random failures.  All of the failure modes should be covered eventually. This is important to creating the metrics correctly when the function is new OR if there are changes in AWS that force us to re-tune events and metrics.


**LOOKUPS:** 'LOOKUPS' variable value is a json string that is unmarshalled into a list of lookupRequest structs. It allows us to configure the tests and expectations . If this is unset, the function will run a default test.  

NOTE: If it gets bad JSON, it's probably not going t handle it well