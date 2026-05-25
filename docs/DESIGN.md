# Design
This document describes the current and future shape of this tool, it will
be a place for us, at Chef, to discuss about the design and usability, the
flow of commands that we expect user to follow plus, we will start the
documentation of this tool in an early stage.

## Architecture

The architecture map below uses real repository paths so each node maps to
code that exists in this project.

```mermaid
flowchart TD
  A[CLI Entry\nmain.go] --> B[Command Router\ncmd/root.go]
  B --> C[Report Commands\ncmd/report.go]
  B --> D[Capture Command\ncmd/capture.go]
  C --> E[Reporting Service\npkg/reporting/*.go]
  C --> F[Formatters\npkg/formatter/*.go]
  C --> G[Report Output Writer\ncmd/report.go saveReport/saveErrorReport]
  D --> E
  E --> H[Chef API Adapter\npkg/reporting/client.go]
  E --> I[Snapshot Writer\npkg/reporting/object_writer.go]
  E --> J[Cookstyle Runner\npkg/reporting/cookstyle.go]
  G --> K[Chef Workstation Filesystem\n$HOME/.chef-workstation/reports and errors]
  I --> L[Captured Node Repo\n./node-<name>-repo]
```

### Data Flows

1. Nodes report flow

```mermaid
sequenceDiagram
  participant U as User
  participant CLI as cmd/report.go (nodes)
  participant RS as pkg/reporting/nodes.go
  participant CA as pkg/reporting/client.go
  participant F as pkg/formatter/*.go
  participant FS as Local Filesystem

  U->>CLI: chef-analyze report nodes
  CLI->>CA: NewChefClient(config + credentials)
  CLI->>RS: GenerateNodesReport(client, filter, anonymize)
  RS->>CA: Search.PartialExec("node", ...)
  RS-->>CLI: []NodeReportItem
  CLI->>F: MakeNodesReportTXT/CSV
  CLI->>FS: saveReport + saveErrorReport
```

2. Cookbooks report flow

```mermaid
sequenceDiagram
  participant U as User
  participant CLI as cmd/report.go (cookbooks)
  participant CR as pkg/reporting/cookbooks.go
  participant API as Chef Infra Server APIs
  participant CS as pkg/reporting/cookstyle.go
  participant F as pkg/formatter/*.go
  participant FS as Local Filesystem

  U->>CLI: chef-analyze report cookbooks
  CLI->>CR: NewCookbooksReport(...)
  CR->>API: ListAvailableVersions + policy/cookbook artifacts
  CLI->>CR: Generate()
  CR->>API: Download cookbook/cookbook artifact
  CR->>API: Lookup nodes using cookbook version
  CR->>CS: runCookstyleFor(record) (optional)
  CR-->>CLI: Records + errors
  CLI->>F: MakeCookbooksReportTXT/CSV
  CLI->>FS: saveReport + saveErrorReport
```

3. Capture command flow

```mermaid
sequenceDiagram
  participant U as User
  participant CLI as cmd/capture.go
  participant NC as pkg/reporting/capture.go
  participant API as Chef Infra Server APIs
  participant OW as pkg/reporting/object_writer.go
  participant R as Local Node Repo

  U->>CLI: chef-analyze capture NODE
  CLI->>NC: NewNodeCapture(...)
  CLI->>NC: Run()
  NC->>API: Fetch node, roles, env, cookbooks, policy, data bags
  NC->>OW: Write JSON objects and config files
  OW->>R: Write roles/, environments/, nodes/, data_bags/, policies/
```

## Help Documentation

### Main help
```bash
$ chef-analyze help
Analyze your Chef Infra Server artifacts to understand the effort to upgrade
your infrastructure by generating reports, automatically fixing violations
and/or deprecations, and generating Effortless packages.

Usage:
  chef-analyze [command]

Available Commands:
  help        Help about any command
  report      Generate reports from a Chef Infra Server

Flags:
  -s, --chef_server_url string   Chef Infra Server URL
  -k, --client_key string        Chef Infra Server API client key
  -n, --client_name string       Chef Infra Server API client username
  -c, --credentials string       Chef credentials file (default $HOME/.chef/credentials)
  -h, --help                     help for chef-analyze
  -o, --ssl-no-verify            Disable SSL certificate verification
  -p, --profile string           Chef Infra Server URL (default "default")

Use "chef-analyze [command] --help" for more information about a command.
$
```
### `report` sub-command help
```
$ chef-analyze report --help
Generate reports from a Chef Infra Server

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
  -o, --ssl-no-verify            Disable SSL certificate verification
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
