data "aws_ssm_parameter" "domain_name" {
  name = "sabadoscodes.domain"
}

data "aws_ssm_parameter" "ui_bucket_name" {
  name = "sabadoscodes.uibucket"
}

data "aws_acm_certificate" "website_cert" {
  domain = data.aws_ssm_parameter.domain_name.value
}

data "aws_route53_zone" "ui_domain" {
  name = data.aws_ssm_parameter.domain_name.value
}