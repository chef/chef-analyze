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
violations and/or deprecations and generating Effortless packages.

Usage:
  chef-analyze [command]

Available Commands:
  help        Help about any command
  config      Manage your local Chef configuration (default: $HOME/.chef/credentials)
  report      Generate reports about your Chef inventory
  remediate   Perform automatic remediations


Flags:
  -s, --chef-server-url string   Chef Infra Server URL
  -c, --config string            Chef config file
  -h, --help                     help for chef-analyze
  -k, --key string               Chef Infra Server API client key
  -u, --user string              Chef Infra Server API client username
  -p, --profile string           The credentials profile to use (default: default)

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
  -c, --config string            Chef config file (default: $HOME/.chef/credentials)
  -p, --profile string           Chef Infra Server URL

Use "chef-analyze report [command] --help" for more information about a command.
$
```

### `config` sub-command help
```
$ chef-analyze config --help
Manage your local Chef configuration (default: $HOME/.chef/credentials)

Usage:
  chef-analyze config [command]

Available Commands:
  init        Initialize a local Chef configuration
  verify      Verify your Chef configuration

Flags:
  -h, --help   help for config

Global Flags:
  -s, --chef_server_url string   Chef Infra Server URL
  -k, --client_key string        Chef Infra Server API client key
  -n, --client_name string       Chef Infra Server API client username
  -c, --config string            Chef config file (default: $HOME/.chef/credentials)
  -p, --profile string           The credentials profile to use (default: default)

Use "chef-analyze config [command] --help" for more information about a command.
```

### `remediate` sub-command help
```
$ chef-analyze remediate --help
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

# Other sub-command alternatives
```
$ chef-analyze cookbooks gen-report
$ chef-analyze nodes gen-report

$ chef-analyze cookbooks upgrade
$ chef-analyze nodes upgrade
```
