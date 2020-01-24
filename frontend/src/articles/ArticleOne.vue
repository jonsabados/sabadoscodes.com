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
      with <a href="https://aws.amazon.com/cognito/" target="_blank">Cognito</a>. After a little bit of research I ended
      on Google's developer console to setup the oauth consent screen, and that wants an email address to show
      to users. I'd rather not use my personal email address, so it means receiving email at an address like
      support@sabados.com needs to happen first. This should be easy enough using
      <a href="https://aws.amazon.com/ses/" target="_blank">SES</a>, so lets dive into that.
    </p>
    <hr />
    <h2>Planning</h2>
    <p>
      Odds are at some point down the line multiple @sabadoscodes.com email addresses will be needed, but that's the
      future, so for now the plan is to focus on setting up a single email address and getting that to forward while not
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
      possible to create the verification record in terraform, no console button pushing needed - score! After a little
      bit of refactoring things out of ui that would be shared came the creation of a
      <a href="https://github.com/jonsabados/sabadoscodes.com/commit/7bff9e527b994275638291dcdedd31bf8bf36566#diff-cb1235822918e9bb58295736c1d68490" target="_blank">mail.tf</a>
      file to define all things mail related. The first bit is the creation of the SES domain itself:
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
      there. Then DKIM is turned on for the domain, and the required DNS entries for that to function are added. The
      terraform folks were good enough to provide an
      <a href="https://www.terraform.io/docs/providers/aws/r/ses_domain_dkim.html" target="_blank">example of this</a>
      so not much thought needed to go into that. The DNS records do make use of count magic to create the 3 required
      entries which is a pretty neat trick. And finally the MX record to tell SMTP relays how to get email to
      sabadoscodes.com addresses.
    </p>
    <p>
      Next thing needed is to enable sending of email. There are two ways to go about this. The first would be to put in a
      ticket to take the SES domain out of sandbox mode which requires filling out questions about how to deal with
      unsubscribe requests and the like, which is out of scope since this is internal mail only. A
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
      I stuffed the data element in a
      <a href="https://github.com/jonsabados/sabadoscodes.com/commit/7bff9e527b994275638291dcdedd31bf8bf36566#diff-c70d44b0b0b1f4e2068871450a5f69a3" target="_blank">mail.data.tf</a>
      file to keep the data grouped with mail definitions but separate. Once <code>terraform apply</code> is run this
      creates an entry in the email addresses section of SES in a pending verification state. An email will be sent to
      the specified email address with a verification link that can be followed to complete verification.
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
      the <code>data.aws_iam_policy_document</code> approach to be cleaner and easier to read. And then the bucket itself.
    </p>
    <h3>The Lambda</h3>
    <p>
      At this point its finally time to write some actual code and get it building!!! The result is some
      <code>Makefile</code> refactoring and a new <a href="https://github.com/jonsabados/sabadoscodes.com/tree/support_email_address/backend/src/" target="_blank">backend directory</a>
      containing the lambda itself.
    </p>
    <h4>S3 operations</h4>
    <p>
      The first step with this is to create some utilities around talking to S3. The S3 client isn't terrible to use,
      but mocking it suuuuucks so its not fun for testing, and having something that looks like
      <code>GetObject(bucket, item string) (io.ReadCloser, error)</code> would be nice, as well as maybe something to
      delete objects, so lets knock that out in our own little S3 package:
    </p>
    <pre class="code-block">
type ObjectFetcher func(ctx context.Context, bucket, object string) (io.ReadCloser, error)

func NewObjectFetcher(client *s3.S3) ObjectFetcher {
    return func(ctx context.Context, bucket, object string) (io.ReadCloser, error) {
        zerolog.Ctx(ctx).Debug().Str("bucket", bucket).Str("key", object).Msg("fetching object")
        res, err := client.GetObjectWithContext(ctx, &s3.GetObjectInput{
            Bucket: aws.String(bucket),
            Key:    aws.String(object),
        })
        if err != nil {
            return nil, errors.WithStack(err)
        }
        return res.Body, nil
    }
}

type ObjectRemover func(ctx context.Context, bucket, object string) error

func NewObjectRemover(client *s3.S3) ObjectRemover {
    return func(ctx context.Context, bucket, object string) error {
        zerolog.Ctx(ctx).Info().Str("bucket", bucket).Str("key", object).Msg("removing object")
        _, err := client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
            Bucket: aws.String(bucket),
            Key:    aws.String(object),
        })
        return err
    }
}

func RawClient(sess *session.Session) *s3.S3 {
    ret := s3.New(sess)
    xray.AWS(ret.Client)
    return ret
}
    </pre>
    <p>
      This may look a little foreign to folks without any sort of FP background, but this is a pattern making use of
      higher order functions to provide dependencies to behavior that avoids having to use structs. I'm not going to
      dive into it too much since its actually a subject I want to write in depth about once I've got articles persisting
      in a store rather than just hard coding the things. But, the net of it is these two constructor functions will
      provide functions that take a context, bucket name and object key and then do things in S3 with them. They will
      also be really easy to mock in unit tests. The last guy, <code>RawClient</code> is just a handy utility function
      to kick back an S3 client that will automagically put stuff in X-Ray which will be talked about a little bit more
      later. Since these two things basically just delegate to S3 clients which are a royal pain to mock tests probably
      aren't worth the squeeze (I say this after loosing 30 minutes of my life to passing the bucket arg as both the
      bucket and key in the object fetcher...)
    </p>
    <h4>Working with email</h4>
    <p>
      Next come the bits for sending email and a thing to forward an email stored in S3:
    </p>
    <pre class="code-block">
type Attachment struct {
    Name     string
    MimeType string
    Body     io.Reader
}

type Sender func(ctx context.Context, from, to, subject, htmlBody, textBody string, attachments ...Attachment) error

func NewSender(sesClient *ses.SES) Sender {
    return func(ctx context.Context, from, to, subject, htmlBody, textBody string, attachments ...Attachment) error {
        e := &email.Email{
            To:          []string{to},
            From:        from,
            Subject:     subject,
            Text:        []byte(textBody),
            HTML:        []byte(htmlBody),
        }
        for _, a := range attachments {
            _, err := e.Attach(a.Body, a.Name, a.MimeType)
            if err != nil {
                return errors.WithStack(err)
            }
        }

        payload, err := e.Bytes()
        if err != nil {
            return errors.WithStack(err)
        }

        zerolog.Ctx(ctx).Info().Str("source", from).Str("to", to).Msg("sending email")
        _, err = sesClient.SendRawEmailWithContext(ctx, &ses.SendRawEmailInput{
            Source: aws.String(from),
            Destinations: []*string{
                aws.String(to),
            },
            RawMessage: &ses.RawMessage{
                Data: payload,
            },
        })
        if err != nil {
            return errors.WithStack(err)
        }
        return nil
    }
}

type Forwarder func(ctx context.Context, messageID, subjectToSend, sendFrom, forwardTo string) error

func NewForwarder(mailBucket string, fetchObject s3.ObjectFetcher, sendEmail Sender, removeObject s3.ObjectRemover) Forwarder {
    return func(ctx context.Context, messageID, subjectToSend, sendFrom, forwardTo string) error {
        originalEmail, err := fetchObject(ctx, mailBucket, messageID)
        if err != nil {
            return errors.WithStack(err)
        }
        defer func() {
            err := originalEmail.Close()
            if err != nil {
                zerolog.Ctx(ctx).Warn().Str("error", fmt.Sprintf("%+v", err)).Msg("error closing stream")
            }
        }()

        htmlBody := "&lt;p&gt;See attached email&lt;/p&gt;"
        textBody := "See attached email"
        err = sendEmail(ctx, sendFrom, forwardTo, subjectToSend, htmlBody, textBody, Attachment{
            Name:     fmt.Sprintf("%s.eml", messageID),
            MimeType: "message/rfc822",
            Body:     originalEmail,
        })
        if err != nil {
            return errors.WithStack(err)
        }

        err = removeObject(ctx, mailBucket, messageID)
        if err != nil {
            return errors.WithStack(err)
        }
        return nil
    }
}

func NewRawClient(sess *session.Session) *ses.SES {
    ret := ses.New(sess)
    xray.AWS(ret.Client)
    return ret
}
    </pre>
    <p>
      Sender can be used to just send emails - since this implementation just attaches the original email to a whole
      new email the SES <a href="https://docs.aws.amazon.com/ses/latest/APIReference/API_SendRawEmail.html" target="_blank">SendRawEmail</a>
      operation needs to be used, and <a href="https://github.com/jordan-wright/email" target="_blank">this package</a>
      is used to build the email since doing string concatenation to build an email by hand would be entirely dreadful
      (I am very deeply disturbed by how many examples of this I saw for go). Forwarder is actually the main guts of
      what we need to do, and thanks to DI will be pretty trivial to
      <a href="https://github.com/jonsabados/sabadoscodes.com/blob/support_email_address/backend/src/go/mail/mail_test.go" target="_blank">test</a>.
      There is also a utility function to kick back an SES client that will do X-Ray magic - this goes hand in hand with making sure
      all aws client calls use the <code>WithContext</code> versions (even without X-Ray that should be the version
      used in lambda's as the context will have a deadline that the clients listen to and so on)
    </p>
    <h4>Tying it together</h4>
    <p>
      At this point all that is left is packing the newly created logic up into a lambda, so from within a <code>main</code>
      package:
    </p>
    <pre class="code-block">
func NewRequestHandler(logger zerolog.Logger, forwardEmail mail.Forwarder, subjectToSend, sendFrom, sendTo string) func(ctx context.Context, events events.SimpleEmailEvent) error {
    return func(ctx context.Context, events events.SimpleEmailEvent) error {
        ctx = logger.WithContext(ctx)
        for _, e := range events.Records {
            logger.Info().Str("id", e.SES.Mail.MessageID).Msg("processing item")
            err := forwardEmail(ctx, e.SES.Mail.MessageID, subjectToSend, sendFrom, sendTo)
            if err != nil {
                logger.Error().Str("error", fmt.Sprintf("%+v", err)).Msg("error sending email")
                return errors.WithStack(err)
            }
        }
        return nil
    }
}

func main() {
    err := xray.Configure(xray.Config{
        LogLevel: "warn",
    })
    if err != nil {
        panic(err)
    }

    sess, err := session.NewSession(&aws.Config{})
    if err != nil {
        panic(err)
    }

    logLevel, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
    if err != nil {
        panic(err)
    }
    logger := zerolog.New(os.Stdout).Level(logLevel)

    s3Client := s3.RawClient(sess)
    getObject := s3.NewObjectFetcher(s3Client)
    deleteObject := s3.NewObjectRemover(s3Client)

    sesClient := mail.NewRawClient(sess)
    mailSender := mail.NewSender(sesClient)

    forwarder := mail.NewForwarder(os.Getenv("MAIL_BUCKET"), getObject, mailSender, deleteObject)
    lambda.Start(NewRequestHandler(logger, forwarder, os.Getenv("SUBJECT_TO_SEND"), os.Getenv("MAIL_FROM"), os.Getenv("MAIL_TO")))
}
    </pre>
    <p>
      The main function itself does all of our DI wiring and then calls <code>lambda.Start</code> with a handler
      that has dependencies provided in the same manner that other bits do. The <code>main</code> function is not
      really testable, but that is OK since its just doing wiring. The handler certainly is
      <a href="https://github.com/jonsabados/sabadoscodes.com/blob/support_email_address/backend/src/go/mail/forwarder/main_test.go" target="_blank">testable though</a>.
      Finally it needs to be built into a deployable lambda, so a couple of targets in the top level makefile do that:
    </p>
    <pre class="code-block">
dist/forwarder: dist/ $(shell find backend/src/go)
    cd backend/src/go && GOOS=linux go build -o ../../../dist/forwarder github.com/jonsabados/sabadoscodes.com/mail/forwarder

dist/forwarderLambda.zip: dist/forwarder
    cd dist && zip forwarder.zip forwarder
    </pre>
    <p>
      The terraform to create the lambda can then reference these files.
    </p>
    <h4>Lambda Terraform</h4>
    <p>
      Now for the terraform to create the lambda:
    </p>
    <pre class="code-block">
resource "aws_cloudwatch_log_group" "support_forward_logs" {
  name              = "/aws/lambda/${aws_lambda_function.support_forward_lambda.function_name}"
  retention_in_days = 7
}

data "aws_iam_policy_document" "support_forward_lambda_policy" {
  statement {
    sid       = "AllowLogging"
    effect    = "Allow"
    actions   = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = [
      "arn:aws:logs:*:*:*"
    ]
  }

  statement {
    sid       = "AllowXRayWrite"
    effect    = "Allow"
    actions   = [
      "xray:PutTraceSegments",
      "xray:PutTelemetryRecords",
      "xray:GetSamplingRules",
      "xray:GetSamplingTargets",
      "xray:GetSamplingStatisticSummaries"
    ]
    resources = ["*"]
  }

  statement {
    sid       = "AllowListBucket"
    effect    = "Allow"
    actions   = [
      "s3:ListBucket"
    ]
    resources = [
      aws_s3_bucket.mail_bucket.arn
    ]
  }

  statement {
    sid       = "AllowMailBucketReadWrite"
    effect    = "Allow"
    actions   = [
      "s3:GetObject",
      "s3:DeleteObject"
    ]
    resources = [
      "${aws_s3_bucket.mail_bucket.arn}/*"
    ]
  }

  statement {
    sid       = "AllowSESSendRawEmail"
    effect    = "Allow"
    actions   = [
      "ses:SendRawEmail"
    ]
    resources = [
      "*"
    ]
  }
}

resource "aws_iam_role" "support_forward_lambda_role" {
  name               = "supportForwarderLambdaRole"
  assume_role_policy = data.aws_iam_policy_document.assume_lambda_role_policy.json
}

resource "aws_iam_role_policy" "support_forward_lambda_role_policy" {
  role   = aws_iam_role.support_forward_lambda_role.name
  policy = data.aws_iam_policy_document.support_forward_lambda_policy.json
}

resource "aws_lambda_function" "support_forward_lambda" {
  filename         = "../dist/forwarder.zip"
  source_code_hash = filebase64sha256("../dist/forwarder.zip")
  handler          = "forwarder"
  function_name    = "supportForwarder"
  role             = aws_iam_role.support_forward_lambda_role.arn
  runtime          = "go1.x"

  tracing_config {
    mode = "Active"
  }

  environment {
    variables = {
      MAIL_BUCKET     = aws_s3_bucket.mail_bucket.bucket
      MAIL_FROM       = local.support_email
      MAIL_TO         = aws_ses_email_identity.support_email.email
      SUBJECT_TO_SEND = "An email has been sent to ${local.support_email}"
      LOG_LEVEL       = "info"
    }
  }
}

resource "aws_lambda_permission" "forwarder_allow_ses_invoke" {
  action         = "lambda:InvokeFunction"
  function_name  = aws_lambda_function.support_forward_lambda.function_name
  principal      = "ses.amazonaws.com"
  source_account = data.aws_caller_identity.current.account_id
  statement_id   = "AllowSESInvokation"
}
    </pre>
    <p>
      There is actually a decent amount to digest in this. The first resource,
      <code>aws_cloudwatch_log_group.support_forward_logs</code> creates a log group in CloudWatch where the lambda
      will log too. Without this the lambda will actually create the log group set to retain things forever, and that
      isn't appealing from a cost perspective so its created with a short retention period.
    </p>
    <p>
      Next up is <code>data.aws_iam_policy_document.support_forward_lambda_policy</code>. This is a data element used
      to define the policy that will be later assigned to the lambda, and it gives the lambda permission to log (because
      lambdas that don't are basically impossible to trouble shoot), gives the lambda permission to record APM metrics
      in X-Ray, allows the lambda permission to list the bucket its reading from and then gives read/delete permissions
      to all objects in that bucket. Finally it gives the lambda permission to send raw emails using SES.
    </p>
    <p>
      After that is <code>aws_iam_role.support_forward_lambda_role</code> which creates the role, and gives it an
      assume role policy that lets lambdas assume it. Since this is going to be a repeated policy that was created
      in a top level <a href="https://github.com/jonsabados/sabadoscodes.com/blob/support_email_address/infrastructure/main.data.tf" target="_blank">main.data.tf</a>
      file. The role by itself doesn't do too much though, and needs the policy created earlier associated with it so
      that is what <code>aws_iam_role_policy.support_forward_lambda_role_policy</code> does.
    </p>
    <p>
      Then comes the lambda itself in <code>aws_lambda_function.support_forward_lambda</code>. It references the local
      builds so deploying new versions is just a matter of doing a <code>make</code> and
      <code>cd infrastructure && terraform apply</code>. Finally comes <code>aws_lambda_permission.forwarder_allow_ses_invoke</code>
      which gives SES permission to invoke the lambda.
    </p>
    <hr />
    <h2>SES Wiring</h2>
    <p>
      Now all the pieces to forward email are in place so a rule set for SES can be created that makes the magic happen.
      This is some pretty straight forward terraform:
    </p>
    <pre class="code-block">
resource "aws_ses_receipt_rule_set" "sabadoscodes_rules" {
  rule_set_name = "sabadoscodes.com"
}

resource "aws_ses_receipt_rule" "forward_support_email" {
  depends_on = [aws_lambda_permission.forwarder_allow_ses_invoke]

  name          = "forward_support_email"
  rule_set_name = aws_ses_receipt_rule_set.sabadoscodes_rules.rule_set_name
  recipients    = [local.support_email]
  enabled       = true
  scan_enabled  = true

  s3_action {
    position    = 1
    bucket_name = aws_s3_bucket.mail_bucket.bucket
  }

  lambda_action {
    position     = 2
    function_arn = aws_lambda_function.support_forward_lambda.arn
  }
}

resource "aws_ses_active_receipt_rule_set" "main" {
  rule_set_name = aws_ses_receipt_rule_set.sabadoscodes_rules.rule_set_name
}
    </pre>
    <p>
      This creates a rule set, adds one rule for the support@sabadoscodes.com email address and then a final
      element makes that ruleset the active one. That bit could get weird in an account with multiple applications using
      SES, but not an issue in my scenario.
    </p>
    <hr />
    <h2>X-Ray</h2>
    <p>
      Once this is all tied together lambda invocations will record APM metrics in X-Ray, which is super helpful for
      diagnosing performance issues and the like. After sending an email to the support email address the X-Ray
      service map will have something like this in it:
    </p>
    <img src="./xray-service-map.jpg" alt="service map screenshot" class="container-fluid"/>
    <p>
      The service map is a good way to see how everything in your infrastructure is interacting.
    </p>
    <p>
      Even better though are traces, which look something like:
    </p>
    <img src="./xray-trace.jpg" alt="x-ray trace screenshot" class="container-fluid"/>
    <p>
      This breakdown of time spent doing what is super useful to have, and it comes nearly for free. Once you set
      a lambda's tracing mode to active, do the <code>xray.AWS</code> call on all your AWS api clients, and then
      remember to always use the <code>WithContext</code> version of things it just works (also, make sure you suck
      in <code>v1.0.0-rc.1</code> of <code>github.com/aws/aws-xray-sdk-go</code> or it'll just grenade when running in
      a lambda due to a fixed bug). It is also possible to augment HTTP clients with X-Ray magic which I would highly
      recommend (if I ever implement a service to service call with sabadoscodes.com I'll talk about doing that then).
    </p>
  </div>
</template>
