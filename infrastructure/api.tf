resource "aws_acm_certificate" "api_cert" {
  domain_name       = "api.${data.aws_ssm_parameter.domain_name.value}"
  validation_method = "DNS"
}

resource "aws_route53_record" "api_cert_verification_record" {
  name    = aws_acm_certificate.api_cert.domain_validation_options[0].resource_record_name
  type    = aws_acm_certificate.api_cert.domain_validation_options[0].resource_record_type
  zone_id = data.aws_route53_zone.main_domain.id
  records = [aws_acm_certificate.api_cert.domain_validation_options[0].resource_record_value]
  ttl     = 300
}

resource "aws_api_gateway_rest_api" "api" {
  name = "sabadoscodes.com"
}

resource "aws_api_gateway_domain_name" "api" {
  domain_name     = "api.${data.aws_ssm_parameter.domain_name.value}"
  certificate_arn = aws_acm_certificate.api_cert.arn
}

resource "aws_api_gateway_deployment" "main" {
  depends_on  = [aws_api_gateway_integration.cors_integration]
  rest_api_id = aws_api_gateway_rest_api.api.id
  stage_name  = "main"
}

resource "aws_api_gateway_stage" "main" {
  deployment_id = aws_api_gateway_deployment.main.id
  rest_api_id   = aws_api_gateway_rest_api.api.id
  stage_name    = "main"
}

resource "aws_route53_record" "api" {
  name    = "api"
  type    = "CNAME"
  zone_id = data.aws_route53_zone.main_domain.id
  records = [aws_api_gateway_domain_name.api.cloudfront_domain_name]
  ttl     = 300
}

resource "aws_api_gateway_base_path_mapping" "test" {
  api_id      = aws_api_gateway_rest_api.api.id
  stage_name  = aws_api_gateway_deployment.main.stage_name
  domain_name = aws_api_gateway_domain_name.api.domain_name
}