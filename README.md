# chef-analyze
[![Build status](https://badge.buildkite.com/a5dfa44b20a6ec189a93bcbda031db452f1d964fa6836f7065.svg)](https://buildkite.com/chef/chef-chef-analyze-master-verify)
[![Code coverage](https://img.shields.io/badge/coverage-89.3%25-green)](https://buildkite.com/chef/chef-chef-analyze-master-verify)

A CLI to analyze artifacts from a Chef Infra Server.

**Umbrella Project**: [Chef Workstation](https://github.com/chef/chef-oss-practices/blob/master/projects/chef-workstation.md)

**Project State**: [Prototyping](https://github.com/chef/chef-oss-practices/blob/master/repo-management/repo-states.md#prototyping)

**Issues [Response Time Maximum](https://github.com/chef/chef-oss-practices/blob/master/repo-management/repo-states.md)**: Not yet defined

**Pull Request [Response Time Maximum](https://github.com/chef/chef-oss-practices/blob/master/repo-management/repo-states.md)**: Not yet defined

## Installation

We highly recommend using [Chef Workstation](https://downloads.chef.io/chef-workstation/), which includes
chef-analyze out of the box. If for some reason you can't use Chef Workstation you can manually install the
[Habitat package `chef/chef-analyze`](https://bldr.habitat.sh/#/pkgs/chef/chef-analyze/latest).

```bash
hab pkg install -b chef/chef-analyze
```

## Development Documentation

The development of this CLI is being done inside a [Chef Habitat Studio](https://www.habitat.sh/docs/glossary/#glossary-studio),
you will need to have [Chef Habitat installed](https://www.habitat.sh/docs/install-habitat/) on your local workstation
to proceed with any development task.

### Building

From within a Chef Habitat Studio, you can build the chef-analyze CLI by:

1. Building a native Habitat package.
    ```
    $ hab studio enter
    [1][default:/src:0]# build
    ```
2. Helper method to build cross-platform binaries.
    ```
    $ hab studio enter
    [1][default:/src:0]# build_cross_platform
    ```
    __NOTE:__ The generated binaries will be located inside the `bin/` directoy.

### Testing

From within a Chef Habitat Studio, you can run both, unit and integration tests:
1. Unit tests. ([Go-based](https://golang.org/pkg/testing/))
    ```
    $ hab studio enter
    [1][default:/src:0]# unit_tests
    ```
2. Integration tests. ([Go-based](https://golang.org/pkg/testing/))
    ```
    $ hab studio enter
    [1][default:/src:0]# integration_tests
    ```
    __NOTE:__ The integration tests require a binary to test against, this helper automatically triggers
    a cross-platform build and uses the generated binary for the running platform.

### Code coverage
This repository requires any change to always increase, or at least, maintain the percentage of code
coverage, to execute the current coverage run:
```
$ hab studio enter
[1][default:/src:0]# code_coverage
```
For details about the code coverage open the generated HTML report located at `coverage/coverage.html`.

### Patching a local Chef Workstation Install
You can override the `chef-analyze` binary that comes inside your local Chef Workstation install by
running `make patch_local_workstation` at the top level folder of this repository. Then just simply
run `chef-analyze` or `chef analyze` to use the top-level chef wrapper.

## Contributing

For information on contributing to this project please see our [Contributing Documentation](https://github.com/chef/chef/blob/master/CONTRIBUTING.md)

## License & Copyright

- Copyright:: Copyright (c) 2019 Chef Software, Inc.
- License:: Apache License, Version 2.0

```text
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
