variable "lambda_name" {
  type = string
}

variable "lambda_policy" {
  type = string
}

variable "workspace_prefix" {
  type = string
}

variable "env_variables" {
  type = map
}

variable "timeout" {
  type    = string
  default = 3
}