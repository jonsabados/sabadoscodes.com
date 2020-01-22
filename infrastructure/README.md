# sabadoscodes.com infrastructure

sabadoscodes.com infrastructure is all maintained via terraform. It is known to work with terraform v0.12.19.

### State and AWS provider

State is stored in an S3 bucket, but the config is left as a partial config to allow for different buckets to
be used for different installations. The AWS provider also does not configure credentials, so you should have things
setup so the AWS cli works and hits your desired AWS env by default. When you run `terraform init` for the first time
it should prompt you for the state bucket name.

### Required manual configuration

There are a couple of things that must be setup by hand before `terraform apply` can be run. First, the domain name
to use must be registered in route53, and then an ACM certificate for that domain should be requested. This 
certificate should also have an alternate name with a www. prefix. Finally two parameters need to be entered in SSM,
they are:
 * `sabadoscodes.domain`: This should be a string and be set to the target domain name (no www prefix)
 * `sabadoscodes.uibucket`: This should be the name of the bucket to use for hosting front end static assets. This 
needs to be a globally unique name, but the bucket shouldn't exist as terraform will create it.

### Creating the infrastructure

After running `terraform init` once and then doing the required manual steps: `terraform apply`
