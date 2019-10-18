# Design
This document describes the shape and future of this tool, it will be
a place for us, at Chef, to discuss about the design and usability plus,
we will start the documentation of this tool in an early stage.

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
  config      Create and verify the Chef config
  report      Generate reports about your Chef inventory
  upgrade     Perform automatic upgrades


Flags:
  -s, --chef-server-url string   Chef Infra Server URL
  -c, --config string            Chef config file
  -h, --help                     help for chef-analyze
  -k, --key string               Chef Infra Server API client key
  -u, --user string              Chef Infra Server API client username

Use "chef-analyze [command] --help" for more information about a command.
$
```
### `report` sub-command help
```
$ chef-analyze report --help
```

### `config` sub-command help
```
$ chef-analyze config --help
```

### `upgrade` sub-command help
```
$ chef-analyze upgrade --help
```
## Tasks
### Creating reports for cookbooks
```
$ chef-analyze report cookbook all
$ chef-analyze report cookbook foo
```

### Creating reports for nodes
```
$ chef-analyze report node all
$ chef-analyze report node bar
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
