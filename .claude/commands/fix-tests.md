---
description: Fix failing tests using TDD workflow
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
  - Edit
  - Task
  - TaskCreate
  - TaskUpdate
  - TaskList
skills:
  - tdd-workflow
  - testing-patterns
  - systematic-debugging
---

# Fix Tests

When I run `make test`, `make integration`, `make test-race` and `make test-race-full` I am getting some test failures. Please always remember to follow the coding guidelines in the `docs/CODING_GUIDELINES.md` file. We should have table-driven tests for all the code and aim for 80%+ coverage. Also, please make sure to run `make polish` and fix all the linting errors.

If we have any skipped tests can we fix them, or should we remove them? Please also make sure to cleanup any temporary or backup files you may have created.

**🚨 CRITICAL — TEST BEHAVIOR, NOT IMPLEMENTATION!** Tests should read like business requirements documentation and remain valid even if the implementation changes completely.

Can you please fix all the failing tests and all the linting errors. Please spawn 10 sub-agents to help you resolve these issues. Follow `dispatching-parallel-agents` to scope each agent to one independent failure domain.

**IMPORTANT**: Do NOT stop until ALL tests are passing and ALL linting issues are resolved.

**CRITICAL**: Avoid using the nolint directive and making modifications to the `.golangci.yml` unless absolutely necessary and for good legitimate reasons.

Apply `verification-before-completion` -- run `make lint`, `make test`, `make integration`, `make test-race` and `make test-race-full` and confirm 0 failures before claiming fixed.
