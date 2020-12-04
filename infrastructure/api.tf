resource "aws_acm_certificate" "api_cert" {
  domain_name       = "${local.workspace_prefix}api.${data.aws_ssm_parameter.domain_name.value}"
  validation_method = "DNS"

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_route53_record" "api_cert_verification_record" {
  name    = aws_acm_certificate.api_cert.domain_validation_options[0].resource_record_name
  type    = aws_acm_certificate.api_cert.domain_validation_options[0].resource_record_type
  zone_id = data.aws_route53_zone.main_domain.id
  records = [aws_acm_certificate.api_cert.domain_validation_options[0].resource_record_value]
  ttl     = 300
}

resource "aws_api_gateway_rest_api" "api" {
  name = "${local.workspace_prefix}sabadoscodes.com"

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_api_gateway_domain_name" "api" {
  domain_name     = "${local.workspace_prefix}api.${data.aws_ssm_parameter.domain_name.value}"
  certificate_arn = aws_acm_certificate.api_cert.arn

  tags = {
    Workspace = terraform.workspace
  }
}

resource "aws_api_gateway_deployment" "main" {
  depends_on  = [aws_api_gateway_integration.cors_integration, aws_api_gateway_integration.self_integration, aws_api_gateway_integration.article_asset_list, aws_api_gateway_integration.article_asset_upload]
  rest_api_id = aws_api_gateway_rest_api.api.id
  stage_name  = "${local.workspace_prefix}main"

  variables = {
    "deployed_at": timestamp()
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "main" {
  deployment_id = aws_api_gateway_deployment.main.id
  rest_api_id   = aws_api_gateway_rest_api.api.id
  stage_name    = "${local.workspace_prefix}main"
}

resource "aws_route53_record" "api" {
  name    = "${local.workspace_prefix}api"
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