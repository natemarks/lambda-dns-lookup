# lambda-dns-lookup

This lambda monitors the results of important DNS lookups to make sure that each returns the expected number of IP addresses.  I use this terraform module to deploy the lambda and the alarm pipeline:
https://github.com/natemarks/tf-aws-lambda-monitor

Once deployed, the lmbda can be configured wiht these environment variables:

DEBUG: [default: false] enable debug logging
RANDOM_FAILURES: [default: false] randomly fire alarms for the difference failure modes
LOOKUPS: json string that tells the lambda what FQDNs to check and how many addresses to expect for each

## Understanding the alarms

Severity 1: Too few addresses -  This is what we're looking for. It will cause intermittent failures for customers who use FQDN-based firewall ACLS. Somebody should wake up for this.

Severity 2: lookup failed -  This is a blind spot for a test that shouldn't fail very often. It's ok to get to it next business day (NBD)

Severity 2: Can't parse LOOKUPS json - This is a blind spot for a test that shouldn't fail very often. It's ok to get to it next business day (NBD)

Severity 2: LOOKUPS env var is unset - This is a blind spot for a test that shouldn't fail very often. It's ok to get to it next business day (NBD)
## Deployment


### Build and push the lambda app to your bucket
 - copy example/config.json to the root
 - edit it accordingly
 - configure your aws client 
 - run the following to build, izp and push the app to your bucket

```bash
make compile
```

### Deploy the lambda -> alarm pipeline


 - copy example/deploy.tf to the root
 - edit it accordingly
 - configure your aws client and make sure you have terraform installed

NOTE: You need to get the OpsGenie endpoint from the opsgenie cloudwatch integration

 - run the following to create the lambda, logs, metrics, filters,alarms, SNS

```bash
terraform int && terraform plan
terraform apply
```
## Future

### use KMS/Parameter store instead of env vars

The function is configured by injecting environment variables using terraform.

NOTE:  There is a [limit](https://aws.amazon.com/premiumsupport/knowledge-center/lambda-environment-variable-size/#:~:text=The%20default%20quota%20value%20of,use%20an%20external%20data%20store.) on the total size of al environment variable data. If the test request has to exceed that, the project should be extended to get the data from parameter store