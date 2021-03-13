resource "aws_api_gateway_resource" "article_asset" {
  parent_id   = aws_api_gateway_resource.article.id
  path_part   = "asset"
  rest_api_id = aws_api_gateway_rest_api.api.id
}

data "aws_iam_policy_document" "article_asset_list_policy" {
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
    sid       = "AllowAssetBucketList"
    effect    = "Allow"
    actions   = [
      "s3:ListBucket"
    ]
    resources = [aws_s3_bucket.article_assets_bucket.arn]
  }
}

module "article_asset_list_lambda" {
  source           = "./lambda"
  workspace_prefix = local.workspace_prefix
  lambda_name      = "articleAssetList"
  lambda_policy    = data.aws_iam_policy_document.article_asset_list_policy.json
  env_variables    = {
    LOG_LEVEL       = "info"
    ALLOWED_ORIGINS = "https://${aws_acm_certificate.ui_cert.domain_name},https://${aws_acm_certificate.ui_cert.subject_alternative_names[0]},http://localhost:8080"
    ASSET_BUCKET    = aws_s3_bucket.article_assets_bucket.bucket
    BASE_ASSET_URL  = "https://${aws_acm_certificate.ui_cert.domain_name}/article-assets"
  }
}

resource "aws_api_gateway_method" "article_asset_list" {
  authorization = "CUSTOM"
  authorizer_id = aws_api_gateway_authorizer.gateway_authorizer.id
  http_method   = "GET"
  resource_id   = aws_api_gateway_resource.article_asset.id
  rest_api_id   = aws_api_gateway_rest_api.api.id
}

resource "aws_api_gateway_integration" "article_asset_list" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.article_asset.id
  http_method             = aws_api_gateway_method.article_asset_list.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = module.article_asset_list_lambda.invoke_arn
}

resource "aws_lambda_permission" "article_asset_list_allow_gateway_invoke" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = module.article_asset_list_lambda.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "arn:aws:execute-api:us-east-1:${data.aws_caller_identity.current.account_id}:${aws_api_gateway_rest_api.api.id}/*/GET/${aws_api_gateway_resource.article.path_part}/${aws_api_gateway_resource.article_asset.path_part}"
}

data "aws_iam_policy_document" "article_asset_upload_policy" {
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
    sid       = "AllowAssetBucketUpload"
    effect    = "Allow"
    actions   = [
      "s3:PutObject",
      "s3:PutObjectAcl"
    ]
    resources = ["${aws_s3_bucket.article_assets_bucket.arn}/*"]
  }
}

module "article_asset_upload_lambda" {
  source           = "./lambda"
  workspace_prefix = local.workspace_prefix
  lambda_name      = "articleAssetUpload"
  lambda_policy    = data.aws_iam_policy_document.article_asset_upload_policy.json
  env_variables    = {
    LOG_LEVEL       = "info"
    ALLOWED_ORIGINS = "https://${aws_acm_certificate.ui_cert.domain_name},https://${aws_acm_certificate.ui_cert.subject_alternative_names[0]},http://localhost:8080"
    ASSET_BUCKET    = aws_s3_bucket.article_assets_bucket.bucket
    BASE_ASSET_URL  = "https://${aws_acm_certificate.ui_cert.domain_name}/article-assets"
  }
}

resource "aws_api_gateway_method" "article_asset_upload" {
  authorization = "CUSTOM"
  authorizer_id = aws_api_gateway_authorizer.gateway_authorizer.id
  http_method   = "POST"
  resource_id   = aws_api_gateway_resource.article_asset.id
  rest_api_id   = aws_api_gateway_rest_api.api.id
}

resource "aws_api_gateway_integration" "article_asset_upload" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.article_asset.id
  http_method             = aws_api_gateway_method.article_asset_upload.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = module.article_asset_upload_lambda.invoke_arn
}

resource "aws_lambda_permission" "article_asset_upload_allow_gateway_invoke" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = module.article_asset_upload_lambda.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "arn:aws:execute-api:us-east-1:${data.aws_caller_identity.current.account_id}:${aws_api_gateway_rest_api.api.id}/*/POST/${aws_api_gateway_resource.article.path_part}/${aws_api_gateway_resource.article_asset.path_part}"
}
