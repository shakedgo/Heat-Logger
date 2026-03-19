# Spec for prediction-confidence

branch: claude/feature/prediction-confidence
figma_component (if used): N/A

## Summary
The `/api/calculate` endpoint currently returns only `{ heatingTime: X }`. This feature adds `sampleCount` (number of historical neighbors used to compute the prediction) and optionally a `min`/`max` range to the response, so the UI can surface a confidence signal to the user — e.g. "based on 8 similar sessions" or a range indicator when the prediction is less certain.

## Functional Requirements
- The `/api/calculate` response must include `sampleCount` — the number of historical records that contributed to the prediction.
- The response should optionally include `confidenceMin` and `confidenceMax` (the lowest and highest heating time among contributing neighbors) when `sampleCount > 1`.
- When there are no historical records and the prediction is a cold-start default, `sampleCount` should be `0` and no range should be returned.
- The frontend should display the sample count near the predicted heating time (e.g. "based on N sessions").
- If `sampleCount` is 0, the UI should indicate this is an initial estimate with no prior data.
- The confidence range (`min`–`max`), if present, may be shown as a subtle secondary detail.

## Figma Design Reference (only if referenced)
- N/A

## Possible Edge Cases
- `sampleCount = 0`: cold-start prediction — UI must handle missing range gracefully.
- `sampleCount = 1`: range is meaningless (min === max) — omit or collapse range display.
- Very wide `min`/`max` range (high variance neighbors) — UI should not alarm the user; treat it as informational only.
- Both predictor versions (v1 and v2) must return the same response shape; the fields can differ in value but must always be present.

## Acceptance Criteria
- `POST /api/calculate` response always includes `sampleCount` (integer ≥ 0).
- When `sampleCount > 1`, response includes `confidenceMin` and `confidenceMax` (floats, in minutes).
- Frontend displays "based on N sessions" (or equivalent) beneath or near the heating time result.
- When `sampleCount = 0`, frontend shows a distinct "initial estimate" label instead of a session count.
- Existing frontend and backend tests continue to pass.
- Both v1 and v2 predictors satisfy the updated `Predictor` interface contract.

## Open Questions
- Should `confidenceMin`/`confidenceMax` be exposed in the API response now, or deferred to a follow-up? (MVP could ship with just `sampleCount`.) - You decide
- What copy/label works best in the UI — "sessions", "data points", "similar showers"? - "similar showers"
- Should the range be visualised (e.g. a small bar) or just shown as text? just text for now

## Testing Guidelines
Create a test file(s) in the ./tests folder for the new feature, and create meaningful tests for the following cases, without going too heavy:
- Cold-start: no records in DB → `sampleCount = 0`, no range fields.
- Single record: one matching neighbor → `sampleCount = 1`, no range (or min === max).
- Multiple records: several matching neighbors → `sampleCount > 1`, `confidenceMin` < `confidenceMax`.
- UI renders "based on N sessions" when `sampleCount > 0`.
- UI renders initial-estimate label when `sampleCount = 0`.
