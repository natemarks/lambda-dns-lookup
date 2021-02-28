module "aws_lambda_monitor" {
  source = "github.com/natemarks/tf-aws-lambda-monitor?ref=v0.0.11"

  aws_account_id               = "0123456789012"
  aws_region                   = "us-east-1"
  function_name                = "lambda-dns-lookup-monitor"
  handler_name                 = "lambda-dns-lookup"
  function_runtime             = "go1.x"
  tracing_mode                 = "PassThrough"
  source_bucket                = "bucket_where_i_keep_my_lambda_app"
  source_key                   = "lambda-dns-lookup-monitor/lambda-dns-lookup.zip"
  opsgenie_https_sns_endpoint  = "https://api.opsgenie.com/v1/json/cloudwatch?apiKey=SOME_API_KEY"
  lambda_log_retention_in_days = 14

  custom_tags = {
    terraform   = "true"
    environment = "production"
  }
  environment_variables = {
    DEBUG           = "false"
    RANDOM_FAILURES = "false"
    LOOKUPS = jsonencode(
      [
        {
          "Target" : "rpapi.cts.imprivata.com",
          "ExpectedResponses" : 4
        }
      ]
    )
  }
  schedule_expression      = "rate(1 minute)"
  schedule_expression_desc = "Every minute"
}
