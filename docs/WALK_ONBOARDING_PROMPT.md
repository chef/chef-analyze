# Walk Onboarding Prompt

Use this prompt when starting a Walk exercise in this repository.

## Prompt

You are contributing to chef-analyze using the Walk workflow.

Follow this sequence:

1. Inspect relevant files and summarize current behavior.
2. Propose a minimal change plan before editing code.
3. Implement behavior-preserving refactors first.
4. Add or update tests that cover modified code paths.
5. Run local validation commands and report exact outputs.
6. Prepare a PR using `walk_pr_template.md`.
7. Include coverage evidence from `coverage/coverage.txt` (the `total:` line).
8. Use a signed commit (`git commit -s`) and push to a Walk branch:
   - `learn/walk/<user>-ex<exercise-number>-<short-topic>`

Constraints:

1. Do not include unrelated file changes.
2. Keep changes small, explicit, and reviewable.
3. Prefer code-level fixes over CI or pipeline changes unless requested.
4. If a command cannot run locally, state the blocker and provide a fallback.

Expected PR sections:

1. Summary
2. Evidence
3. Risk & Rollback
4. Review Focus
5. Track