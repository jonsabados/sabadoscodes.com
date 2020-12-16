resource "aws_dynamodb_table" "article_store" {
  name         = "${local.workspace_prefix}ArticleStore"
  billing_mode = "PAY_PER_REQUEST"

  hash_key  = "Slug"
  range_key = "PublishDate"

  attribute {
    name = "Slug"
    type = "S"
  }

  attribute {
    name = "PublishDate"
    type = "N"
  }

  tags = {
    Workspace = terraform.workspace
  }
}