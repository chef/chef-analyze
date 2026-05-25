# Contributing to chef-analyze

This repository follows the standard Chef open source contribution process and
also uses a Walk workflow for guided contribution exercises.

## Prerequisites

1. Install Chef Habitat.
2. Clone this repository.
3. Enter a Habitat Studio from repository root:

```bash
hab studio enter
```

## Local Development Workflow

1. Create a branch from `main`.
2. Use the Walk naming style for exercise branches:
   - `learn/walk/<user>-ex<exercise-number>-<short-topic>`
3. Implement the change in small, reviewable commits.
4. Use signed commits (required for DCO):

```bash
git commit -s -m "<message>"
```

## Build and Test

Run in Habitat Studio whenever possible:

```bash
unit_tests
integration_tests
code_coverage
```

Coverage evidence for PRs:

```bash
grep '^total:' coverage/coverage.txt
```

## Walk PR Requirements

1. Open a PR to `main` from your Walk branch.
2. Use the template in `walk_pr_template.md`.
3. Include:
   - clear summary
   - tests/logs evidence
   - coverage total percentage
   - risk and rollback statement
4. Keep PR scope focused and avoid unrelated changes.

## Suggested Walk Sequence

1. Analyze current code and identify the smallest safe change.
2. Refactor behavior-preserving code first.
3. Add or update tests for changed behavior.
4. Validate coverage and gather proof for the PR body.
5. Open PR and respond to review comments.

## Onboarding Prompt

For a ready-to-use contributor prompt, see:

- `docs/WALK_ONBOARDING_PROMPT.md`