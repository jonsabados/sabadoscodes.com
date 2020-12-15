data "aws_iam_policy_document" "assume_lambda_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      identifiers = ["lambda.amazonaws.com"]
      type        = "Service"
    }
    effect  = "Allow"
    sid     = "AllowLambdaAssumeRole"
  }
}

resource "aws_iam_role" "lambda_role" {
  name               = "${var.workspace_prefix}${var.lambda_name}LambdaRole"
  assume_role_policy = data.aws_iam_policy_document.assume_lambda_role_policy.json

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_iam_role_policy" "lambda_role_policy" {
  role   = aws_iam_role.lambda_role.name
  policy = var.lambda_policy
}

resource "aws_lambda_function" "lambda" {
  filename         = "../dist/${var.lambda_name}Lambda.zip"
  source_code_hash = filebase64sha256("../dist/${var.lambda_name}Lambda.zip")
  handler          = var.lambda_name
  function_name    = "${var.workspace_prefix}${var.lambda_name}"
  role             = aws_iam_role.lambda_role.arn
  runtime          = "go1.x"
  timeout          = var.timeout

  tracing_config {
    mode = "Active"
  }

  environment {
    variables = var.env_variables
  }

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_cloudwatch_log_group" "lambda_logs" {
  name              = "/aws/lambda/${aws_lambda_function.lambda.function_name}"
  retention_in_days = 7

  tags = {
    Workspace = terraform.workspace
  }
}