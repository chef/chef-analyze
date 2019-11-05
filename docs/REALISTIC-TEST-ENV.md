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
  [test]
  client_name     = 'chef-workstation'
  client_key      = '/Users/YOUR-USERNAME/.chef/chef-workstation.pem'
  chef_server_url = 'https://ip-10-0-22-232.us-west-2.compute.internal/organizations/gtms'
  ```
  _**NOTE: Update `YOUR-USERNAME` with your local username.**_

4) Add the following entry to your `/etc/hosts`
```
54.212.196.10 ip-10-0-22-232.us-west-2.compute.internal
```

You should be able to run `chef analyze report nodes --no-ssl-check --profile test`
to generate reports from the test environment.
