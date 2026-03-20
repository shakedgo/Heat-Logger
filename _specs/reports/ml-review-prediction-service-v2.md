# ML Review: `prediction_service_v2.go`

**Date:** 2026-03-20
**Reviewer:** ml-reviewer agent
**File:** `backend/internal/services/prediction_service_v2.go`

---

## Critical Issues

### C1 — `AnchorBoost` defaults to `0.0`, silently zeroing the best records

`NewPredictionServiceV2` never sets `AnchorEpsilon` or `AnchorBoost` in `defaultCfg`, so both stay at Go's zero value.
- `AnchorEpsilon=0` means the anchor condition only fires for satisfaction **exactly** 50.0
- When it fires, `w *= AnchorBoost` → `w *= 0.0` — the weight is **killed**

Every perfect-outcome record (satisfaction=50) gets weight=0 in the default config. These are the most valuable training examples in the system.

**Fix:** add to `defaultCfg` in `NewPredictionServiceV2`:
```go
AnchorEpsilon: 5.0,
AnchorBoost:   3.0,
```

### C2 — Confidence interval uses raw `HeatingTime`, prediction uses `impliedTarget`

`v2HeatingTimeRange` (line 461) reports the range of historical `HeatingTime` values. But the actual prediction is computed from `impliedTarget`, which adjusts those values based on satisfaction. The interval will often not contain the prediction — the frontend shows a range that's entirely wrong.

---

## Statistical Concerns

### S1 — Sharp cliff in `impliedTarget` at satisfaction 48→49

At s=49: no change (inside dead band). At s=48: +13.2% increase. The hot side at the same distance (s=51) gets only −1%. The cold asymmetry is ~13× more aggressive — may be intentional for NeverCold bias, but worth confirming.

### S2 — Reliability weighting double-penalizes extreme-cold records

Records with satisfaction near 1 already have tiny Gaussian spatial weights. The additional `gaussian(sat-50, 22)` reliability factor gives them a further 12× suppression — these are the highest-signal cold records and are being systematically silenced.

### S3 — `NeverCold=true` is inoperative when any records exist

The `NeverCold` flag only triggers on the empty-records early-return path (line 157). In all other cases, `smartRound` takes over, which is unaware of `NeverCold` and may still produce `math.Floor`.

### S4 — Step cap slows cold-start escape

If a user's first similar record has `HeatingTime=5` and the true answer is 15 min, the 35% step cap limits each session to 6.75→9.1→12.3→16.6 — roughly 4 sessions to converge. Not a bug, but a behavioral concern.

---

## Suggestions

### SG1 — Three dead functions

`weightedMean`, `weightedMeanAnchors`, and `latestUserRecord` are defined but never called. They use raw `HeatingTime` while the live code uses `impliedTarget`. Risk: a future dev calls the wrong variant. Should be deleted.

### SG2 — `AnchorBlend`/`AnchorBoost` ordering inconsistency

Ordering differs between the struct definition and `defaultCfg`, making the missing `AnchorBoost` visually harder to spot.

---

## Test Gaps

| # | Gap |
|---|-----|
| TG1 | No test catches the AnchorBoost=0 / zero-weight bug — existing tests pass for the wrong reason (step cap coincidentally produces a passing value) |
| TG2 | No test exercises the full default config end-to-end |
| TG3 | `NeverCold=true` test only covers the empty-records path, not the normal prediction path where it's inoperative |
| TG4 | No assertion that `ConfidenceMin ≤ HeatingTime ≤ ConfidenceMax` |
| TG5 | `TestV2_SingleRecord_NoRange` uses `Satisfaction=50` — hits the broken anchor branch; use `Satisfaction=30` to test the actual kNN path |

---

## Priority Summary

| Priority | Item | Action |
|----------|------|--------|
| P0 | C1: AnchorBoost=0 default | Add `AnchorEpsilon: 5.0, AnchorBoost: 3.0` to `defaultCfg` |
| P1 | C2: Confidence interval mismatch | Fix `v2HeatingTimeRange` to use `impliedTarget` values |
| P2 | S3: NeverCold inoperative | Plumb `NeverCold` flag into `smartRound` |
| P3 | SG1: Dead code | Delete `weightedMean`, `weightedMeanAnchors`, `latestUserRecord` |
| P3 | TG1–TG5: Test gaps | Add targeted tests for each gap |
