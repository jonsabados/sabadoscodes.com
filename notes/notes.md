make shell sane, install autojump and git bash completion stuff (brew) and a useful PS1 - .bash_profile
```bash
# git command completion
[[ -r "/usr/local/etc/profile.d/bash_completion.sh" ]] && . "/usr/local/etc/profile.d/bash_completion.sh"
# autojump
[ -f /usr/local/etc/profile.d/autojump.sh ] && . /usr/local/etc/profile.d/autojump.sh

alias ls='ls -G'

export PS1='\e[0;32m\u\e[0;34m@\e[0;32m\h \e[0;32m\w\e[0;31m$(__git_ps1)\n\e[0;33m\$\e[m '

set -o vi
```

Manual stuff - register domain with route53

install pip (`sudo easy_install pip`)

install aws cli (https://docs.aws.amazon.com/cli/latest/userguide/install-macos.html)

configure aws creds - hop over to iam and create an access key

choose region and run aws configure

```bash
aws configure
AWS Access Key ID [None]: ***
AWS Secret Access Key [None]: ***
Default region name [None]: us-east-1
Default output format [None]: json
```

while were at it install jq - `brew install jq`, its going to be used in scripts

make sure aws works - `aws s3 ls`

create state bucket
```
aws s3api create-bucket --acl private --bucket $MY_STATE_BUCKET | jq
{
  "Location": "/bucket-name-will-be-here"
}
```

install terraform - https://www.terraform.io/downloads.html, grab zip, download and copy binary to /usr/local/bin

```bash
$ terraform -version
Terraform v0.12.19
```

initialize terraform - hop over to infrastructure and run `terraform init` - tell it which bucket and state file

turn on versioning in bucket
`aws s3api put-bucket-versioning --bucket $MY_STATE_BUCKET --versioning-configuration Status=Enabled`

drop in provider config, make region variable. Run `terraform init` again

hop over to github and create a repo, link to local

```
$ git remote add origin remote add origin git@github.com:jonsabados/sabadoscodes.com.git
$ git push --set-upstream origin master
The authenticity of host 'github.com (140.82.113.4)' can't be established.
RSA key fingerprint is SHA256:nThbg6kXUpJWGl7E1IGOCspRomTxdCARLviKw6E5SY8.
Are you sure you want to continue connecting (yes/no)? yes
Warning: Permanently added 'github.com,140.82.113.4' (RSA) to the list of known hosts.
git@github.com: Permission denied (publickey).
fatal: Could not read from remote repository.

Please make sure you have the correct access rights
and the repository exists.
```

setup ssh key in github - note something about `cat ~/.ssh/id_rsa.pub | pbcopy`
do the initial push again
```
$ git push --set-upstream origin master
Enumerating objects: 8, done.
Counting objects: 100% (8/8), done.
Delta compression using up to 8 threads
Compressing objects: 100% (6/6), done.
Writing objects: 100% (8/8), 1.41 KiB | 722.00 KiB/s, done.
Total 8 (delta 0), reused 0 (delta 0)
To github.com:jonsabados/sabadoscodes.com.git
 * [new branch]      master -> master
Branch 'master' set up to track remote branch 'master' from 'origin'.
```

Start spinning up issues so commits can have  thing to go against

Request cert in ACM, hop over to route53 and manually add the validation CNAME. Validation will be quick but take the
opportunity to go to the gym. Realize you had a typo in the domain name, redo cert request and use the handy "create
record in route53" buttons that pop up...

install node using node version manager - https://github.com/nvm-sh/nvm

install vue cli https://cli.vuejs.org/ - `npm install -g @vue/cli`

create the front end project
```
Vue CLI v4.1.2
? Please pick a preset: Manually select features
? Check the features needed for your project:
 ◉ Babel
 ◉ TypeScript
 ◯ Progressive Web App (PWA) Support
 ◉ Router
 ◉ Vuex
 ◉ CSS Pre-processors
 ◉ Linter / Formatter
❯◉ Unit Testing
 ◯ E2E Testing
```

create makefile for frontend project

infrastructure: add datasource for cert, put domain name and UI bucket name in SSM by hand and add data sources for those
so that they are not hard coded.
Make note of the price class for cloudfront

Data:
```
data "aws_ssm_parameter" "domain_name" {
  name = "sabadoscodes.domain"
}

data "aws_ssm_parameter" "ui_bucket_name" {
  name = "sabadoscodes.uibucket"
}

data "aws_acm_certificate" "website_cert" {
  domain = data.aws_ssm_parameter.domain_name.value
}
```

UI tf:
```
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
```

run a terraform apply (this one takes some time, make note of adding `wait_for_deployment = false`)...

next add stuff for dns
data:
```
data "aws_route53_zone" "ui_domain" {
  name = data.aws_ssm_parameter.domain_name.value
}
```

resources:
```
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
```

Run another terraform apply

Next create shell script to sync the bucket

```
#!/bin/bash

UI_BUCKET=$(aws ssm get-parameter --output json --name sabadoscodes.uibucket | jq .Parameter.Value -r)

echo "Syncing dist/ to ${UI_BUCKET}"

# put everything in the bucket with a max age of 1 year
aws s3 sync ./dist "s3://$UI_BUCKET" --cache-control max-age=31536000 --delete --acl public-read
# then switch the max age on index.html to 60 seconds. Note, if stuff goes wrong with this or if something happens
# to hit cloudflare at the exact moment its cache expires its possible that cloudflare will cache it for a very long
# time, and we will need to invalidate the cache.
aws s3 cp "s3://$UI_BUCKET/index.html" "s3://$UI_BUCKET/index.html" --metadata-directive REPLACE  --cache-control max-age=60 --acl public-read
```

run this bugger. Also talk about how this would be run by jenkins or something in the real world.

Just to avoid fun with cache invalidation and favicon fun go ahead and spin up gimp to create a crappy logo

next get push state history working

```
  custom_error_response {
    error_code         = 404
    response_code      = 200
    response_page_path = "/index.html"
  }
```