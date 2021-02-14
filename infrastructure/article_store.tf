resource "aws_dynamodb_table" "article_store" {
  name         = "${local.workspace_prefix}ArticleStore"
  billing_mode = "PAY_PER_REQUEST"

  hash_key  = "Slug"
  range_key = "SortKey"

  attribute {
    name = "Slug"
    type = "S"
  }

  attribute {
    name = "SortKey"
    type = "S"
  }

  attribute {
    name = "PublishDate"
    type = "N"
  }

  global_secondary_index {
    name = "PublishDateIndex"

    hash_key           = "PublishDate"
    range_key          = "Slug"
    projection_type    = "INCLUDE"
    non_key_attributes = ["Title"]
  }

  tags = {
    Workspace = terraform.workspace
  }
}