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

while were at it install jq - `brew install jq`

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

hop over to github and create a repo