data "aws_ssm_parameter" "ui_bucket_name" {
  name = "sabadoscodes.uibucket"
}

data "aws_acm_certificate" "website_cert" {
  domain = data.aws_ssm_parameter.domain_name.value
}