# chef-analyze
[![Build status](https://badge.buildkite.com/a5dfa44b20a6ec189a93bcbda031db452f1d964fa6836f7065.svg?branch=main)](https://buildkite.com/chef/chef-chef-analyze-main-verify)
[![Code coverage](https://img.shields.io/badge/coverage-95.0%25-brightgreen)](https://buildkite.com/chef/chef-chef-analyze-main-verify)

A CLI to analyze artifacts from a Chef Infra Server.

**Umbrella Project**: [Chef Workstation](https://github.com/chef/chef-oss-practices/blob/main/projects/chef-workstation.md)

**Project State**: [Prototyping](https://github.com/chef/chef-oss-practices/blob/main/repo-management/repo-states.md#prototyping)

**Issues [Response Time Maximum](https://github.com/chef/chef-oss-practices/blob/main/repo-management/repo-states.md)**: Not yet defined

**Pull Request [Response Time Maximum](https://github.com/chef/chef-oss-practices/blob/main/repo-management/repo-states.md)**: Not yet defined

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

### Contract Tests and Golden Updates

Formatter contract tests validate boundary output using golden fixtures in
`pkg/formatter/testdata/`.

Run contract tests only:

```bash
go test ./pkg/formatter -run '^TestContract_'
```

Golden update process:

1. Change formatter behavior intentionally.
2. Run `go test ./pkg/formatter` and inspect failing contract diff output.
3. Update the specific golden fixture(s) in `pkg/formatter/testdata/`.
4. Re-run `go test ./pkg/formatter -run '^TestContract_'` until green.
5. Include rationale for the fixture change in the PR description.

For API-subsystem TODO and edge sweep notes, see
`docs/API_SUBSYSTEM_SWEEP.md`.

### Capture log hook verification

The capture command emits a sourcing summary log line with attempts, resolved,
unresolved, and elapsed time in milliseconds.

Example verification flow:

1. Run a capture:
    ```
    $ chef-analyze capture NODE-NAME
    ```
2. Confirm the output contains:
    ```
    - Capture sourcing summary: attempts=<n> resolved=<n> unresolved=<n> elapsed_ms=<n>
    ```

### Code coverage
This repository requires any change to always increase, or at least, maintain the percentage of code
coverage, to execute the current coverage run:
```
$ hab studio enter
[1][default:/src:0]# code_coverage
```
For details about the code coverage open the generated HTML report located at `coverage/coverage.html`.

The `code_coverage` helper also writes a text report to `coverage/coverage.txt` and prints it to stdout.
The final `total:` line in that file is the overall coverage percentage.

To extract only the total line for PR evidence:

```bash
$ grep '^total:' coverage/coverage.txt
total:                                  (statements)    95.0%
```

PR snippet template (includes total percentage):

```text
Evidence
- Tests/logs/metrics: `unit_tests` and/or `integration_tests`
- Coverage:
    - Total: <XX.X%>
    - Source: `coverage/coverage.txt` (line starts with `total:`)
```

### Architecture Diagram Updates

To regenerate the architecture diagram and update the dependency-change summary:

```bash
make update-architecture
```

This command updates:

- `docs/DESIGN.md` (auto-generated architecture Mermaid block)
- `docs/ARCHITECTURE_CHANGE_SUMMARY.md` (what changed since baseline)
- `docs/architecture/last_edges.txt` (the baseline edge snapshot)

### Security Patch Set Automation

Run strict security checks and optional curated minor dependency updates with:

```bash
bash scripts/security_alert_patchset.sh --apply-minor
```

To auto-push current branch and open a PR (requires authenticated `gh` CLI):

```bash
bash scripts/security_alert_patchset.sh --apply-minor --create-pr --pr-title "Security hygiene patch set"
```

The script runs strict `gosec` parity checks for `cmd/s3_utils.go`, executes
core validation tests, and can apply the curated dependency patch set used in
this repository.

### Reporting Logging and Metrics Validation

For the `pkg/reporting` logging/metrics pattern and validation commands, see:

- `docs/REPORTING_OBSERVABILITY.md`

### Patching a local Chef Workstation Install
You can override the `chef-analyze` binary that comes inside your local Chef Workstation install by
running `make patch_local_workstation` at the top level folder of this repository. Then just simply
run `chef-analyze` or `chef analyze` to use the top-level chef wrapper.

## Contributing

For information on contributing to this project, see [CONTRIBUTING.md](CONTRIBUTING.md).

For Walk onboarding, see [docs/WALK_ONBOARDING_PROMPT.md](docs/WALK_ONBOARDING_PROMPT.md).

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
