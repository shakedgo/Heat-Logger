# Plan: Prediction Confidence & Sample Count

## Context
The `/api/calculate` endpoint returns only `{ heatingTime }`. Users have no signal for how confident the prediction is or how much data backs it. This adds `sampleCount`, `confidenceMin`, and `confidenceMax` to the response so the UI can show "based on N similar showers" and a heating time range. `confidenceMin`/`max` are included now (not deferred) since the data is already in memory in both predictors.

---

## Step 1 — Extend `PredictionResponse` struct
**File:** `backend/internal/services/prediction_service.go` (~line 37)

```go
type PredictionResponse struct {
    HeatingTime   float64  `json:"heatingTime"`
    SampleCount   int      `json:"sampleCount"`
    ConfidenceMin *float64 `json:"confidenceMin,omitempty"`
    ConfidenceMax *float64 `json:"confidenceMax,omitempty"`
}
```
Pointer + `omitempty` means the JSON fields are absent (not `null`) when `sampleCount <= 1`.

---

## Step 2 — V1: populate fields in `PredictHeatingTime`
**File:** `backend/internal/services/prediction_service.go` (~line 49)

After `getCombinedPrediction` returns `heatingTime`, call `findSimilarRecords` on the merged record pool (userRecords + globalRecords) — for metadata only, no change to prediction logic:

```go
allRecords := append(userRecords, globalRecords...)
similar := s.findSimilarRecords(req, allRecords)

resp := &PredictionResponse{
    HeatingTime: math.Round(heatingTime),
    SampleCount: len(similar),
}
if len(similar) > 1 {
    min, max := heatingTimeRange(similar)
    resp.ConfidenceMin = &min
    resp.ConfidenceMax = &max
}
return resp, nil
```

Add private helper at bottom of same file:
```go
func heatingTimeRange(records []SimilarRecord) (min, max float64) {
    min, max = records[0].Record.HeatingTime, records[0].Record.HeatingTime
    for _, r := range records[1:] {
        if r.Record.HeatingTime < min { min = r.Record.HeatingTime }
        if r.Record.HeatingTime > max { max = r.Record.HeatingTime }
    }
    return
}
```

---

## Step 3 — V2: populate fields in `Predict`
**File:** `backend/internal/services/prediction_service_v2.go`

1. **Cold-start early return** (~line 162) — add `SampleCount: 0` explicitly.
2. **Final return** (~line 245) — replace with:

```go
resp := &PredictionResponse{HeatingTime: estAll, SampleCount: len(top)}
if len(top) > 1 {
    min, max := v2HeatingTimeRange(top)
    resp.ConfidenceMin = &min
    resp.ConfidenceMax = &max
}
return resp, nil
```

Add helper at bottom of same file:
```go
func v2HeatingTimeRange(recs []recWrap) (min, max float64) {
    min, max = recs[0].rec.HeatingTime, recs[0].rec.HeatingTime
    for _, r := range recs[1:] {
        if r.rec.HeatingTime < min { min = r.rec.HeatingTime }
        if r.rec.HeatingTime > max { max = r.rec.HeatingTime }
    }
    return
}
```

---

## Step 4 — Frontend: App.vue
**File:** `frontend/src/App.vue`

- Add `latestSampleCount: null`, `latestConfidenceMin: null`, `latestConfidenceMax: null` to `data()`
- In `handleCalculate`: destructure and store all four fields from `response.data`
- Clear all four fields on feedback submit / form reset
- Pass 3 new props to `<InputForm>`: `:latestSampleCount`, `:latestConfidenceMin`, `:latestConfidenceMax`

---

## Step 5 — Frontend: InputForm.vue
**File:** `frontend/src/components/InputForm.vue`

- Add 3 new props: `latestSampleCount`, `latestConfidenceMin`, `latestConfidenceMax` (all `Number`, default `null`)
- Extend `latestHeatingTime` watcher to also mirror the 3 new fields into `currentEntry`
- Add below the heating time value in `.cv-suggestion`:

```html
<div class="cv-meta">
  <span v-if="currentEntry.sampleCount > 0">
    Based on {{ currentEntry.sampleCount }} similar shower{{ currentEntry.sampleCount === 1 ? '' : 's' }}
  </span>
  <span v-else>Initial estimate</span>
  <span v-if="currentEntry.sampleCount > 1 && currentEntry.confidenceMin != null" class="cv-range">
    &nbsp;&bull;&nbsp;Range: {{ currentEntry.confidenceMin.toFixed(1) }} – {{ currentEntry.confidenceMax.toFixed(1) }} min
  </span>
</div>
```

Add styles in scoped `<style>`:
```scss
.cv-meta { font-size: 0.8rem; color: #047857; margin-top: 4px; }
.cv-range { opacity: 0.8; }
```

---

## Step 6 — Tests
**File:** `backend/internal/services/prediction_confidence_test.go` (new file, same package — reuses existing `MockRecordService`)

| Test | Setup | Assert |
|---|---|---|
| V1 cold-start | 0 records | `SampleCount=0`, both nil |
| V1 single match | 1 record within ±2°C/±3min | `SampleCount=1`, both nil |
| V1 multi-match | 3+ matching records | `SampleCount>=2`, `ConfidenceMin < ConfidenceMax` |
| V2 cold-start | 0 records | `SampleCount=0`, both nil |
| V2 single record | 1 record | `SampleCount=1`, nil range |
| V2 multi-records | 5+ records | `SampleCount>1`, `ConfidenceMin <= ConfidenceMax` |

Existing 8 tests need no changes — they only assert `HeatingTime`; new int/pointer fields don't break them.

---

## Verification
1. `cd backend && go test ./internal/services/...` — all tests pass
2. Start dev server (`./run-dev.sh`), request a prediction with no history → response has `sampleCount: 0`, no range fields
3. Submit a few feedback records, re-predict → `sampleCount > 0`, "based on N similar showers" appears in UI
4. After 2+ records → range appears as text below the count
