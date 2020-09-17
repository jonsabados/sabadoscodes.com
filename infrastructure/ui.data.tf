locals {
  workspace_domain_prefix = terraform.workspace == "default" ? "" : "${terraform.workspace}."
}

data "aws_ssm_parameter" "ui_bucket_name" {
  name = "sabadoscodes.uibucket"
}
