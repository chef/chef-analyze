# Design
This document describes the current and future shape of this tool, it will
be a place for us, at Chef, to discuss about the design and usability, the
flow of commands that we expect user to follow plus, we will start the
documentation of this tool in an early stage.

## Help Documentation

### Main help
```bash
$ chef-analyze help
Analyze your Chef inventory by generating reports to understand the effort
to upgrade to the latest version of the Chef tools, automatically fixing
violations and/or deprecations, and generating Effortless packages.

Usage:
  chef-analyze [command]

Available Commands:
  help        Help about any command
  report      Generate reports about your Chef inventory

Flags:
  -s, --chef_server_url string   Chef Infra Server URL
  -k, --client_key string        Chef Infra Server API client key
  -n, --client_name string       Chef Infra Server API client username
  -c, --credentials string       Chef credentials file (default $HOME/.chef/credentials)
  -h, --help                     help for chef-analyze
  -o, --no-ssl-check             Disable SSL certificate verification
  -p, --profile string           Chef Infra Server URL (default "default")

Use "chef-analyze [command] --help" for more information about a command.
$
```
### `report` sub-command help
```
$ chef-analyze report --help
Generate reports about your Chef inventory

Usage:
  chef-analyze report [command]

Available Commands:
  cookbooks   Generates a cookbook oriented report
  nodes       Generates a nodes oriented report

Flags:
  -h, --help   help for report

Global Flags:
  -s, --chef_server_url string   Chef Infra Server URL
  -k, --client_key string        Chef Infra Server API client key
  -n, --client_name string       Chef Infra Server API client username
  -c, --credentials string       Chef credentials file (default $HOME/.chef/credentials)
  -o, --no-ssl-check             Disable SSL certificate verification
  -p, --profile string           Chef Infra Server URL (default "default")

Use "chef-analyze report [command] --help" for more information about a command.
$
```

## Tasks
### Creating reports for cookbooks
```
$ chef-analyze report cookbooks
$ chef-analyze report cookbooks foo
```

### Creating reports for nodes
```
$ chef-analyze report nodes
$ chef-analyze report nodes bar
```

### Filters: all nodes in an environment
```
$ chef-analyze report node all --environment qa
```
