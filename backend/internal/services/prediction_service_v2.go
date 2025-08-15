package services

import (
	"fmt"
	"math"
	"sort"
	"time"

	"heat-logger/internal/models"
)

// -------------------------------
// Prediction v2 (Gaussian‑kNN)
// -------------------------------
// This file adds a new, production‑ready predictor that avoids brittle box filters,
// blends user + global data with distance kernels, and uses symmetric "success anchors".
//
// Key ideas:
//  - Distance weighting: Gaussian on duration (minutes) and temperature (°C).
//  - Recency decay: half‑life in days.
//  - Anchor boost: near‑perfect satisfaction (|s-50|<=ε) gets extra weight on both sides.
//  - User vs global: explicit userBoost multiplier instead of hard fallbacks.
//  - Frequency control: down‑weight repeated (duration,temp) cells so one context doesn't dominate.
//  - Safety clamps: configurable % step cap vs last user record and absolute [min,max] bounds.
//  - Risk policy: "NeverCold" prefers rounding up; otherwise round to nearest minute.
//
// Integrate by constructing PredictionServiceV2 and calling Predict(req).
// You can keep the old service side‑by‑side during rollout.

type recWrap struct {
	rec     models.DailyRecord
	isUser  bool
	weight  float64
	anchor  bool
	cellKey string
}

type PredictionServiceV2 struct {
	recordService RecordServiceInterface
	cfg           PredictionConfigV2
}

type PredictionConfigV2 struct {
	// Gaussian kernel sigmas
	SigmaDuration float64 // minutes
	SigmaTemp     float64 // °C

	// Neighborhood size
	K    int // top‑K neighbors used for final estimate
	MinK int // ensure at least MinK are considered even if weights are tiny

	// Anchor behavior
	AnchorEpsilon float64 // satisfaction band around 50 considered "near‑perfect"
	AnchorBoost   float64 // multiplicative weight boost for anchors
	AnchorBlend   float64 // 0..1, how much anchor‑only estimate pulls the result

	// Recency behavior
	RecencyHalfLifeDays float64 // exponential half‑life for time decay

	// Source balance
	UserBoost float64 // multiplier applied to *user* records

	// Safety
	StepCapFraction float64 // e.g., 0.35 => limit change vs last user record to ±35%
	MinMinutes      float64
	MaxMinutes      float64

	// Risk policy
	NeverCold bool // if true, ceil at the end; else round to nearest
}

// NewPredictionServiceV2 with sensible defaults.
func NewPredictionServiceV2(recordService RecordServiceInterface, cfg *PredictionConfigV2) *PredictionServiceV2 {
	defaultCfg := PredictionConfigV2{
		SigmaDuration:       4.0,   // Std-dev for Gaussian weighting on shower duration (min) — smaller = more sensitive to duration similarity.
		SigmaTemp:           3.0,   // Std-dev for Gaussian weighting on ambient temperature (°C) — smaller = more sensitive to temperature similarity.
		K:                   25,    // Number of nearest neighbors (records) to consider from history (user + global).
		MinK:                6,     // Minimum number of records required for a prediction — ensures stability when history is sparse.
		RecencyHalfLifeDays: 5.0,   // Weight decay half-life in days — newer feedback counts more, halves in influence every N days.
		AnchorBlend:         0.35,  // Blend ratio between nearest-neighbor average and “perfect anchor” values — higher = perfects pull prediction more strongly.
		UserBoost:           2,     // Multiplier for weights from the current user’s history — increases personalisation over global data.
		StepCapFraction:     0.35,  // Max fractional change (vs. previous prediction) allowed in one step — smooths large jumps.
		MinMinutes:          5,     // Lower bound for predicted heating time (minutes) — safety/clamping.
		MaxMinutes:          120,   // Upper bound for predicted heating time (minutes) — safety/clamping.
		NeverCold:           false, // If true, bias rounding upward to avoid under-heating (“cold” risk).
	}

	if cfg != nil {
		// override defaults with provided values
		if cfg.SigmaDuration > 0 {
			defaultCfg.SigmaDuration = cfg.SigmaDuration
		}
		if cfg.SigmaTemp > 0 {
			defaultCfg.SigmaTemp = cfg.SigmaTemp
		}
		if cfg.K > 0 {
			defaultCfg.K = cfg.K
		}
		if cfg.MinK > 0 {
			defaultCfg.MinK = cfg.MinK
		}
		if cfg.AnchorEpsilon > 0 {
			defaultCfg.AnchorEpsilon = cfg.AnchorEpsilon
		}
		if cfg.AnchorBoost > 0 {
			defaultCfg.AnchorBoost = cfg.AnchorBoost
		}
		if cfg.AnchorBlend >= 0 && cfg.AnchorBlend <= 1 {
			defaultCfg.AnchorBlend = cfg.AnchorBlend
		}
		if cfg.RecencyHalfLifeDays > 0 {
			defaultCfg.RecencyHalfLifeDays = cfg.RecencyHalfLifeDays
		}
		if cfg.UserBoost > 0 {
			defaultCfg.UserBoost = cfg.UserBoost
		}
		if cfg.StepCapFraction > 0 && cfg.StepCapFraction < 1 {
			defaultCfg.StepCapFraction = cfg.StepCapFraction
		}
		if cfg.MinMinutes > 0 {
			defaultCfg.MinMinutes = cfg.MinMinutes
		}
		if cfg.MaxMinutes > 0 {
			defaultCfg.MaxMinutes = cfg.MaxMinutes
		}
		defaultCfg.NeverCold = cfg.NeverCold
	}
	return &PredictionServiceV2{
		recordService: recordService,
		cfg:           defaultCfg,
	}
}

// Predict computes the recommended heating time using Gaussian‑kNN with anchors.
func (s *PredictionServiceV2) Predict(req PredictionRequest) (*PredictionResponse, error) {
	// 1) Fetch data
	userRecords, err := s.recordService.GetRecordsForPredictionByUser(req.UserID, 400)
	if err != nil {
		return nil, err
	}
	globalRecords, err := s.recordService.GetGlobalRecordsForPrediction(req.UserID, 1200)
	if err != nil {
		return nil, err
	}

	// 2) Combine into a single slice with source flag
	all := make([]recWrap, 0, len(userRecords)+len(globalRecords))
	for _, r := range userRecords {
		all = append(all, recWrap{rec: r, isUser: true})
	}
	for _, r := range globalRecords {
		all = append(all, recWrap{rec: r, isUser: false})
	}
	if len(all) == 0 {
		// No data at all — conservative default of 30 minutes
		out := 30.0
		if s.cfg.NeverCold {
			out = math.Ceil(out)
		} else {
			out = math.Round(out)
		}
		return &PredictionResponse{HeatingTime: clamp(out, s.cfg.MinMinutes, s.cfg.MaxMinutes)}, nil
	}

	// 3) Precompute cell frequencies to avoid O(n²) scans
	cellCounts := make(map[string]int, len(all))
	for i := range all {
		key := freqCellKey(all[i].rec)
		all[i].cellKey = key
		cellCounts[key]++
	}

	// 4) Compute weights
	now := time.Now().UTC()
	for i := range all {
		r := &all[i]
		// Gaussian distance on duration & temperature
		wDur := gaussian(req.Duration-r.rec.ShowerDuration, s.cfg.SigmaDuration)
		wTmp := gaussian(req.Temperature-r.rec.AverageTemperature, s.cfg.SigmaTemp)
		w := wDur * wTmp

		// Recency decay
		days := math.Abs(now.Sub(r.rec.Date).Hours()) / 24.0
		w *= expHalfLife(days, s.cfg.RecencyHalfLifeDays)

		// Anchor boost on BOTH sides near 50
		if math.Abs(r.rec.Satisfaction-50.0) <= s.cfg.AnchorEpsilon {
			w *= s.cfg.AnchorBoost
			r.anchor = true
		}

		// Reliability: softly down‑weight very poor outcomes (wide sigma so it never hits 0)
		w *= gaussian(r.rec.Satisfaction-50.0, 22.0)

		// Cell frequency dampening: repeated contexts shouldn't dominate
		if cnt := cellCounts[r.cellKey]; cnt > 1 {
			w *= 1.0 / math.Sqrt(float64(cnt))
		}

		// Source balance
		if r.isUser {
			w *= s.cfg.UserBoost
		}

		r.weight = w
	}

	// 5) Select top‑K by weight (keep at least MinK)
	sort.Slice(all, func(i, j int) bool { return all[i].weight > all[j].weight })
	k := s.cfg.K
	if k < s.cfg.MinK {
		k = s.cfg.MinK
	}
	if k > len(all) {
		k = len(all)
	}
	top := all[:k]

	// 6) Weighted estimate using implied targets (all) + anchor‑only estimate (if anchors exist)
	estAll := weightedMeanTargets(top)
	estAnchors, anchorWeightSum := weightedMeanTargetsAnchors(top)

	// Blend toward anchors proportionally to their weight presence
	if anchorWeightSum > 0 {
		alpha := s.cfg.AnchorBlend * math.Min(1.0, anchorWeightSum/(sumWeights(top)+1e-9))
		estAll = (1.0-alpha)*estAll + alpha*estAnchors
	}

	// 7) Safety clamp vs last similar user record (context‑aware) to avoid big jumps
	if last, ok := latestSimilarUserRecord(userRecords, req, s.cfg.SigmaDuration*2.0, s.cfg.SigmaTemp*2.0); ok {
		capFrac := s.cfg.StepCapFraction
		minStep := last.HeatingTime * (1.0 - capFrac)
		maxStep := last.HeatingTime * (1.0 + capFrac)
		estAll = clamp(estAll, minStep, maxStep)
	}

	// 8) Absolute bounds and smart rounding (avoid 48.0x → ceil → 49 loop when feedback is hot)
	estAll = clamp(estAll, s.cfg.MinMinutes, s.cfg.MaxMinutes)
	if lastSat, ok := lastUserFeedback(userRecords); ok {
		estAll = smartRound(estAll, lastSat)
	} else {
		estAll = math.Round(estAll)
	}

	return &PredictionResponse{HeatingTime: estAll}, nil
}

// ------------- helpers --------------

func gaussian(delta, sigma float64) float64 {
	if sigma <= 0 {
		return 0
	}
	x := delta / sigma
	return math.Exp(-0.5 * x * x)
}

func expHalfLife(days, halfLife float64) float64 {
	if halfLife <= 0 {
		return 1.0
	}
	// exp(-ln2 * days / halfLife)
	return math.Exp(-math.Ln2 * days / halfLife)
}

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

func freqCellKey(r models.DailyRecord) string {
	d := int(math.Round(r.ShowerDuration))
	t := int(math.Round(r.AverageTemperature))
	return fmt.Sprintf("%d|%d", d, t)
}

func latestUserRecord(userRecs []models.DailyRecord) (models.DailyRecord, bool) {
	if len(userRecs) == 0 {
		return models.DailyRecord{}, false
	}
	latest := userRecs[0]
	for _, r := range userRecs[1:] {
		if r.Date.After(latest.Date) {
			latest = r
		}
	}
	return latest, true
}

func weightedMean(recs []recWrap) float64 {
	totalW := 0.0
	sum := 0.0
	for _, r := range recs {
		if r.weight <= 0 {
			continue
		}
		sum += r.rec.HeatingTime * r.weight
		totalW += r.weight
	}
	if totalW == 0 {
		return 30.0
	}
	return sum / totalW
}

func weightedMeanAnchors(recs []recWrap) (mean float64, weightSum float64) {
	totalW := 0.0
	sum := 0.0
	for _, r := range recs {
		if !r.anchor || r.weight <= 0 {
			continue
		}
		sum += r.rec.HeatingTime * r.weight
		totalW += r.weight
	}
	if totalW == 0 {
		return 0, 0
	}
	return sum / totalW, totalW
}

func sumWeights(recs []recWrap) float64 {
	total := 0.0
	for _, r := range recs {
		total += r.weight
	}
	return total
}

// --- v2 learning helpers: implied target and context-aware clamp ---

// impliedTarget converts a historical record into an implied target time based on satisfaction feedback.
// - Satisfaction ~50 -> keep the same time
// - Satisfaction >50 (too hot) -> reduce time with graduated percentages
// - Satisfaction <50 (too cold) -> increase time proportionally to severity with mild overshoot
func impliedTarget(r models.DailyRecord) float64 {
	s := r.Satisfaction
	h := r.HeatingTime

	// Near-perfect: tiny/no change
	if math.Abs(s-50.0) <= 1.0 {
		return h
	}

	if s > 50.0 {
		// Graduated reductions similar to v1 behavior
		switch {
		case s >= 85:
			return h * 0.75
		case s >= 80:
			return h * 0.80
		case s >= 75:
			return h * 0.83
		case s >= 65:
			return h * 0.87
		case s >= 60:
			return h * 0.92
		case s >= 55:
			return h * 0.97
		default:
			// Slightly hot (50<s<55): small nudge
			return h * 0.99
		}
	}

	// Too cold: proportional increase based on severity, with mild overshoot for very cold
	coldSeverity := (50.0 - s) / 50.0 // 0..1
	if coldSeverity < 0 {
		coldSeverity = 0
	}
	// Base learning percent scales from 12% up to 40%
	basePercent := 0.12 + 0.28*coldSeverity
	// Mild overshoot up to +10% extra when extremely cold
	overshoot := 1.0 + 0.10*coldSeverity
	factor := 1.0 + basePercent*overshoot
	return h * factor
}

// weightedMeanTargets computes weighted mean over implied targets instead of raw times
func weightedMeanTargets(recs []recWrap) float64 {
	totalW := 0.0
	sum := 0.0
	for _, r := range recs {
		if r.weight <= 0 {
			continue
		}
		tgt := impliedTarget(r.rec)
		sum += tgt * r.weight
		totalW += r.weight
	}
	if totalW == 0 {
		return 30.0
	}
	return sum / totalW
}

// weightedMeanTargetsAnchors computes weighted mean of implied targets restricted to anchor records
func weightedMeanTargetsAnchors(recs []recWrap) (mean float64, weightSum float64) {
	totalW := 0.0
	sum := 0.0
	for _, r := range recs {
		if !r.anchor || r.weight <= 0 {
			continue
		}
		tgt := impliedTarget(r.rec)
		sum += tgt * r.weight
		totalW += r.weight
	}
	if totalW == 0 {
		return 0, 0
	}
	return sum / totalW, totalW
}

// latestSimilarUserRecord returns the latest user record close to the request context
func latestSimilarUserRecord(userRecs []models.DailyRecord, req PredictionRequest, maxDeltaDur, maxDeltaTemp float64) (models.DailyRecord, bool) {
	var (
		found  bool
		latest models.DailyRecord
	)
	for _, r := range userRecs {
		if math.Abs(r.ShowerDuration-req.Duration) > maxDeltaDur {
			continue
		}
		if math.Abs(r.AverageTemperature-req.Temperature) > maxDeltaTemp {
			continue
		}
		if !found || r.Date.After(latest.Date) {
			latest = r
			found = true
		}
	}
	return latest, found
}

// smartRound: bias safe, but avoid sticking on the upper minute when user said "too hot".
func smartRound(est float64, lastSat float64) float64 {
	frac := est - math.Floor(est)
	if lastSat > 50 && frac <= 0.25 { // recently hot -> allow snap-down if close
		return math.Floor(est)
	}
	if lastSat < 50 { // recently cold -> keep bias-to-hot
		return math.Ceil(est)
	}
	return math.Round(est) // near-perfect recently -> unbiased
}

// lastUserFeedback returns the most recent satisfaction for the user.
func lastUserFeedback(userRecs []models.DailyRecord) (float64, bool) {
	var latest models.DailyRecord
	found := false
	for _, r := range userRecs {
		if !found || r.Date.After(latest.Date) {
			latest = r
			found = true
		}
	}
	if !found {
		return 50, false
	}
	return latest.Satisfaction, true
}
