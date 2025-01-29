<template>
  <div class="history-list">
    <h2>History</h2>
    <div class="history-entries" v-if="history.length > 0">
      <div v-for="entry in sortedHistory" :key="entry.id" class="history-entry">
        <div class="entry-date">
          {{ formatDate(entry.date) }}
        </div>
        <div class="entry-details">
          <div class="entry-stats">
            <div>Temperature: {{ entry.averageTemperature }}°C</div>
            <div>Duration: {{ entry.showerDuration }} min</div>
            <div>Heating Time: <span class="heating-time">{{ entry.heatingTime.toFixed(1) }} min</span></div>
          </div>
          <div class="satisfaction-bar">
            <div class="satisfaction-label">Satisfaction:</div>
            <div class="satisfaction-scale">
              <div class="scale-marker cold">❄️</div>
              <div class="scale-bar">
                <div 
                  class="satisfaction-indicator"
                  :style="{ left: ((entry.satisfaction - 1) * 11.11) + '%' }"
                  :class="getSatisfactionClass(entry.satisfaction)"
                >
                  <span class="indicator-value">{{ entry.satisfaction }}</span>
                </div>
              </div>
              <div class="scale-marker hot">🔥</div>
            </div>
          </div>
        </div>
        <button class="delete-btn" @click="handleDelete(entry.id)" title="Delete record">
          ×
        </button>
      </div>
    </div>
    <div v-else class="no-history">
      No history available yet.
    </div>
  </div>
</template>

<script>
export default {
  name: 'HistoryList',
  props: {
    history: {
      type: Array,
      default: () => []
    }
  },
  computed: {
    sortedHistory() {
      if (!Array.isArray(this.history)) {
        console.warn('History is not an array:', this.history);
        return [];
      }
      // Filter out invalid entries and sort
      const validEntries = this.history.filter(entry => 
        entry.id && 
        entry.date !== "0001-01-01T00:00:00Z" &&
        entry.satisfaction > 0
      );
      return validEntries.reverse().slice(0, 10);
    }
  },
  methods: {
    formatDate(dateString) {
      return new Date(dateString).toLocaleDateString();
    },
    getSatisfactionClass(satisfaction) {
      if (satisfaction < 4) return 'cold';
      if (satisfaction > 6) return 'hot';
      return 'perfect';
    },
    handleDelete(id) {
      if (confirm('Are you sure you want to delete this record?')) {
        this.$emit('delete', id);
      }
    }
  },
  mounted() {
    console.log('HistoryList mounted, entries:', this.history.length);
  }
}
</script>

<style lang="scss" scoped>
.history-list {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);

  h2 {
    margin-top: 0;
    color: #2c3e50;
  }

  .history-entries {
    max-height: 400px;
    overflow-y: auto;
    padding-right: 10px;
  }

  .no-history {
    text-align: center;
    color: #666;
    padding: 20px;
  }
}

.history-entry {
  padding: 15px;
  border-bottom: 1px solid #eee;
  display: flex;
  align-items: flex-start;
  position: relative;

  &:last-child {
    border-bottom: none;
  }

  .entry-date {
    min-width: 100px;
    color: #666;
  }

  .entry-details {
    flex-grow: 1;
    margin-left: 20px;

    .entry-stats {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
      gap: 8px;
      margin-bottom: 10px;
      
      > div {
        font-size: 0.95em;
        color: #2c3e50;
        white-space: nowrap;
      }

      .heating-time {
        color: #42b983;
        font-weight: 500;
      }
    }
  }

  .delete-btn {
    position: absolute;
    top: 10px;
    right: 10px;
    width: 24px;
    height: 24px;
    padding: 0;
    border-radius: 50%;
    background: #dc3545;
    color: white;
    border: none;
    font-size: 18px;
    line-height: 1;
    cursor: pointer;
    opacity: 0;
    transition: opacity 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;

    &:hover {
      background: #c82333;
    }
  }

  &:hover {
    .delete-btn {
      opacity: 1;
    }
  }
}

.satisfaction-bar {
  margin-top: 10px;

  .satisfaction-label {
    font-weight: 500;
    margin-bottom: 4px;
  }

  .satisfaction-scale {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 24px;
  }
}

.scale-bar {
  flex-grow: 1;
  height: 8px;
  background: linear-gradient(
    to right,
    #00b4d8,  // cold
    #90e0ef,  // cool
    #caf0f8,  // perfect
    #ffba08,  // warm
    #dc2f02   // hot
  );
  border-radius: 4px;
  position: relative;
}

.satisfaction-indicator {
  position: absolute;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: white;
  border: 2px solid #333;
  top: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: bold;
  cursor: help;

  .indicator-value {
    color: #333;
  }

  &.cold {
    border-color: #00b4d8;
  }

  &.perfect {
    border-color: #90e0ef;
  }

  &.hot {
    border-color: #dc2f02;
  }
}

.scale-marker {
  font-size: 16px;

  &.cold {
    color: #00b4d8;
  }

  &.hot {
    color: #dc2f02;
  }
}
</style> 