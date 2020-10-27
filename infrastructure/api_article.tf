resource "aws_api_gateway_resource" "article" {
  parent_id   = aws_api_gateway_rest_api.api.root_resource_id
  path_part   = "article"
  rest_api_id = aws_api_gateway_rest_api.api.id
}