resource "aws_s3_bucket" "backup_bucket" {
  bucket = "${local.workspace_prefix}sabadocodes.backups.${data.aws_caller_identity.current.account_id}"
  acl    = "private"

  versioning {
    enabled = true
  }

  lifecycle_rule {
    enabled = true

    noncurrent_version_transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    noncurrent_version_transition {
      days          = 60
      storage_class = "GLACIER"
    }

    noncurrent_version_expiration {
      days = 90
    }
  }
}

data "aws_iam_policy_document" "backup_lambda_policy" {
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

  statement {
    sid       = "AllowAssetBucketRead"
    effect    = "Allow"
    actions   = [
      "s3:GetObject"
    ]
    resources = ["${aws_s3_bucket.article_assets_bucket.arn}/*"]
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

  statement {
    sid       = "AllowBackupBucketUpload"
    effect    = "Allow"
    actions   = [
      "s3:PutObject",
      "s3:PutObjectAcl"
    ]
    resources = ["${aws_s3_bucket.backup_bucket.arn}/*"]
  }
}

module "backup_lambda" {
  source           = "./lambda"
  workspace_prefix = local.workspace_prefix
  lambda_name      = "backup"
  lambda_policy    = data.aws_iam_policy_document.backup_lambda_policy.json
  timeout          = 15

  env_variables = {
    LOG_LEVEL     = "info"
    ASSET_BUCKET  = aws_s3_bucket.article_assets_bucket.bucket
    TARGET_BUCKET = aws_s3_bucket.backup_bucket.bucket
    ARTICLE_TABLE = aws_dynamodb_table.article_store.name
  }
}

resource "aws_cloudwatch_event_rule" "every_day_at_midnight" {
  name                = "${local.workspace_prefix}every-day-at-midnight"
  description         = "Fires every day at midnight"
  schedule_expression = "cron(0 0 * * ? *)"
}

resource "aws_cloudwatch_event_target" "run_backup" {
  rule      = aws_cloudwatch_event_rule.every_day_at_midnight.name
  target_id = "lambda"
  arn       = module.backup_lambda.arn
}

resource "aws_lambda_permission" "allow_cloudwatch_to_call_backup" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = module.backup_lambda.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.every_day_at_midnight.arn
}