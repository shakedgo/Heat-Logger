# Spec for Estimated Savings Summary

branch: claude/feature/estimated-savings-summary
figma_component (if used): N/A

## Summary

Add a summary card at the top of the history view that shows users how much total heating time they have saved compared to a configurable baseline. The goal is to make the app's value visible — users should be able to see concretely that the predictions are saving them time (and energy) over a "heat for X minutes no matter what" approach. This is computed entirely client-side from existing `/api/history` data; no backend changes are required.

## Functional Requirements

- Display a summary card above the history list in `HistoryList.vue`
- Compute "minutes saved" by comparing each record's `heatingTime` to a baseline value
- The baseline is the user's personal average heating time from their first N records (e.g. first 5), or a sensible app-wide default (e.g. 45 minutes) if there is insufficient history
- Only include records where satisfaction is in the "near-perfect" range (40–60) in the savings calculation — these are the records where the prediction was accurate and the user actually used the suggested time
- Show savings broken down into two time ranges: "This week" and "This month"
- Also display the user's average satisfaction score for the selected time range alongside the savings figure
- If there are no qualifying records in a time range, show a neutral placeholder (e.g. "No data yet")
- The card should be visually distinct but lightweight — not a full chart, just a compact stat summary
- The baseline value used for calculation should be visible to the user (e.g. "vs. your baseline of 45 min") so the figure is interpretable

## Figma Design Reference (only if referenced)

N/A

## Possible Edge Cases

- User has zero history: card should render gracefully with empty-state messaging
- User has fewer than N records to establish a personal baseline: fall back to the app default baseline
- All records fall outside the near-perfect satisfaction range: savings will be zero or negative; display neutrally without implying the app is performing poorly
- Records where `heatingTime` exceeds the baseline (i.e. prediction was higher than baseline): these count as negative savings and should either be excluded or shown as 0 to avoid confusing/discouraging the user
- Week/month boundaries: use the record's stored date, not the current time, to assign records to buckets
- User switches `userId` mid-session: the summary should update to reflect only the active user's records

## Acceptance Criteria

- A summary card is visible above the history list whenever the history view is shown
- The card displays "minutes saved this week" and "minutes saved this month" stats
- The card displays the average satisfaction score for each time range
- The baseline used in calculation is shown to the user
- Savings are only computed from near-perfect (satisfaction 40–60) records
- If a time range has no qualifying records, a neutral placeholder is shown instead of a number
- The card renders without errors when history is empty
- The computed savings figure is consistent with manual calculation against the raw history data

## Open Questions

- Should the baseline be user-configurable (via a settings input), or always derived automatically from early history? user-configurable
- Should negative savings (predictions higher than baseline) be shown as 0 or excluded entirely, to avoid discouraging users?excluded entirely
- Should the qualifying satisfaction range be derived from the same threshold the v2 prediction algorithm uses internally for anchor boost, so that "counts as saved" in the UI matches "counts as a success" in the ML model — rather than using an independently chosen range like 40–60? yes, derive it from the algorithm
- Is "this week" calendar-week (Mon–Sun) or rolling 7 days? Same question for "this month." rolling

## Testing Guidelines

Create a test file(s) in the ./tests folder for the new feature, and create meaningful tests for the following cases, without going too heavy:

- Savings calculation returns correct value when all records are in the near-perfect range
- Savings calculation returns 0 (not negative) when all predictions exceeded the baseline
- Records outside the near-perfect satisfaction range are excluded from the savings total
- Week and month bucketing correctly assigns records to the right time range
- Baseline falls back to the app default when fewer than N records exist
- Empty history renders the card without errors
