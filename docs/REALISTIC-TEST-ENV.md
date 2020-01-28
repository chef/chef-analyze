# Realistic Test Environment
Follow these steps to configure your local workstation to access a real
test environemnt.

1) Access the following [S3 bucket](https://s3.console.aws.amazon.com/s3/buckets/chef-workstation/environments/test/)
using your `chef-engineering` account.

2) Download the `chef-workstation.pem` from the bucket and store it at
`$HOME/.chef/chef-workstation.pem`

3) Create a `$HOME/.chef/credentials` config file (or if you already have
one, then just add the following profile) with:
  ```toml
  [default]
  client_name     = 'chef-workstation'
  client_key      = '/Users/YOUR-USERNAME/.chef/chef-workstation.pem'
  chef_server_url = 'https://54.212.196.10/organizations/gtms'
  ```
  _**NOTE: Update `YOUR-USERNAME` with your local username.**_

You should be able to run `chef analyze report nodes --ssl-no-verify`
to generate reports from the test environment.

### Unable to access test environment
If you are unable to access the above test environment, you might have to manually add your
IP Address to the static instance's security group. Log in to the Chef Engineering AWS Account
and navigate to the following [link](https://us-west-2.console.aws.amazon.com/ec2/home?region=us-west-2#SecurityGroups:groupId=sg-0dfb155aa6932036f;sort=groupId),
clic on the Security Group and select the **Inbound** tab, then clic **Edit** and **Add Rule**,
on the new rule choose the `Type=All Traffic` and the `Source=My IP`, finally clic **Save**.
