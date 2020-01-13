resource "aws_s3_bucket" "ui_bucket" {
  bucket = data.aws_ssm_parameter.ui_bucket_name.value
  acl    = "public-read"

  website {
    index_document = "index.html"
  }
}

resource "aws_cloudfront_origin_access_identity" "default" {}

resource "aws_cloudfront_distribution" "ui_cdn" {
  enabled             = true
  wait_for_deployment = false
  price_class         = "PriceClass_100"
  default_root_object = "index.html"
  aliases             = [
    data.aws_ssm_parameter.domain_name.value,
    "www.${data.aws_ssm_parameter.domain_name.value}"
  ]

  default_cache_behavior {
    allowed_methods        = [
      "HEAD",
      "GET"
    ]
    cached_methods         = [
      "HEAD",
      "GET"
    ]
    target_origin_id       = "ui_bucket"
    viewer_protocol_policy = "redirect-to-https"

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }
  }

  origin {
    origin_id   = "ui_bucket"
    domain_name = aws_s3_bucket.ui_bucket.bucket_regional_domain_name

    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.default.cloudfront_access_identity_path
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    ssl_support_method  = "sni-only"
    acm_certificate_arn = data.aws_acm_certificate.website_cert.arn
  }
}

resource "aws_route53_record" "default_domain_name" {
  name    = data.aws_ssm_parameter.domain_name.value
  type    = "A"
  zone_id = data.aws_route53_zone.ui_domain.zone_id

  alias {
    name                   = aws_cloudfront_distribution.ui_cdn.domain_name
    zone_id                = aws_cloudfront_distribution.ui_cdn.hosted_zone_id
    evaluate_target_health = true
  }
}

resource "aws_route53_record" "www_domain_name" {
  name    = "www"
  type    = "CNAME"
  zone_id = data.aws_route53_zone.ui_domain.zone_id
  records = [aws_cloudfront_distribution.ui_cdn.domain_name]
  ttl     = 60
}