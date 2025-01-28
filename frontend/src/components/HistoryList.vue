<template>
  <div class="history-list">
    <h2>History</h2>
    <div class="history-entries" v-if="history.length > 0">
      <div v-for="entry in sortedHistory" :key="entry.id" class="history-entry">
        <div class="entry-date">
          {{ formatDate(entry.date) }}
        </div>
        <div class="entry-details">
          <div>Temperature: {{ entry.averageTemperature }}°C</div>
          <div>Duration: {{ entry.showerDuration }} min</div>
          <div>Satisfaction: {{ entry.satisfaction }}/10 
            <span class="satisfaction-text">({{ getSatisfactionText(entry.satisfaction) }})</span>
          </div>
        </div>
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
      return [...this.history].reverse().slice(0, 10);
    }
  },
  methods: {
    formatDate(dateString) {
      return new Date(dateString).toLocaleDateString();
    },
    getSatisfactionText(satisfaction) {
      if (satisfaction < 5) return 'too cold';
      if (satisfaction > 5) return 'too hot';
      return 'perfect';
    }
  }
}
</script>

<style scoped>
.history-list {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

h2 {
  margin-top: 0;
  color: #2c3e50;
}

.history-entries {
  max-height: 400px;
  overflow-y: auto;
}

.history-entry {
  padding: 15px;
  border-bottom: 1px solid #eee;
  display: flex;
  align-items: flex-start;
}

.history-entry:last-child {
  border-bottom: none;
}

.entry-date {
  min-width: 100px;
  color: #666;
}

.entry-details {
  flex-grow: 1;
  margin-left: 20px;
}

.entry-details > div {
  margin-bottom: 5px;
}

.satisfaction-text {
  color: #666;
  font-size: 0.9em;
}

.no-history {
  text-align: center;
  color: #666;
  padding: 20px;
}
</style> 