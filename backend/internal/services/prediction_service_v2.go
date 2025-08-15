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
		SigmaDuration:       6.0,
		SigmaTemp:           4.0,
		K:                   60,
		MinK:                8,
		AnchorEpsilon:       5.0,
		AnchorBoost:         1.6,
		AnchorBlend:         0.35,
		RecencyHalfLifeDays: 10.0,
		UserBoost:           1.4,
		StepCapFraction:     0.35,
		MinMinutes:          5.0,
		MaxMinutes:          120.0,
		NeverCold:           true,
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

	// 6) Weighted estimate (all) + anchor‑only estimate (if anchors exist)
	estAll := weightedMean(top)
	estAnchors, anchorWeightSum := weightedMeanAnchors(top)

	// Blend toward anchors proportionally to their weight presence
	if anchorWeightSum > 0 {
		alpha := s.cfg.AnchorBlend * math.Min(1.0, anchorWeightSum/(sumWeights(top)+1e-9))
		estAll = (1.0-alpha)*estAll + alpha*estAnchors
	}

	// 7) Safety clamp vs last user record to avoid big jumps
	if last, ok := latestUserRecord(userRecords); ok {
		capFrac := s.cfg.StepCapFraction
		minStep := last.HeatingTime * (1.0 - capFrac)
		maxStep := last.HeatingTime * (1.0 + capFrac)
		estAll = clamp(estAll, minStep, maxStep)
	}

	// 8) Absolute bounds and rounding policy
	estAll = clamp(estAll, s.cfg.MinMinutes, s.cfg.MaxMinutes)
	if s.cfg.NeverCold {
		estAll = math.Ceil(estAll)
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
