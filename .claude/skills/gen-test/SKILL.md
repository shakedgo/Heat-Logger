---
name: gen-test
description: Generate Go unit tests using testify for a given service file or function. Use when the user asks to add tests, write unit tests, increase test coverage, or test a specific function. Also triggers on "test this", "add tests for", "write tests for", "cover this function".
disable-model-invocation: true
agent: general-purpose
context: fork
allowed-tools: Read, Grep, Glob, Edit, Write
---

# Generate Go Tests

Generate thorough Go unit tests for the target file or function using testify, following the patterns already established in this codebase.

## Arguments
The user will specify a file path or function name. If none given, ask which service/function to test.

## Steps

1. **Read the target file** — understand all exported functions, their signatures, inputs, and return types.

2. **Read existing tests** in the same package for style reference:
   - `backend/internal/services/prediction_service_test.go`
   - `backend/internal/services/record_service_test.go`

3. **Generate tests** following these conventions from the codebase:
   - Package: `package services` (same package as source)
   - Use `github.com/stretchr/testify/assert` for assertions
   - Use `github.com/stretchr/testify/mock` for mocking interfaces
   - Test naming: `TestFunctionName_Scenario_ExpectedBehavior`
   - Use Arrange/Act/Assert comments
   - Mock external dependencies (DB calls) using the mock pattern already in `prediction_service_test.go`
   - Cover: happy path, edge cases (empty input, zero values, boundary conditions), error cases

4. **Write the test file** alongside the source file (e.g., `foo.go` → `foo_test.go`). If `_test.go` already exists, append new test functions.

5. **Verify** the tests compile by checking imports match the actual types used.

## Output
Tell the user what tests were added and which scenarios they cover.
