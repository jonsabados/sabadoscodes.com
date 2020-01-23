resource "aws_ses_domain_identity" "ses_domain" {
  domain = data.aws_ssm_parameter.domain_name.value
}

resource "aws_route53_record" "ses_domain_verification_record" {
  zone_id = data.aws_route53_zone.main_domain.zone_id
  name    = "_amazonses.${data.aws_ssm_parameter.domain_name.value}"
  type    = "TXT"
  ttl     = "600"
  records = [aws_ses_domain_identity.ses_domain.verification_token]
}

resource "aws_ses_domain_dkim" "ses_domain_dkim" {
  domain = aws_ses_domain_identity.ses_domain.domain
}

resource "aws_route53_record" "dkim_dns_records" {
  count   = 3
  zone_id = data.aws_route53_zone.main_domain.zone_id
  name    = "${element(aws_ses_domain_dkim.ses_domain_dkim.dkim_tokens, count.index)}._domainkey.${data.aws_ssm_parameter.domain_name.value}"
  type    = "CNAME"
  ttl     = "600"
  records = ["${element(aws_ses_domain_dkim.ses_domain_dkim.dkim_tokens, count.index)}.dkim.amazonses.com"]
}

resource "aws_route53_record" "mx_record" {
  zone_id = data.aws_route53_zone.main_domain.zone_id
  name    = data.aws_ssm_parameter.domain_name.value
  type    = "MX"
  ttl     = "600"
  records = ["10 inbound-smtp.us-east-1.amazonaws.com"]
}

resource "aws_ses_email_identity" "support_email" {
  email = data.aws_ssm_parameter.support_email.value
}

locals {
  mail_bucket_name = "mail.${data.aws_ssm_parameter.domain_name.value}"
  support_email    = "support@${data.aws_ssm_parameter.domain_name.value}"
}

data "aws_iam_policy_document" "mail_bucket_policy" {
  statement {
    sid       = "AllowSESPuts"
    effect    = "Allow"
    principals {
      identifiers = ["ses.amazonaws.com"]
      type        = "Service"
    }
    actions   = ["s3:PutObject"]
    resources = ["arn:aws:s3:::${local.mail_bucket_name}/*"]
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "aws:Referer"
    }
  }
}

resource "aws_s3_bucket" "mail_bucket" {
  bucket = local.mail_bucket_name
  policy = data.aws_iam_policy_document.mail_bucket_policy.json
  acl    = "private"
}

resource "aws_cloudwatch_log_group" "support_forward_logs" {
  name              = "/aws/lambda/${aws_lambda_function.support_forward_lambda.function_name}"
  retention_in_days = 7
}

data "aws_iam_policy_document" "support_forward_lambda_policy" {
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
    sid       = "AllowListBucket"
    effect    = "Allow"
    actions   = [
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.mail_bucket.arn
    ]
  }

  statement {
    sid       = "AllowMailBucketReadWrite"
    effect    = "Allow"
    actions   = [
      "s3:GetObject",
      "s3:DeleteObject"
    ]
    resources = [
      "${aws_s3_bucket.mail_bucket.arn}/*"
    ]
  }

  statement {
    sid       = "AllowSESSendRawEmail"
    effect    = "Allow"
    actions   = [
      "ses:SendRawEmail"
    ]
    resources = [
      "*"
    ]
  }
}

resource "aws_iam_role" "support_forward_lambda_role" {
  name               = "supportForwarderLambdaRole"
  assume_role_policy = data.aws_iam_policy_document.assume_lambda_role_policy.json
}

resource "aws_iam_role_policy" "support_forward_lambda_role_policy" {
  role   = aws_iam_role.support_forward_lambda_role.name
  policy = data.aws_iam_policy_document.support_forward_lambda_policy.json
}

resource "aws_lambda_function" "support_forward_lambda" {
  filename         = "../dist/forwarder.zip"
  source_code_hash = filebase64sha256("../dist/forwarder.zip")
  handler          = "forwarder"
  function_name    = "supportForwarder"
  role             = aws_iam_role.support_forward_lambda_role.arn
  runtime          = "go1.x"

  tracing_config {
    mode = "Active"
  }

  environment {
    variables = {
      MAIL_BUCKET     = aws_s3_bucket.mail_bucket.bucket
      MAIL_FROM       = local.support_email
      MAIL_TO         = aws_ses_email_identity.support_email.email
      SUBJECT_TO_SEND = "An email has been sent to ${local.support_email}"
      LOG_LEVEL       = "info"
    }
  }
}

resource "aws_lambda_permission" "forwarder_allow_ses_invoke" {
  action         = "lambda:InvokeFunction"
  function_name  = aws_lambda_function.support_forward_lambda.function_name
  principal      = "ses.amazonaws.com"
  source_account = data.aws_caller_identity.current.account_id
  statement_id   = "AllowSESInvokation"
}

resource "aws_ses_receipt_rule_set" "sabadoscodes_rules" {
  rule_set_name = "sabadoscodes.com"
}

resource "aws_ses_receipt_rule" "forward_support_email" {
  depends_on = [aws_lambda_permission.forwarder_allow_ses_invoke]

  name          = "forward_support_email"
  rule_set_name = aws_ses_receipt_rule_set.sabadoscodes_rules.rule_set_name
  recipients    = [local.support_email]
  enabled       = true
  scan_enabled  = true

  s3_action {
    position    = 1
    bucket_name = aws_s3_bucket.mail_bucket.bucket
  }

  lambda_action {
    position     = 2
    function_arn = aws_lambda_function.support_forward_lambda.arn
  }
}

resource "aws_ses_active_receipt_rule_set" "main" {
  rule_set_name = aws_ses_receipt_rule_set.sabadoscodes_rules.rule_set_name
}