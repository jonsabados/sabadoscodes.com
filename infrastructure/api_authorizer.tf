data "aws_iam_policy_document" "auth_lambda_policy" {
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

resource "aws_iam_role" "auth_lambda_role" {
  name               = "${local.workspace_prefix}authLambdaRole"
  assume_role_policy = data.aws_iam_policy_document.assume_lambda_role_policy.json

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_iam_role_policy" "auth_lambda_role_policy" {
  role   = aws_iam_role.auth_lambda_role.name
  policy = data.aws_iam_policy_document.auth_lambda_policy.json
}

resource "aws_lambda_function" "auth_lambda" {
  filename         = "../dist/authorizerLambda.zip"
  source_code_hash = filebase64sha256("../dist/authorizerLambda.zip")
  handler          = "authorizer"
  function_name    = "${local.workspace_prefix}authorizer"
  role             = aws_iam_role.auth_lambda_role.arn
  runtime          = "go1.x"

  tracing_config {
    mode = "Active"
  }

  environment {
    variables = {
      "LOG_LEVEL": "debug"
      "ROOT_USER": data.aws_ssm_parameter.root_user.value,
      "GOOGLE_CLIENT_ID": data.aws_ssm_parameter.google_client_id.value,
      "ACCOUNT_ID": data.aws_caller_identity.current.account_id,
      "API_ID": aws_api_gateway_rest_api.api.id,
      "STAGE": "${local.workspace_prefix}main" // referencing the stage creates a circular dependency
    }
  }

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_cloudwatch_log_group" "auth_lambda_logs" {
  name              = "/aws/lambda/${aws_lambda_function.auth_lambda.function_name}"
  retention_in_days = 7

  tags = {
    Workspace = terraform.workspace
  }
}

data "aws_iam_policy_document" "api_gateway_authorizer_invocation_policy" {
  statement {
    sid       = "AllowAuthLambdaInvocation"
    effect    = "Allow"
    actions   = [
      "lambda:InvokeFunction"
    ]
    resources = [
      aws_lambda_function.auth_lambda.arn
    ]
  }
}

resource "aws_iam_role" "api_gateway_authorizer_invocation_role" {
  name = "${local.workspace_prefix}api_gateway_auth_invocation"
  path = "/"

  assume_role_policy = data.aws_iam_policy_document.assume_gateway_role_role_policy.json

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_iam_role_policy" "api_gateway_authorizer_invocation_role_policy" {
  name = "default"
  role = aws_iam_role.api_gateway_authorizer_invocation_role.id

  policy = data.aws_iam_policy_document.api_gateway_authorizer_invocation_policy.json
}

resource "aws_api_gateway_authorizer" "gateway_authorizer" {
  name                           = "${local.workspace_prefix}sabadoscodes.com-auth"
  rest_api_id                    = aws_api_gateway_rest_api.api.id
  authorizer_uri                 = aws_lambda_function.auth_lambda.invoke_arn
  authorizer_credentials         = aws_iam_role.api_gateway_authorizer_invocation_role.arn
  type                           = "TOKEN"
}