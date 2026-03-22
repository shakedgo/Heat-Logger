import assert from 'node:assert/strict'
import { describe, it } from 'node:test'
import {
  isNearPerfect,
  getRecordsInRange,
  computeSavings,
  computeSavingsSummary,
  computeEnergySavings,
  computeCostSavings,
  NEAR_PERFECT_MIN,
  NEAR_PERFECT_MAX,
  DEFAULT_BASELINE,
  DEFAULT_HEATER_POWER_KW,
  DEFAULT_PRICE_PER_KWH
} from '../frontend/src/utils/savingsCalculator.js'

function makeRecord(overrides = {}) {
  return {
    id: 1,
    date: new Date().toISOString(),
    heatingTime: 30,
    satisfaction: 50,
    showerDuration: 10,
    averageTemperature: 20,
    ...overrides
  }
}

function daysAgo(n) {
  return new Date(Date.now() - n * 24 * 60 * 60 * 1000).toISOString()
}

describe('isNearPerfect', () => {
  it('returns true for satisfaction within [45, 55]', () => {
    assert.ok(isNearPerfect(45))
    assert.ok(isNearPerfect(50))
    assert.ok(isNearPerfect(55))
  })

  it('returns false for satisfaction outside [45, 55]', () => {
    assert.ok(!isNearPerfect(44))
    assert.ok(!isNearPerfect(56))
    assert.ok(!isNearPerfect(1))
    assert.ok(!isNearPerfect(100))
  })
})

describe('computeSavings', () => {
  it('returns correct savings when all records are near-perfect', () => {
    const records = [
      makeRecord({ heatingTime: 30, satisfaction: 50 }),
      makeRecord({ heatingTime: 35, satisfaction: 48 }),
      makeRecord({ heatingTime: 40, satisfaction: 52 })
    ]
    const result = computeSavings(records, 45)
    // (45-30) + (45-35) + (45-40) = 15 + 10 + 5 = 30
    assert.equal(result.totalMinutesSaved, 30)
    assert.equal(result.count, 3)
    assert.equal(result.averageSatisfaction, 50)
  })

  it('returns 0 when all predictions exceeded the baseline', () => {
    const records = [
      makeRecord({ heatingTime: 50, satisfaction: 50 }),
      makeRecord({ heatingTime: 60, satisfaction: 48 })
    ]
    const result = computeSavings(records, 45)
    assert.equal(result.totalMinutesSaved, 0)
    assert.equal(result.count, 0)
    assert.equal(result.averageSatisfaction, null)
  })

  it('excludes records outside the near-perfect satisfaction range', () => {
    const records = [
      makeRecord({ heatingTime: 30, satisfaction: 50 }), // included
      makeRecord({ heatingTime: 30, satisfaction: 20 }), // excluded (too low)
      makeRecord({ heatingTime: 30, satisfaction: 80 })  // excluded (too high)
    ]
    const result = computeSavings(records, 45)
    assert.equal(result.totalMinutesSaved, 15)
    assert.equal(result.count, 1)
    assert.equal(result.averageSatisfaction, 50)
  })

  it('handles empty records', () => {
    const result = computeSavings([], 45)
    assert.equal(result.totalMinutesSaved, 0)
    assert.equal(result.count, 0)
    assert.equal(result.averageSatisfaction, null)
  })
})

describe('getRecordsInRange', () => {
  it('correctly assigns records to week and month buckets', () => {
    const records = [
      makeRecord({ date: daysAgo(2) }),  // in week and month
      makeRecord({ date: daysAgo(10) }), // in month only
      makeRecord({ date: daysAgo(35) })  // in neither
    ]

    const weekRecords = getRecordsInRange(records, 7)
    assert.equal(weekRecords.length, 1)

    const monthRecords = getRecordsInRange(records, 30)
    assert.equal(monthRecords.length, 2)
  })
})

describe('computeSavingsSummary', () => {
  it('falls back to default baseline when none provided', () => {
    const records = [makeRecord({ heatingTime: 30, satisfaction: 50, date: daysAgo(1) })]
    const result = computeSavingsSummary(records, undefined)
    assert.equal(result.baseline, DEFAULT_BASELINE)
    assert.equal(result.week.totalMinutesSaved, DEFAULT_BASELINE - 30)
  })

  it('returns zero savings for empty history', () => {
    const result = computeSavingsSummary([], 45)
    assert.equal(result.week.totalMinutesSaved, 0)
    assert.equal(result.week.count, 0)
    assert.equal(result.month.totalMinutesSaved, 0)
    assert.equal(result.month.count, 0)
  })

  it('separates week and month correctly', () => {
    const records = [
      makeRecord({ heatingTime: 30, satisfaction: 50, date: daysAgo(3) }),  // week + month
      makeRecord({ heatingTime: 35, satisfaction: 50, date: daysAgo(15) })  // month only
    ]
    const result = computeSavingsSummary(records, 45)
    assert.equal(result.week.totalMinutesSaved, 15)   // 45-30
    assert.equal(result.week.count, 1)
    assert.equal(result.month.totalMinutesSaved, 25)   // (45-30)+(45-35)
    assert.equal(result.month.count, 2)
  })
})

describe('computeEnergySavings', () => {
  it('calculates kWh from minutes saved and heater power', () => {
    // 30 min saved × 3 kW / 60 = 1.5 kWh
    assert.equal(computeEnergySavings(30, 3), 1.5)
  })

  it('uses default heater power when none provided', () => {
    // 60 min × 3 kW (default) / 60 = 3 kWh
    assert.equal(computeEnergySavings(60, undefined), 60 * DEFAULT_HEATER_POWER_KW / 60)
  })

  it('returns 0 for 0 minutes saved', () => {
    assert.equal(computeEnergySavings(0, 3), 0)
  })
})

describe('computeCostSavings', () => {
  it('calculates cost from kWh and price', () => {
    // 1.5 kWh × 0.67 ILS = 1.005 → rounded to 1.01
    assert.equal(computeCostSavings(1.5, 0.67), 1.01)
  })

  it('uses default price when none provided', () => {
    assert.equal(computeCostSavings(1, undefined), DEFAULT_PRICE_PER_KWH)
  })

  it('returns 0 for 0 kWh', () => {
    assert.equal(computeCostSavings(0, 0.67), 0)
  })
})
