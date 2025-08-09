# Water-Heater Predictor — Backend Changes (Go)

**Scope:** Concrete code changes to stabilize learning and make predictions consistent with your 1–100 scale (50 = perfect).
**Note:** Feature gaps are intentionally excluded.

---

## Step 1 — Align “success anchors” to 50 (±δ)

**What to change**

* In your anchor finder (e.g., `findWeightedSuccessAnchors`), treat *success* as `abs(Satisfaction-50) <= δ` (δ≈5–7).
* Weight anchors by **recency** and **closeness to 50** (e.g., `w = exp(-λ·daysAgo) * (1 - abs(score-50)/δ)`).

**Why**

* Your scale defines 50 as ideal; using >55 as “success” biases the model to run cooler over time.

**Done when**

* The top 1–3 anchors are near 50 and recent, not just “hot” days.

---

## Step 2 — Make adjustments symmetric around 50

**What to change**

* In `applyGraduatedAdjustment`, compute `e = score - 50`.

  * If `|e| <= 5` ⇒ no change (deadband).
  * If `e > 0` (too hot) ⇒ **reduce** time.
  * If `e < 0` (too cold) ⇒ **increase** time.
* Keep the same magnitude rule for both directions.

**Why**

* Your previous logic reduced time for hot days but didn’t mirror for cold days → drift.

**Done when**

* Equal deviations above/below 50 produce equal-sized (opposite-sign) adjustments.

---

## Step 3 — Adjust in minutes, not % of prior time

**What to change**

* Replace `adjustment = percent * previousHeatingTime` with:

  * `adjustmentMinutes = k * f(|e|)` (e.g., piecewise-linear or small quadratic), **capped** (±8 min/day).
* Apply sign from Step 2.

**Why**

* Percent-based updates overreact for long times and underreact for short times.

**Done when**

* The same error size yields the same minute change regardless of prior heating time.

---

## Step 4 — Replace hard similarity gates with soft kernels

**What to change**

* In `findSimilarRecords` (or equivalent):

  * Remove hard filters like `|Δtemp|>2` or `|Δduration|>3`.
  * Use Gaussian-like weights:
    `wT = exp(-(ΔT/σT)^2)`, `wD = exp(-(ΔD/σD)^2)`; keep your **recency decay**; final `w = wT*wD*recency`.
  * Ignore only negligible weights (e.g., `w < 1e-3`).

**Why**

* Hard thresholds throw away data and cause brittle jumps; kernels smooth influence.

**Done when**

* Most historical records contribute small-but-nonzero weight; similar ones dominate.

---

## Step 5 — Continuous user/global blending

**What to change**

* Replace discrete user-weight rules with:

  * `N_eff = sum(weights of user neighbors)`
  * `userWeight = 1 - exp(-N_eff/τ)` (τ≈10–20), **clamped** to `[0.2, 0.9]`.
* Blend user and global predictions with `userWeight`.

**Why**

* Provides a smooth ramp from global → personal model with recency and density.

**Done when**

* New users lean global; established users lean personal, without sudden jumps.

---

## Step 6 — Unify global estimation (drop hand-made clusters)

**What to change**

* Remove if/else “clusters” (e.g., duration/temperature buckets).
* Compute the **global** baseline via the same soft-kNN as Step 4 but over **all users**.

**Why**

* Manual clusters are lossy; the kernel method consistently captures gradients.

**Done when**

* Global baseline smoothly reflects temp/duration without step changes.

---

## Step 7 — Learn cold-start coefficients (simple linear/ridge)

**What to change**

* Replace hard-coded fallback `time = a + b*duration + c*temp` with learned `(a,b,c)` from historical data (ridge preferred).
* Store coefficients; reuse for new users or sparse contexts.

**Why**

* Data-driven coefficients outperform guesswork and improve seasonality transfer.

**Done when**

* Cold-start predictions match aggregate behavior without manual tuning.

---

## Step 8 — Remove asymmetric gates elsewhere (overshoot/dampening)

**What to change**

* Review all conditionals tied to satisfaction:

  * If any path is enabled for `score<50` but not mirrored for `score>50` (or vice versa), make it symmetric.
  * Keep the same deadband as in Step 2.

**Why**

* Hidden asymmetries reintroduce bias after Step 2.

**Done when**

* Every heuristic that reacts to error does so symmetrically around 50.

---

## Step 9 — Add monotonicity guards (final sanity pass)

**What to change**

* After blending and adjustments, apply guards:

  * With temperature fixed, `PredTime` should **not decrease** when `Duration` increases.
  * With duration fixed, `PredTime` should **not increase** when `Temperature` increases.
* Allow a tiny tolerance (e.g., ±0.5 min) to avoid over-clamping.

**Why**

* Prevents counter-intuitive outputs from noisy neighbors.

**Done when**

* Increasing shower duration never yields a shorter preheat, and vice versa for temp.

---

## Step 10 — Preserve precision; round only for UI (+ optional uncertainty)

**What to change**

* Keep internal predictions as `float64`.
* Return both:

  * `predMinutes` (raw float)
  * `uiMinutes` (rounded to nearest 0.5 or 1.0 for display)
* (Optional) Add `uncertaintyMinutes` from weighted neighbor dispersion (IQR/STD).

**Why**

* Rounding early loses learning signal; exposing a rounded value keeps UX clean.

**Done when**

* API consumers can use `uiMinutes`, while learning and diagnostics use `predMinutes`.

---

## Acceptance Checklist

* [ ] Anchors centered at 50±δ with recency/closeness weights.
* [ ] Symmetric, deadbanded adjustments around 50.
* [ ] Minute-based, capped update magnitudes.
* [ ] Soft kernel similarity (temp & duration) + recency; no hard gates.
* [ ] Smooth user/global blend via `N_eff` with clamp.
* [ ] Global baseline via the same kNN, no manual clusters.
* [ ] Cold-start uses learned linear/ridge coefficients.
* [ ] No asymmetric conditionals left in adjustment logic.
* [ ] Monotonicity guards applied post-blend.
* [ ] API returns raw + rounded minutes (and optional uncertainty).
