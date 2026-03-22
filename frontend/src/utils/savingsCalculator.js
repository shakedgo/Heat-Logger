// Derived from v2 prediction algorithm: AnchorEpsilon = 5.0
// "Near-perfect" = satisfaction within 5 of 50 → [45, 55]
export const NEAR_PERFECT_MIN = 45
export const NEAR_PERFECT_MAX = 55
export const DEFAULT_BASELINE = 45
export const DEFAULT_HEATER_POWER_KW = 3
export const DEFAULT_PRICE_PER_KWH = 0.67 // ILS, Israel 2026

export function isNearPerfect(satisfaction) {
  return satisfaction >= NEAR_PERFECT_MIN && satisfaction <= NEAR_PERFECT_MAX
}

export function getRecordsInRange(records, days) {
  const cutoff = Date.now() - days * 24 * 60 * 60 * 1000
  return records.filter(r => new Date(r.date).getTime() >= cutoff)
}

export function computeSavings(records, baselineMinutes) {
  const qualifying = records.filter(
    r => isNearPerfect(r.satisfaction) && r.heatingTime < baselineMinutes
  )

  if (qualifying.length === 0) {
    return { totalMinutesSaved: 0, count: 0, averageSatisfaction: null }
  }

  const totalMinutesSaved = qualifying.reduce(
    (sum, r) => sum + (baselineMinutes - r.heatingTime),
    0
  )

  const averageSatisfaction =
    qualifying.reduce((sum, r) => sum + r.satisfaction, 0) / qualifying.length

  return {
    totalMinutesSaved: Math.round(totalMinutesSaved * 10) / 10,
    count: qualifying.length,
    averageSatisfaction: Math.round(averageSatisfaction)
  }
}

export function computeEnergySavings(minutesSaved, heaterPowerKW) {
  const kw = heaterPowerKW ?? DEFAULT_HEATER_POWER_KW
  return Math.round((minutesSaved * kw / 60) * 100) / 100
}

export function computeCostSavings(kwhSaved, pricePerKwh) {
  const price = pricePerKwh ?? DEFAULT_PRICE_PER_KWH
  return Math.round(kwhSaved * price * 100) / 100
}

export function computeSavingsSummary(allRecords, baselineMinutes) {
  const baseline = baselineMinutes ?? DEFAULT_BASELINE
  const weekRecords = getRecordsInRange(allRecords, 7)
  const monthRecords = getRecordsInRange(allRecords, 30)

  return {
    week: computeSavings(weekRecords, baseline),
    month: computeSavings(monthRecords, baseline),
    baseline
  }
}
