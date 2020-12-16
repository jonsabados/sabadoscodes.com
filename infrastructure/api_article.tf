resource "aws_api_gateway_resource" "article" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_rest_api.api.root_resource_id
  path_part   = "article"
}

resource "aws_api_gateway_resource" "article_slug" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.article.id
  path_part   = "slug"
}

resource "aws_api_gateway_resource" "article_by_slug" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.article_slug.id
  path_part   = "{slug}"
}

// TODO - lambda :)

resource "aws_api_gateway_method" "put_article_by_slug" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.article_by_slug.id
  http_method   = "PUT"
  authorization = "CUSTOM"
  authorizer_id = aws_api_gateway_authorizer.gateway_authorizer.id

  request_parameters = {
    "method.request.path.slug" = true
  }
}