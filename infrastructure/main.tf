resource "aws_route53_record" "main_domain_txt_records" {
  name    = "${local.workspace_domain_prefix}${data.aws_ssm_parameter.domain_name.value}"
  type    = "TXT"
  zone_id = data.aws_route53_zone.main_domain.id
  records = [data.aws_ssm_parameter.google_console_txt_record.value]
  ttl     = 900
}