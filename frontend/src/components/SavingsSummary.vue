<template>
  <div class="savings-summary card" v-if="history.length > 0">
    <div class="summary-header">
      <h3>Savings Summary</h3>
      <button class="baseline-toggle" @click="showBaselineInput = !showBaselineInput">
        vs. baseline of <strong>{{ baseline }} min</strong>
        <font-awesome-icon icon="sliders" />
      </button>
    </div>

    <div class="baseline-config" v-if="showBaselineInput">
      <div class="config-row">
        <label for="baseline-input">Baseline (min):</label>
        <input id="baseline-input" type="number" :value="baseline" @input="updateBaseline($event.target.value)" min="1" step="1">
      </div>
      <div class="config-row">
        <label for="heater-power-input">Heater power (kW):</label>
        <input id="heater-power-input" type="number" :value="heaterPowerKW" @input="updateHeaterPower($event.target.value)" min="0.1" step="0.1">
      </div>
      <div class="config-row">
        <label for="price-input">Price (₪/kWh):</label>
        <input id="price-input" type="number" :value="pricePerKwh" @input="updatePricePerKwh($event.target.value)" min="0.01" step="0.01">
      </div>
    </div>

    <div class="stats-row">
      <div class="stat-box week">
        <div class="stat-label">Last 7 Days</div>
        <template v-if="summary.week.count > 0">
          <div class="stat-value">{{ summary.week.totalMinutesSaved }} <span class="stat-unit">min saved</span></div>
          <div class="stat-sub">{{ summary.week.count }} qualifying · avg satisfaction <strong>{{ summary.week.averageSatisfaction }}</strong></div>
          <div class="stat-energy">~{{ weekEnergy }} kWh · <strong>~₪{{ weekCost }}</strong></div>
        </template>
        <div v-else class="stat-empty">No data yet</div>
      </div>
      <div class="stat-box month">
        <div class="stat-label">Last 30 Days</div>
        <template v-if="summary.month.count > 0">
          <div class="stat-value">{{ summary.month.totalMinutesSaved }} <span class="stat-unit">min saved</span></div>
          <div class="stat-sub">{{ summary.month.count }} qualifying · avg satisfaction <strong>{{ summary.month.averageSatisfaction }}</strong></div>
          <div class="stat-energy">~{{ monthEnergy }} kWh · <strong>~₪{{ monthCost }}</strong></div>
        </template>
        <div v-else class="stat-empty">No data yet</div>
      </div>
    </div>

    <div class="summary-footer">
      Only counting predictions rated near-perfect ({{ nearPerfectMin }}–{{ nearPerfectMax }})
      <br>Energy and cost figures are estimates based on your configured heater power and electricity price
    </div>
  </div>
</template>

<script>
import { computeSavingsSummary, computeEnergySavings, computeCostSavings, NEAR_PERFECT_MIN, NEAR_PERFECT_MAX, DEFAULT_BASELINE, DEFAULT_HEATER_POWER_KW, DEFAULT_PRICE_PER_KWH } from '../utils/savingsCalculator.js'

export default {
  name: 'SavingsSummary',
  props: {
    history: {
      type: Array,
      default: () => []
    }
  },
  data() {
    const stored = localStorage.getItem('heatLogger_savingsBaseline')
    const storedPower = localStorage.getItem('heatLogger_heaterPowerKW')
    const storedPrice = localStorage.getItem('heatLogger_pricePerKwh')
    return {
      baseline: stored ? parseInt(stored, 10) : DEFAULT_BASELINE,
      heaterPowerKW: storedPower ? parseFloat(storedPower) : DEFAULT_HEATER_POWER_KW,
      pricePerKwh: storedPrice ? parseFloat(storedPrice) : DEFAULT_PRICE_PER_KWH,
      showBaselineInput: false,
      nearPerfectMin: NEAR_PERFECT_MIN,
      nearPerfectMax: NEAR_PERFECT_MAX
    }
  },
  computed: {
    summary() {
      return computeSavingsSummary(this.history, this.baseline)
    },
    weekEnergy() {
      return computeEnergySavings(this.summary.week.totalMinutesSaved, this.heaterPowerKW)
    },
    monthEnergy() {
      return computeEnergySavings(this.summary.month.totalMinutesSaved, this.heaterPowerKW)
    },
    weekCost() {
      return computeCostSavings(this.weekEnergy, this.pricePerKwh)
    },
    monthCost() {
      return computeCostSavings(this.monthEnergy, this.pricePerKwh)
    }
  },
  methods: {
    updateBaseline(value) {
      const parsed = parseInt(value, 10)
      if (!isNaN(parsed) && parsed > 0) {
        this.baseline = parsed
        localStorage.setItem('heatLogger_savingsBaseline', String(parsed))
      }
    },
    updateHeaterPower(value) {
      const parsed = parseFloat(value)
      if (!isNaN(parsed) && parsed > 0) {
        this.heaterPowerKW = parsed
        localStorage.setItem('heatLogger_heaterPowerKW', String(parsed))
      }
    },
    updatePricePerKwh(value) {
      const parsed = parseFloat(value)
      if (!isNaN(parsed) && parsed > 0) {
        this.pricePerKwh = parsed
        localStorage.setItem('heatLogger_pricePerKwh', String(parsed))
      }
    }
  }
}
</script>

<style lang="scss" scoped>
.savings-summary {
  background: white;
  padding: 20px 24px;
  border-radius: 16px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  margin-bottom: 0;
}

.summary-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;

  h3 {
    margin: 0;
    font-size: 17px;
    color: var(--heading);
  }
}

.baseline-toggle {
  background: #f3f4f6;
  border: none;
  padding: 4px 10px;
  border-radius: 20px;
  font-size: 12px;
  color: #6b7280;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;

  strong { color: #4f46e5; }

  &:hover { background: #e5e7eb; }
}

.baseline-config {
  background: #f9fafb;
  border-radius: 10px;
  padding: 10px 14px;
  margin-bottom: 14px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  font-size: 13px;
  color: #4b5563;

  .config-row {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  input {
    width: 70px;
    padding: 4px 8px;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    font-size: 13px;
    text-align: center;
  }
}

.stats-row {
  display: flex;
  gap: 14px;
  margin-bottom: 12px;
}

.stat-box {
  flex: 1;
  border-radius: 12px;
  padding: 16px;
  text-align: center;

  &.week { background: linear-gradient(135deg, #eef2ff, #e0e7ff); }
  &.month { background: linear-gradient(135deg, #f0fdf4, #dcfce7); }
}

.stat-label {
  font-size: 11px;
  color: #6b7280;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;

  .week & { color: #4f46e5; }
  .month & { color: #16a34a; }
}

.stat-unit {
  font-size: 13px;
  font-weight: 400;
  color: #6b7280;
}

.stat-sub {
  margin-top: 6px;
  font-size: 11px;
  color: #9ca3af;

  strong { color: #6b7280; }
}

.stat-energy {
  margin-top: 6px;
  font-size: 12px;
  color: #6b7280;

  strong {
    .week & { color: #4f46e5; }
    .month & { color: #16a34a; }
  }
}

.stat-empty {
  color: #9ca3af;
  font-size: 14px;
  padding: 8px 0;
}

.summary-footer {
  text-align: center;
  font-size: 11px;
  color: #9ca3af;
}

/* Dark mode */
[data-theme='dark'] .savings-summary {
  background: #1f1f1f;

  .summary-header h3 { color: #e7e7ea; }

  .baseline-toggle {
    background: #2a2a2a;
    color: #a1a1aa;
    strong { color: #818cf8; }
    &:hover { background: #333; }
  }

  .baseline-config {
    background: #2a2a2a;
    color: #cbd5e1;
    input {
      background: #1f1f1f;
      color: #e5e7eb;
      border-color: #3a3a3a;
    }
  }

  .stat-box.week { background: linear-gradient(135deg, #1e1b4b, #272567); }
  .stat-box.month { background: linear-gradient(135deg, #052e16, #14532d); }

  .stat-value {
    .week & { color: #a5b4fc; }
    .month & { color: #4ade80; }
  }

  .stat-label { color: #a1a1aa; }
  .stat-unit { color: #a1a1aa; }
  .stat-sub { color: #71717a; strong { color: #a1a1aa; } }
  .stat-energy { color: #a1a1aa; strong { .week & { color: #a5b4fc; } .month & { color: #4ade80; } } }
  .stat-empty { color: #71717a; }
  .summary-footer { color: #71717a; }
}
</style>
