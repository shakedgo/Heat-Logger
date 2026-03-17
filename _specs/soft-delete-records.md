# Spec for soft-delete-records

branch: claude/feature/soft-delete-records

## Summary

Instead of permanently removing records when a user deletes them, mark them as deleted (soft delete). This enables an "Undo" experience after deletion, supports future audit capabilities, and aligns with the backlog items for undo-on-delete and the architecture goal of non-destructive record management.

## Functional Requirements

- Add a `deleted_at` (nullable timestamp) field to the `DailyRecord` database model
- When a user deletes a single record, set `deleted_at` to the current timestamp rather than removing the row
- When a user triggers "delete all", soft-delete all records (set `deleted_at` on each) rather than truncating
- The history list (`GET /api/history`) must only return records where `deleted_at` is NULL
- After a single-record delete, display a brief (5–10 second) "Undo" toast; if the user taps Undo, clear `deleted_at` on that record to restore it
- The CSV export (`GET /api/history/export`) must also exclude soft-deleted records
- Prediction algorithms must not use soft-deleted records as training data (filter them out when querying history)

## Possible Edge Cases

- User deletes a record and immediately navigates away before the Undo window closes — soft-deleted record should remain soft-deleted (no auto-restore)
- "Delete all" followed by "Undo" — undo should only be available for single-record deletion; bulk delete has no undo
- Very large history: soft-deleted rows accumulate over time and are never purged — consider noting this as a follow-up (scheduled cleanup) but do not implement now
- Re-running predictions after an undo should immediately include the restored record
- Exporting CSV while an Undo window is open (record is soft-deleted but not yet confirmed) — export should not include the record until it is restored

## Acceptance Criteria

- Deleting a record via the UI no longer removes the row from the database; `deleted_at` is set instead
- The history list and CSV export never surface soft-deleted records
- An "Undo" toast appears for 5–10 seconds after a single-record delete; clicking it restores the record and it reappears in the history list
- Prediction results are unaffected by soft-deleted records (they are excluded from neighbor queries)
- "Delete all" soft-deletes all records with no undo option; the UI confirmation message reflects this
- No regression in existing history, export, or prediction behavior for non-deleted records

## Open Questions

- Should soft-deleted records ever be hard-purged (e.g., after 30 days)? Out of scope for now but should be noted as a follow-up.
- Should an admin/debug endpoint exist to list or restore soft-deleted records? Not required for this iteration.
- Is the Undo window duration (5–10 s) configurable, or hardcoded? Keep hardcoded for now to keep scope small.

## Testing Guidelines

Create test file(s) in the `./tests` folder for the new feature, and create meaningful tests for the following cases, without going too heavy:

- Deleting a record sets `deleted_at` and removes it from `GET /api/history` and CSV export responses
- Restoring (undo) a record clears `deleted_at` and it reappears in `GET /api/history`
- "Delete all" soft-deletes all records; subsequent `GET /api/history` returns empty list
- Prediction service excludes soft-deleted records from neighbor queries
- Undo is not available / has no effect after the undo window has expired (or after a page reload)
