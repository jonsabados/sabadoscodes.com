<template>
  <main role="main">
    <div v-if="articleFound">
      <h1>Spinning up sabadoscodes.com</h1>
      <hr />
      <h2>Planning</h2>
      <p>
        Although its just a silly personal playground sabadoscodes.com is making use of some of the same technologies
        that I might use in a professional setting, so writing about spinning it up seemed like as good an opportunity
        as any to document what all went into it. Like any real world project defining requirements was the starting
        point so there can be some direction to things, even if those requirements get to be super light and just for
        myself. So, lets knock those out:
      </p>
      <ol>
        <li>Implementing it should make me happy.</li>
        <li>It should be cheap. Like I should be able to fund this effort using spare change I find in the couch.</li>
        <li>
          It should give me the opportunity to write code, and not just front end stuff. JavaScript is fun and all, but
          I want to be able to put some Go, Java and whatever else strikes my fancy out there.
        </li>
        <li>
          Infrastructure should be codified. The days of logging into servers and executing commands, or logging into
          a console and poking a bunch of buttons are dead and should stay dead. Its not fun, its tedious and error
          prone. The only time one should be doing something like that is in a lab setting figuring out what one might
          later codify.
        </li>
      </ol>
      <p>So turns out setting up requirements is really easy when the only stakeholder is yourself :).</p>
      <p>
        Given those requirements the next step is to come up with an overall architecture.</p>
      <p>
        For #1 and #2 I say that it is going to be running in AWS, and its going to only be using pay per use style
        resources (no <a href="https://aws.amazon.com/ec2/" target="_blank">EC2</a> or
        <a href="https://aws.amazon.com/rds/" target="_blank">RDS</a> instances for me).
      </p>
      <p>I don't have to think too hard about #3 since I'm going to focus on just getting a website up and running
        and I know I can use <a href="https://aws.amazon.com/lambda/" target="_blank">Lambdas</a> behind either
        an <a href="https://aws.amazon.com/api-gateway/" target="_blank">API Gateway</a> or an
        <a href="https://docs.aws.amazon.com/elasticloadbalancing/latest/application/introduction.html">ALB</a>.
      </p>
      <p>The front is a no brainer - well just stuff html and assets in a
        <a href="https://aws.amazon.com/s3/" target="_blank">S3 bucket</a> and stick
        <a href="https://aws.amazon.com/cloudfront/" target="_blank">CloudFront</a> in front of it.
      </p>
      <p>
        Using <a href="https://www.terraform.io/" target="_blank">terraform</a> to define all of the AWS resources will
        take care of #4. I could probably use <a href="https://aws.amazon.com/cloudformation/" target="_blank">CloudFormation</a>
        instead, but I like that terraform can be used for all sorts of things outside of AWS as well, and writing
        custom providers for it is pretty easy in the event it doesn't already work with a thing.
      </p>
      <hr />
      <h2>Work Environment</h2>
      <p>
        Turns out I'm also working with a fresh install of Catalina on my mac, so the next step is to get that ready to
        go. I'm not going to dive into too much detail, but a few things are going to be needed:
      </p>
      <ul>
        <li>install homebrew - see <a href="https://brew.sh/" target="_blank">https://brew.sh/</a></li>
        <li>install pip: <code>sudo easy_install pip</code></li>
        <li>install the AWS CLI - see <a href="https://docs.aws.amazon.com/cli/latest/userguide/install-macos.html">https://docs.aws.amazon.com/cli/latest/userguide/install-macos.html</a></li>
        <li>configure AWS CLI credentials - go to IAM to create an access key and then run <code>aws configure</code></li>
        <li>install jq: <code>brew install jq</code></li>
        <li>install terraform, available <a href="https://www.terraform.io/downloads.html">here</a></li>
        <li>install node & npm, <a href="https://github.com/nvm-sh/nvm/blob/master/README.md" target="_blank">nvm</a> FTW</li>
        <li>install the vue CLI <code>npm install -g @vue/cli</code></li>
      </ul>
      <hr />
      <h2>Manual Stuff</h2>
      <p>
        Although all of the infrastructure will be codified there are a couple of things that need to be done manually,
        lets just knock those out before we start writing any code.
      </p>
      <p>
        First thing that is needed is going to be a domain name. Since everything else is being done in AWS and there
        will be a need to have records pointing to the CloudFront distribution
        <a href="https://aws.amazon.com/route53/" target="_blank">Route 53</a> is the way to go here.
      </p>
      <p>
        The site will also need a valid cert. This is also easy in AWS land, so off to <a href="https://aws.amazon.com/certificate-manager/" target="_blank">ACM</a>
        to request one. The cert should work for the top level domain, as well as <a href="https://www.sabadoscodes.com">www.sabadoscodes.com</a>
        so the request will be for <code>sabadoscodes.com</code> with an alternate name of <code>www.sabadoscodes.com</code>.
        ACM does DNS based validation to ensure your requesting certs for stuff you control so it asks you to create a
        specific CNAME. When your domain is in Route53 it is smart enough to know that this is a thing that can be done
        progmatically so it gives you a magic button to make the CNAME happen.
      </p>
      <p>
        Terraform state will also need some place to live that isn't just on my laptops hard drive. An S3 bucket seems
        like a pretty good candidate for that, so lets go ahead and create one by executing
        <code>aws s3api create-bucket --acl private --bucket bucket_name_redacted</code>.
        At this point I had a flash back to someone using a later version of terraform than everyone else, which
        modified the state so that it would only work with that version of terraform. The day was saved when we realized
        we had object versioning turned on in the state bucket allowing us to easily undo the damage, so lets turn
        on object versioning as well: <code>aws s3api put-bucket-versioning --bucket bucket_name_redacted --versioning-configuration Status=Enabled</code>.
      </p>
      <p>
        Finally a place will be needed to commit code. So off to GitHub to create
        <a href="https://github.com/jonsabados/sabadoscodes.com" target="_blank">this repo</a>.
      </p>
      <hr />
      <h2>Code!!!!</h2>
      <p>
        Now for the fun stuff, time to write some code! Or at least generate some...
      </p>
      <p>
        First up is creating the Vue.js app that will be the front end. After creating an empty project directory and
        running <code>git init</code> in it comes running <code>vue create sabadoscodes.com</code>. The default options
        for this are pretty terrible as they don't include any sort of unit testing (WTF????), and there are other goodies
        worth using so choosing "Manually select features" is the way to go when prompted, when filled out the screen
        looks like:
      </p>
      <samp>
        Vue CLI v4.1.2<br/>
        ? Please pick a preset: Manually select features<br/>
        ? Check the features needed for your project:<br/>
        ◉ Babel<br/>
        ◉ TypeScript<br/>
        ◯ Progressive Web App (PWA) Support<br/>
        ◉ Router<br/>
        ◉ Vuex<br/>
        ◉ CSS Pre-processors<br/>
        ◉ Linter / Formatter<br/>
        ❯◉ Unit Testing<br/>
        ◯ E2E Testing<br/>
      </samp>
      <p>
        Typescript is a must since typed languages make life sooooooo much better. Since this is more than just a single
        page app were definitely going to want a router. State management will also probably be a thing so Vuex is good
        to have. I just happen to know I'm going to suck in bootstrap which is in Sass so suck in CSS Pre-processors
        (plus doing plain old css sucks after having done stuff in less or sass). Every project needs a linter too so we
        want that. And finally if your not writing tests you shouldn't be writing code to begin with, so add in unit testing.
      </p>
      <p>
        Next prompt is <br />
        <samp>? Use class-style component syntax? (Y/n)</samp><br />
        the answer to that for me is <kbd>Y</kbd>,
        followed by<br />
        <samp>? Use Babel alongside TypeScript (required for modern mode, auto-detected polyfills, transpiling JSX)?</samp><br />
        That seems like a good idea, so another <kbd>Y</kbd>. The next question is<br />
        <samp>? Use history mode for router? (Requires proper server setup for index fallback in production) (Y/n)</samp><br/>
        this is a definite <kbd>Y</kbd> as it will make routes look like proper pages rather than stuff behind a # (this will add in some complications
        for reloads but we will address that later).
      </p>
      <p>
        Next prompt is to choose CSS pre-processor, either Sass/SCSS
        option should work just fine so I chose the node one. After that comes linting, followed by unit test solution.
        For unit testing with typescript Jest is the way to go as it allows for import mocking (others may have had more
        luck I wasted a few hours of my life trying to get proxyquire or rewire working with typescript and vue in the
        past).
      </p>
      <p>
        Next is infrastructure stuff, first step is to create an infrastructure directory and drop a <code>backend.tf</code>
        file in there with the following contents:
      </p>
      <pre class="code-block">
  terraform {
    backend "s3" {
      region = "us-east-1"
    }
  }
      </pre>
      <p>
        Those familiar with terraform may notice that the bucket name to use is not in this. This is a partial configuration,
        terraform will prompt for which bucket to use on the first run of <code>terraform-init</code>. This is handy
        since bucket names must be globally unique, so it allows for different installations of the same project without
        having to modify the code.
      </p>
      <p>
        Next up is the provider configuration. Since checking in credentials would be a terrible idea the AWS provider
        config should be done in a providers.tf file that looks something like:
      </p>
      <pre class="code-block">
variable "aws_region" {
  type    = string
  default = "us-east-1"
}

provider "aws" {
  version = "~> 2.8"
  region  = var.aws_region
}
      </pre>
      <p>I'm running in us-east-1 since it has all the goodies and is generally the least expensive.</p>
      <p>
        Next were going to need to create the bits needed for the UI. First up is going to be the S3 bucket everything is
        hosted in. The whole globally unique name thing complicates life a little bit though if we want to allow for
        different installations of the same codebase, so it needs to be parameterized somehow. Fortunately AWS has a good
        answer for that that will work with terraform in the form of the parameter store in
        <a href="https://aws.amazon.com/systems-manager/" target="_blank">AWS Systems Manager</a>. After defining an ssm
        parameter via the AWS console with a name of <code>sabadoscodes.uibucket</code> and a value of whatever bucket
        name desired we can suck that in as a data item in terraform, and then pull it into an S3 bucket. In terraform
        that might look something like:
      </p>
      <pre class="code-block">
data "aws_ssm_parameter" "ui_bucket_name" {
  name = "sabadoscodes.uibucket"
}

resource "aws_s3_bucket" "ui_bucket" {
  bucket = data.aws_ssm_parameter.ui_bucket_name.value
  acl    = "public-read"

  website {
    index_document = "index.html"
  }
}
      </pre>
      <p>
        A very important note here: this bucket has an ACL of <code>public-read</code> because its going to be backing a
        website and it is therefore intentionally exposed to the world. This should really only ever be
        used on buckets that are explicitly world readable. The internet is full of horror stories of private information
        being leaked via S3 buckets with improper access. Default to setting the ACL to private unless there is a damn
        good reason to do otherwise, and if a bucket is not private make sure all who might put stuff in it know so.
      </p>
      <p>
        At this point it would be possible to point folks directly to the bucket, but that is all sorts of tacky. So to
        get the bucket working as a backer for the sabadoscodes.com domain we first need to create a CloudFront
        distribution that has sabadoscodes.com setup as a name, as well as www.sabadoscodes.com (www isn't strictly
        necessary but  is a good thing to allow in case people type it in). Since it'd be good to allow for multiple
        installations to use the same code base we might as well make the domain name an SSM parameter as well and
        suck it in as a data element. The following additions to the project terraform take care of this (you
        will notice an import of the ACM cert that was created back in the manual steps):
      </p>
      <pre class="code-block">
data "aws_ssm_parameter" "domain_name" {
  name = "sabadoscodes.domain"
}

data "aws_acm_certificate" "website_cert" {
  domain = data.aws_ssm_parameter.domain_name.value
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
      </pre>
      <p>
        A couple of things are worth noting on this. First the <code>wait_for_deployment = false</code> is a general
        quality of life thing. CloudFront modifications will leave the distribution in a state of updating for quite a
        while and without that terraform will wait for that which can be painful. Next, since were on a budget and
        not expecting traffic from outside the US is <code>price_class = "PriceClass_100"</code> which will make
        international requests a bit more sluggish than they could be, but will save $.
      </p>
      <p>
        At this point running <code>terraform apply</code> will create the S3 bucket and CloudFront distribution. But
        typing <code>sabadoscodes.com</code> into a browser still isn't possible since DNS needs setup for that, so just
        a little more terraform:
      </p>
      <pre class="code-block">
data "aws_route53_zone" "ui_domain" {
  name = data.aws_ssm_parameter.domain_name.value
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
      </pre>
      <p>
        This defines a data element referencing the Route 53 zone for the domain we created back in the manual steps,
        and then creates two records in that zone for <code>sabadoscodes.com</code> and <code>www.sabadoscodes.com</code>.
        The <code>www</code> record is easy, just create a CNAME pointing to the CloudFront distributions name. The
        <code>sabadoscodes.com</code> entry is more difficult since you can't use CNAMEs for top level entries, they
        can only be A records - if whatever is backing sabados.com is dynamic that means that the A record for it would
        have to continuously update as that changed. Fortunately Route 53 makes this easy for us by allowing us to
        create an A record that is an alias to a supported AWS resource, CloudFront being one of those. After applying
        this terraform:
      </p>
      <pre>
<samp>jon@jonathans-mbp ~/Projects/sabadoscodes.com (master)</samp>
<samp>$</samp> <kbd>dig sabadoscodes.com</kbd>

<samp>; &lt;&lt;&gt;&gt; DiG 9.10.6 &lt;&lt;&gt;&gt; sabadoscodes.com
;; global options: +cmd
;; Got answer:
;; -&gt;&gt;HEADER&lt;&lt;- opcode: QUERY, status: NOERROR, id: 9546
;; flags: qr rd ra; QUERY: 1, ANSWER: 4, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 1452
;; QUESTION SECTION:
;sabadoscodes.com.&emsp;&emsp;IN&emsp;A

;; ANSWER SECTION:
sabadoscodes.com.&emsp;60&emsp;IN&emsp;A&emsp;143.204.29.68
sabadoscodes.com.&emsp;60&emsp;IN&emsp;A&emsp;143.204.29.122
sabadoscodes.com.&emsp;60&emsp;IN&emsp;A&emsp;143.204.29.125
sabadoscodes.com.&emsp;60&emsp;IN&emsp;A&emsp;143.204.29.38

;; Query time: 44 msec
;; SERVER: 192.168.86.1#53(192.168.86.1)
;; WHEN: Tue Jan 21 17:13:33 MST 2020
;; MSG SIZE  rcvd: 109</samp>

<samp>jon@jonathans-mbp ~/Projects/sabadoscodes.com (master)</samp>
<samp>$</samp> <kbd>dig sabadoscodes.com</kbd>

; &lt;&lt;&gt;&gt; DiG 9.10.6 &lt;&lt;&gt;&gt; www.sabadoscodes.com
;; global options: +cmd
;; Got answer:
;; -&gt;&gt;HEADER&lt;&lt;- opcode: QUERY, status: NOERROR, id: 39016
;; flags: qr rd ra; QUERY: 1, ANSWER: 5, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 1452
;; QUESTION SECTION:
;www.sabadoscodes.com.&emsp;&emsp;IN&emsp;A

;; ANSWER SECTION:
www.sabadoscodes.com.&emsp;37&emsp;IN&emsp;CNAME&emsp;d35g9vx3f3pk9d.cloudfront.net.
d35g9vx3f3pk9d.cloudfront.net. 37 IN&emsp;A&emsp;143.204.29.122
d35g9vx3f3pk9d.cloudfront.net. 37 IN&emsp;A&emsp;143.204.29.125
d35g9vx3f3pk9d.cloudfront.net. 37 IN&emsp;A&emsp;143.204.29.38
d35g9vx3f3pk9d.cloudfront.net. 37 IN&emsp;A&emsp;143.204.29.68

;; Query time: 17 msec
;; SERVER: 192.168.86.1#53(192.168.86.1)
;; WHEN: Tue Jan 21 17:13:38 MST 2020
;; MSG SIZE  rcvd: 156
      </pre>
      <p>
        Now there is somewhere to point a browser! But its just going to be a 404, so time to push content up.
        Theoretically one could just run <code>npm build</code> in the front end directory and then sync the files by
        hand. That has a couple problems though 1) I'm bad at remembering trivial details of the exact bucket name in
        use, and 2) I'm bad at remembering complex sets arguments for commands I don't use often. We can script around
        both of these problems easily though.
      </p>
      <p>
        First for the bucket name - we put this into SSM's parameter store, and all things AWS can be interacted with
        through the CLI so something like
        <code>UI_BUCKET=$(aws ssm get-parameter --output json --name sabadoscodes.uibucket | jq .Parameter.Value -r)</code>
        can be used - this uses the AWS CLI to fetch a parameter and spit it out in JSON format, then it is parsed by
        jq which is a handy little swiss army knife for JSON which spits out the raw value that can be stuffed into
        a variable.
      </p>
      <p>
        Next when it comes to actually getting files up to S3 we need to take HTTP cache control into account. Ideally
        we don't want repeat visitors to the site to have to re-request every asset every time, so it would be good
        to set things up so cache-control headers directing browsers (and CloudFront) to cache resources are sent. This
        is done in S3 using metadata, and CloudFront listens to cache-control headers and respects/forwards them so we
        can do that with an argument to various <code>aws s3</code> commands. Also, the vue.js builds produce unique
        names for everything asides from index.html and stuff in the <code>public</code> folder so pretty much everything
        but <code>index.html</code> can be set to cache for very long periods.
      </p>
      <p>
        The result of combining these to things is a script that looks something like:
      </p>
      <pre>
#!/bin/bash

UI_BUCKET=$(aws ssm get-parameter --output json --name sabadoscodes.uibucket | jq .Parameter.Value -r)

echo "Syncing dist/ to ${UI_BUCKET}"

# put everything in the bucket with a max age of 1 year
aws s3 sync ./dist "s3://$UI_BUCKET" --cache-control max-age=31536000 --delete --acl public-read
# then switch the max age on index.html to 60 seconds. Note, if stuff goes wrong with this or if something happens
# to hit cloudflare at the exact moment its cache expires its possible that cloudflare will cache it for a very long
# time, and we will need to invalidate the cache.
aws s3 cp "s3://$UI_BUCKET/index.html" "s3://$UI_BUCKET/index.html" --metadata-directive REPLACE  --cache-control max-age=60 --acl public-read
      </pre>
      <p>
        The first aws s3 sync puts everything into the bucket and tells S3 to send cache control headers to cache things
        for 1 year. After this is an S3 command to copy index.html onto itself replacing the cache control metadata.
        This process does have a noted issue where index.html can be cached for long periods if the perfect storm
        hits, but given an expected traffic load of basically 0 this is good enough for now.
      </p>
      <p>
        Now we can build the frontend by running <code>npm run build</code> from within the frontend directory,
        and push it to prod by running the bucket sync script from the same directory. At this point pointing a browser
        at <a href="https://sabadoscodes.com">https://sabadoscodes.com</a> works. There's still an issue though, since
        push state navigation is in place navigating to alternate routes will cause the url in the browser to update
        as if a request was made for a whole new page, but this is front end trickery and reloading the browser at this
        point will result in a 404. This can be fixed though by telling CloudFront to map 404's from the S3 origin
        into 200's that are mapped to <i>index.html</i>. A tiny addition to the
        <code>resource "aws_cloudfront_distribution" "ui_cdn"</code> definition:
      </p>
      <pre class="code-block">
  custom_error_response {
    error_code         = 404
    response_code      = 200
    response_page_path = "/index.html"
  }
      </pre>
      <p>
        and another <code>terraform apply</code> gets things up and running. At this point its pretty static and the
        application backend has yet to be touched, but its a good starting point that can be iterated on.
      </p>
      <p>
        UPDATE: The state of the repo when this article was written can be seen via
        <a href="https://github.com/jonsabados/sabadoscodes.com/releases/tag/first_article" target="_blank">this tag</a>.
      </p>
    </div>
    <div v-else>
      <h1>Article Not Found</h1>
      <p>No article having an id of {{ articleId }} is present in the system.</p>
    </div>
  </main>
</template>

<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'

@Component
export default class Article extends Vue {
  articleFound = false

  get articleId():string {
    return this.$route.params.id
  }

  mounted() {
    this.routeUpdated()
  }

  @Watch('$route')
  routeUpdated() {
    try {
      const articleId = parseInt(this.$route.params.id)
      this.articleFound = articleId === 0
    } catch {
      this.articleFound = false
    }
  }
}
</script>

<style lang="scss">
.code-block {
  color: #e83e8c;
}
</style>
