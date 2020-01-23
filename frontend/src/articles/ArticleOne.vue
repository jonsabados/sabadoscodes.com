<template>
  <div class="article">
    <h1>Setting up a support email address</h1>
    <hr />
    <h2>The need to receive email</h2>
    <p>
      At this point sabadoscodes.com has the start of a front end with a single hard coded article about the creation
      of sabadoscodes.com so the next goal is to get some sort of authentication and authorization in place as its a
      blocker for any sort of API and front end for publishing articles. This is one area that I'm not super keen on
      re-inventing the wheel, so it makes sense to use something like
      <a href="https://developers.google.com/identity/sign-in/web/sign-in" target="_blank">google sign in</a> combined
      with <a href="https://aws.amazon.com/cognito/" target="_blank">Cognito</a>. After a little bit of research and
      heading over to Google's developer console to setup the oauth consent screen which wants an email address to show
      to users - I'd rather not use my personal email address, so it means receiving email at an address like
      support@sabados.com needs to happen first. This should be easy enough using
      <a href="https://aws.amazon.com/ses/" target="_blank">SES</a>, so lets dive into that.
    </p>
    <hr />
    <h2>Planning</h2>
    <p>
      Odds are at some point down the line multiple @sabadoscodes.com email addresses will be needed, but that's the
      future so for now the plan is to focus on setting up a single email address and getting that to forward while not
      boxing things into a corner that makes expanding on that difficult. One option would be to just use
      <a href="https://aws.amazon.com/workmail/" target="_blank">WorkMail</a> but that has a monthly per-user cost, and
      one of the main goals is to just use pay-per-use services so that is out. SES can be set up with receiving rules
      that target a variety of things including S3 buckets and lambdas though, so that seems like a good option. Bonus,
      there is even an example for forwarding emails <a href="https://aws.amazon.com/blogs/messaging-and-targeting/forward-incoming-email-to-an-external-destination/" target="_blank">here</a>!
      The example lambda is in python and all the instructions are manual, and since I'm itching to write some Go and
      want all the infrastructure to be codified in terraform it can't be used directly, but it makes for a good
      template, so the plan will just be to spin up an SES domain and then repeat what's done in python in Go (this will
      also leave things in a good place for further extension with a generic email forwarder that might use some sort
      of data store to map recipients). Easy...
    </p>
    <hr />
    <h2>SES domain</h2>
    <p>
      The first step is to get an SES domain established. This involves creating the domain and then adding DNS records
      to verify ownership. Doing this by hand wouldn't be terrible since it's a one time thing and the domain could be
      sucked into terraform as a data element, but it turns out that you can create a domain with terraform, and that
      adds the verification token needed as an export of the resource and since Route53 is being used for DNS its also
      possible to create the verification record in terraform, no console button pushing needed - score! Previous work
      had created data elements pulling from SES for the domain name so there was a little bit of refactoring to move
      that stuff to a <code>main.data.tf</code> which you can see in this commit <b>TODO - link when it exists</b>.
      After this refactoring things are setup to create a <code>mail.tf (link)</code> file to define all things mail
      related and then create the domain itself:
    </p>
    <pre class="code-block">
resource "aws_ses_domain_identity" "ses_domain" {
  domain = data.aws_ssm_parameter.domain_name.value
}

resource "aws_route53_record" "ses_domain_verification_record" {
  zone_id = data.aws_route53_zone.main_domain.zone_id
  name    = "_amazonses.${data.aws_ssm_parameter.domain_name.value}"
  type    = "TXT"
  ttl     = "600"
  records = [aws_ses_domain_identity.ses_domain.verification_token]
}

resource "aws_ses_domain_dkim" "ses_domain_dkim" {
  domain = aws_ses_domain_identity.ses_domain.domain
}

resource "aws_route53_record" "dkim_dns_records" {
  count   = 3
  zone_id = data.aws_route53_zone.main_domain.zone_id
  name    = "${element(aws_ses_domain_dkim.ses_domain_dkim.dkim_tokens, count.index)}._domainkey.${data.aws_ssm_parameter.domain_name.value}"
  type    = "CNAME"
  ttl     = "600"
  records = ["${element(aws_ses_domain_dkim.ses_domain_dkim.dkim_tokens, count.index)}.dkim.amazonses.com"]
}

resource "aws_route53_record" "mx_record" {
  zone_id = data.aws_route53_zone.main_domain.zone_id
  name    = data.aws_ssm_parameter.domain_name.value
  type    = "MX"
  ttl     = "600"
  records = ["10 inbound-smtp.us-east-1.amazonaws.com"]
}
    </pre>
    <p>
      The first resource defines the domain itself, which will be created in a pending verification state. The next
      resource adds the verification TXT record needed to complete the verification so nothing manual will be required
      there. Then DKIM is turned on for the domain, and the required DNS entries for that to function are added (the
      terraform folks were good enough to provide an
      <a href="https://www.terraform.io/docs/providers/aws/r/ses_domain_dkim.html" target="_blank">example of this</a>
      so not much thought needed to go into that. The DNS records do make use of count magic to create the 3 required
      entries which is a pretty neat trick. And finally the MX record to tell SMTP relays how to get email to
      sabadoscodes.com addresses.
    </p>
    <p>
      Next we need to enable sending of email. There are two ways to go about this, the first would be to put in a
      ticket to take the SES domain out of sandbox mode which requires filling out questions about how to deal with
      unsubscribe requests and the like, which is out of scope since this is internal mail only. Theoretically the
      ticket could be opened indicating this, but its likely additional mail will be sent down the line to external
      users and unsubscribe type functionality will be needed then so that probably isn't the best route. The other
      option would be to create a verified email address that can be sent to in sandbox mode, so that sounds like a good
      plan. Because I don't want my personal email hard coded in the public repository, and it would be kinda neat to
      keep multiple instances of the site with different configs possible it makes sense to put a record in the SSM
      parameter store for this which is done by hand. Then the following elements are needed in terraform:
    </p>
    <pre class="code-block">
data "aws_ssm_parameter" "support_email" {
  name = "sabadoscodes.email.support"
}

resource "aws_ses_email_identity" "support_email" {
  email = data.aws_ssm_parameter.support_email.value
}
    </pre>
    <p>
      I stuffed the data element in a <code>mail.data.tf (link)</code> file to keep the data grouped with mail definitions
      but separate. Once <code>terraform apply</code> is run this creates an entry in the email addresses section of
      SES in a pending verification state. An email will be sent to the specified email address with a verification
      link that can be followed to complete verification.
    </p>
    <hr />
    <h2>Forwarding Email</h2>
    <p>
      Because SES receiving rules going to lambda's don't include the email itself this will involve creating a rule
      that drops messages in an S3 bucket, and then triggers a lambda that will do the needful with that record. So
      first let's create the bucket (this is based on the AWS provided example referenced earlier):
    </p>
    <pre class="code-block">
locals {
  mail_bucket_name = "mail.${data.aws_ssm_parameter.domain_name.value}"
}

data "aws_iam_policy_document" "mail_bucket_policy" {
  statement {
    sid       = "AllowSESPuts"
    effect    = "Allow"
    principals {
      identifiers = ["ses.amazonaws.com"]
      type        = "Service"
    }
    actions   = ["s3:PutObject"]
    resources = ["arn:aws:s3:::${local.mail_bucket_name}/*"]
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "aws:Referer"
    }
  }
}

resource "aws_s3_bucket" "mail_bucket" {
  bucket = local.mail_bucket_name
  policy = data.aws_iam_policy_document.mail_bucket_policy.json
  acl    = "private"
}
    </pre>
    <p>
      First is a local variable definition for the bucket name since its going to be referenced repeatedly. Then comes
      a data element which defines the policy were going to stick on the bucket. In theory it could be done by
      stuffing a raw json string in for the policy attribute of <code>aws_s3_bucket.mail_bucket</code>, but I find
      the <code>aws_iam_policy_document</code> approach to be cleaner and easier to read. And then the bucket itself.
    </p>
    <p>
      At this point its finally time to write some actual code and get it building!!! The result is some
      <code>Makefile</code> refactoring and a new <code>backend (link)</code> directory containing the lambda itself
    </p>
  </div>
</template>
