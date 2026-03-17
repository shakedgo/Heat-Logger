# Heat-Logger Improvements Backlog

## Medium Priority

### 7. Delete-all confirmation should show record count
The current message is "This cannot be undone." Change it to "Delete **N** records? This cannot be undone." so the user understands the scope.

### 8. Form should preserve inputs after feedback submission
After submitting satisfaction, the form resets to blank. Preserve `averageTemperature` and `showerDuration` so the user can quickly log another session without re-typing.

---

## Lower Priority

### 9. Add structured logging to prediction algorithm
`PredictionServiceV2` has no logging. Add structured log lines (userId, top-k neighbor count, blend ratio, final clamped value) to make it possible to debug unexpected predictions without adding print statements.

### 10. Standardize API error response shape
Some endpoints return `{error: "..."}`, others `{success: true, message: "..."}`. Pick one shape (e.g. `{error?: string, data?: any}`) and apply it everywhere.

### 11. Pagination / virtual scrolling for history
Currently all records are fetched and rendered at once. Add pagination or virtual scrolling before list sizes grow beyond ~200 entries.

### 12. Split prediction service helpers into a separate file
`prediction_service_v2.go` is ~467 lines. Extract math helpers (`gaussian`, `expHalfLife`, `clamp`, etc.) into `prediction_math.go` for easier testing and readability.

### 13. Add time-based history filtering
Let users filter history to "last 7 days", "this month", etc. Currently the full history is always shown and exported.

### 14. Mobile UX pass
The satisfaction bar in HistoryList can overflow on small screens. Review and fix layout on viewports < 600px.

### 15. Add a health-check banner
On app mount, ping a `/api/health` endpoint. If it fails, show a persistent banner ("Backend unavailable") instead of silently failing on every action.

---

## Architecture / Future

- **Multi-user auth**: UserId is client-generated with no auth. Consider adding proper user accounts or at least device-bound tokens to prevent data mixing.
- **Analytics view**: Show trends like average satisfaction over time, prediction accuracy delta, and temperature sensitivity.
- **Batch operations**: Select multiple history records for bulk delete or export of a filtered subset.
