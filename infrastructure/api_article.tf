resource "aws_api_gateway_resource" "article" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_rest_api.api.root_resource_id
  path_part   = "article"
}

resource "aws_api_gateway_resource" "article_slug" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.article.id
  path_part   = "slug"
}

resource "aws_api_gateway_resource" "article_by_slug" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.article_slug.id
  path_part   = "{slug}"
}

data "aws_iam_policy_document" "article_save_access_policy" {
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

  statement {
    sid       = "AllowArticleStoreAccess"
    effect    = "Allow"
    actions   = [
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:DescribeStream",
      "dynamodb:DescribeTable"
    ]
    resources = [
      "arn:aws:dynamodb:*:*:table/${aws_dynamodb_table.article_store.name}"
    ]
  }
}

module "article_save_lambda" {
  source           = "./lambda"
  workspace_prefix = local.workspace_prefix
  lambda_name      = "articleSave"
  lambda_policy    = data.aws_iam_policy_document.article_save_access_policy.json
  env_variables    = {
    LOG_LEVEL        = "info"
    ALLOWED_ORIGINS  = "https://${aws_acm_certificate.ui_cert.domain_name},https://${aws_acm_certificate.ui_cert.subject_alternative_names[0]},http://localhost:8080"
    BASE_ARTICLE_URL = "https://${aws_api_gateway_domain_name.api.domain_name}/article/slug"
    ARTICLE_TABLE    = aws_dynamodb_table.article_store.name
  }
}

resource "aws_api_gateway_method" "put_article_by_slug" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.article_by_slug.id
  http_method   = "PUT"
  authorization = "CUSTOM"
  authorizer_id = aws_api_gateway_authorizer.gateway_authorizer.id

  request_parameters = {
    "method.request.path.slug" = true
  }
}

resource "aws_api_gateway_integration" "article_save" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.article_by_slug.id
  http_method             = aws_api_gateway_method.put_article_by_slug.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = module.article_save_lambda.invoke_arn
}

resource "aws_lambda_permission" "article_save_allow_gateway_invoke" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = module.article_save_lambda.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "arn:aws:execute-api:us-east-1:${data.aws_caller_identity.current.account_id}:${aws_api_gateway_rest_api.api.id}/*/PUT/${aws_api_gateway_resource.article.path_part}/${aws_api_gateway_resource.article_slug.path_part}/${aws_api_gateway_resource.article_by_slug.path_part}"
}

data "aws_iam_policy_document" "article_read_access_policy" {
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

  statement {
    sid       = "AllowArticleStoreAccess"
    effect    = "Allow"
    actions   = [
      "dynamodb:Scan",
      "dynamodb:GetItem",
      "dynamodb:DescribeStream",
      "dynamodb:DescribeTable"
    ]
    resources = [
      "arn:aws:dynamodb:*:*:table/${aws_dynamodb_table.article_store.name}"
    ]
  }
}

module "article_get_lambda" {
  source           = "./lambda"
  workspace_prefix = local.workspace_prefix
  lambda_name      = "articleGet"
  lambda_policy    = data.aws_iam_policy_document.article_read_access_policy.json
  env_variables    = {
    LOG_LEVEL       = "info"
    ALLOWED_ORIGINS = "https://${aws_acm_certificate.ui_cert.domain_name},https://${aws_acm_certificate.ui_cert.subject_alternative_names[0]},http://localhost:8080"
    ARTICLE_TABLE   = aws_dynamodb_table.article_store.name
  }
}

resource "aws_api_gateway_method" "get_article_by_slug" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.article_by_slug.id
  http_method   = "GET"
  authorization = "CUSTOM"
  authorizer_id = aws_api_gateway_authorizer.gateway_authorizer.id

  request_parameters = {
    "method.request.path.slug" = true
  }
}

resource "aws_api_gateway_integration" "article_get" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.article_by_slug.id
  http_method             = aws_api_gateway_method.get_article_by_slug.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = module.article_get_lambda.invoke_arn
}

resource "aws_lambda_permission" "article_get_allow_gateway_invoke" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = module.article_get_lambda.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "arn:aws:execute-api:us-east-1:${data.aws_caller_identity.current.account_id}:${aws_api_gateway_rest_api.api.id}/*/GET/${aws_api_gateway_resource.article.path_part}/${aws_api_gateway_resource.article_slug.path_part}/${aws_api_gateway_resource.article_by_slug.path_part}"
}

module "article_list_lambda" {
  source           = "./lambda"
  workspace_prefix = local.workspace_prefix
  lambda_name      = "articleList"
  lambda_policy    = data.aws_iam_policy_document.article_read_access_policy.json
  env_variables    = {
    LOG_LEVEL       = "info"
    ALLOWED_ORIGINS = "https://${aws_acm_certificate.ui_cert.domain_name},https://${aws_acm_certificate.ui_cert.subject_alternative_names[0]},http://localhost:8080"
    ARTICLE_TABLE   = aws_dynamodb_table.article_store.name
  }
}

resource "aws_api_gateway_method" "list_articles" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.article.id
  http_method   = "GET"
  authorization = "CUSTOM"
  authorizer_id = aws_api_gateway_authorizer.gateway_authorizer.id

  request_parameters = {
    "method.request.querystring.published" = false
  }
}

resource "aws_api_gateway_integration" "article_list" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.article.id
  http_method             = aws_api_gateway_method.list_articles.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = module.article_list_lambda.invoke_arn
}

resource "aws_lambda_permission" "article_list_allow_gateway_invoke" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = module.article_list_lambda.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "arn:aws:execute-api:us-east-1:${data.aws_caller_identity.current.account_id}:${aws_api_gateway_rest_api.api.id}/*/GET/${aws_api_gateway_resource.article.path_part}"
}