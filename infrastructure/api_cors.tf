data "aws_iam_policy_document" "cors_lambda_policy" {
  statement {
    sid       = "AllowLogging"
    effect    = "Allow"
    actions   = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = [
      "arn:aws:logs:*:*:*"
    ]
  }

  statement {
    sid       = "AllowXRayWrite"
    effect    = "Allow"
    actions   = [
      "xray:PutTraceSegments",
      "xray:PutTelemetryRecords",
      "xray:GetSamplingRules",
      "xray:GetSamplingTargets",
      "xray:GetSamplingStatisticSummaries"
    ]
    resources = ["*"]
  }
}

resource "aws_iam_role" "cors_lambda_role" {
  name               = "corsLambdaRole"
  assume_role_policy = data.aws_iam_policy_document.assume_lambda_role_policy.json
}

resource "aws_iam_role_policy" "cors_lambda_role_policy" {
  role   = aws_iam_role.cors_lambda_role.name
  policy = data.aws_iam_policy_document.cors_lambda_policy.json
}

resource "aws_lambda_function" "cors_lambda" {
  filename         = "../dist/corsLambda.zip"
  source_code_hash = filebase64sha256("../dist/corsLambda.zip")
  handler          = "cors"
  function_name    = "cors"
  role             = aws_iam_role.cors_lambda_role.arn
  runtime          = "go1.x"

  environment {
    variables = {
      ALLOWED_ORIGINS = "https://${data.aws_ssm_parameter.domain_name.value},https://www.${data.aws_ssm_parameter.domain_name.value},http://localhost:8080"
    }
  }
}

resource "aws_lambda_permission" "cors_allow_gateway_invoke" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.cors_lambda.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "arn:aws:execute-api:us-east-1:${data.aws_caller_identity.current.account_id}:${aws_api_gateway_rest_api.api.id}/*/OPTIONS/${aws_api_gateway_resource.wildcard.path_part}"
}

resource "aws_cloudwatch_log_group" "cors_lambda_logs" {
  name              = "/aws/lambda/${aws_lambda_function.cors_lambda.function_name}"
  retention_in_days = 7
}

resource "aws_api_gateway_resource" "wildcard" {
  parent_id   = aws_api_gateway_rest_api.api.root_resource_id
  path_part   = "{proxy+}"
  rest_api_id = aws_api_gateway_rest_api.api.id
}

resource "aws_api_gateway_method" "options" {
  authorization = "NONE"
  http_method   = "OPTIONS"
  resource_id   = aws_api_gateway_resource.wildcard.id
  rest_api_id   = aws_api_gateway_rest_api.api.id
}

resource "aws_api_gateway_integration" "cors_integration" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.wildcard.id
  http_method             = aws_api_gateway_method.options.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.cors_lambda.invoke_arn
}