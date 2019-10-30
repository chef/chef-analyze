# chef-analyze
[![Build status](https://badge.buildkite.com/a5dfa44b20a6ec189a93bcbda031db452f1d964fa6836f7065.svg)](https://buildkite.com/chef/chef-chef-analyze-master-verify)
[![Code coverage](https://img.shields.io/badge/coverage-78.0%25-yellow)](https://buildkite.com/chef/chef-chef-analyze-master-verify)

A CLI to analyze your Chef Inventory.

**Umbrella Project**: [Chef Workstation](https://github.com/chef/chef-oss-practices/blob/master/projects/chef-workstation.md)

**Project State**: [Active](https://github.com/chef/chef-oss-practices/blob/master/repo-management/repo-states.md#active)

**Issues [Response Time Maximum](https://github.com/chef/chef-oss-practices/blob/master/repo-management/repo-states.md)**: 14 days

**Pull Request [Response Time Maximum](https://github.com/chef/chef-oss-practices/blob/master/repo-management/repo-states.md)**: 14 days

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
2. Helper method to build a cross-platform binaries.
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
