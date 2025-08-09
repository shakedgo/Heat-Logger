<template>
  <div class="history-list card">
    <div class="header">
      <h2>History</h2>
      <div class="actions">
        <button class="export-btn" @click="exportHistory">
          <font-awesome-icon icon="file-export" /> Export CSV
        </button>
        <button class="delete-all-btn" @click="handleDeleteAll">
          <font-awesome-icon icon="trash" /> Delete All
        </button>
      </div>
    </div>
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
              <div class="scale-marker cold">
                <font-awesome-icon icon="snowflake" />
              </div>
              <div class="scale-bar">
                <div class="satisfaction-indicator" :style="{ left: ((entry.satisfaction - 1) * 0.99) + '%' }"
                  :class="getSatisfactionClass(entry.satisfaction)">
                  <span class="indicator-value">{{ entry.satisfaction }}</span>
                </div>
              </div>
              <div class="scale-marker hot">
                <font-awesome-icon icon="fire" />
              </div>
            </div>
          </div>
        </div>
        <button class="delete-btn" @click="handleDelete(entry.id)" title="Delete record">
          <font-awesome-icon icon="times" />
        </button>
      </div>
    </div>
    <div v-else class="no-history">
      <div class="illustration">🧼</div>
      <div class="title">No history yet</div>
      <div class="hint">Fill the form and calculate to start building your smart schedule.</div>
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
  emits: ['delete', 'deleteAll'],
  computed: {
    sortedHistory() {
      if (!Array.isArray(this.history)) {
        console.warn('History is not an array:', this.history);
        return [];
      }
      const validEntries = this.history.filter(entry =>
        entry.id &&
        entry.date !== "0001-01-01T00:00:00Z" &&
        entry.satisfaction > 0
      );
      return validEntries;
    }
  },
  methods: {
    formatDate(dateString) {
      return new Date(dateString).toLocaleDateString();
    },
    getSatisfactionClass(satisfaction) {
      if (satisfaction < 40) return 'cold';
      if (satisfaction > 60) return 'hot';
      return 'perfect';
    },
    async handleDelete(id) {
      try {
        await this.$confirm({ title: 'Delete record', message: 'Are you sure you want to delete this record?' })
        this.$emit('delete', id)
        this.$toast('Record deleted', { type: 'success' })
      } catch (_) {}
    },
    async handleDeleteAll() {
      try {
        await this.$confirm({ title: 'Delete all', message: 'Delete all records? This cannot be undone.' })
        this.$emit('deleteAll')
        this.$toast('All records deleted', { type: 'success' })
      } catch (_) {}
    },
    async exportHistory() {
      try {
        const response = await this.$api.get('/history/export', {
          responseType: 'blob',
          headers: {
            'Accept': 'text/csv'
          }
        });

        const blob = new Blob([response.data], { type: 'text/csv' });
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `heating_history_${new Date().toLocaleDateString()}.csv`;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
      } catch (error) {
        console.error('Failed to export history:', error);
        alert('Failed to export history. Please try again.');
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
  padding: 24px;
  border-radius: 16px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .actions {
    display: flex;
    gap: 1rem;
  }

   .export-btn {
    background-color: #3b82f6;
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 10px;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .export-btn:hover {
    background-color: #2563eb;
  }

  .delete-all-btn {
    background-color: #ef4444;
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 10px;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .delete-all-btn:hover {
    background-color: #dc2626;
  }

  .history-entries {
    & {
      max-height: 520px;
      overflow-y: auto;
      padding-right: 10px;
    }

    .history-entry {
      & {
        padding: 15px;
        border-bottom: 1px solid #eee;
        display: flex;
        align-items: flex-start;
        position: relative;
      }

      &:last-child {
        border-bottom: none;
      }

      .entry-date { min-width: 120px; color: var(--muted); font-variant-numeric: tabular-nums; }

      .entry-details {
        & {
          flex-grow: 1;
          margin-left: 20px;
        }

        .entry-stats {
          & {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 8px;
            margin-bottom: 10px;
          }

          >div { font-size: 0.95em; color: var(--text); white-space: nowrap; }

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
         width: 28px;
         height: 28px;
        padding: 0;
         border-radius: 999px;
         background: #ef4444;
        color: white;
        border: none;
        font-size: 14px;
        cursor: pointer;
        opacity: 0;
        transition: opacity 0.2s, transform .08s ease;
        display: flex;
        align-items: center;
        justify-content: center;

        &:hover {
         background: #dc2626;
         transform: translateY(-1px);
        }
      }

      &:hover {
        .delete-btn {
          opacity: 1;
        }
      }
    }
  }

  .no-history {
    text-align: center;
    color: #666;
    padding: 32px 20px;
    display: grid;
    gap: 6px;
    .illustration { font-size: 28px; }
    .title { font-weight: 600; color: #1f2937; }
    .hint { color: #6b7280; font-size: 14px; }
  }
}

[data-theme='dark'] .history-list {
  background: #1f1f1f;
  color: #e4e4e7;

  .export-btn { background-color: #3b82f6; }
  .export-btn:hover { background-color: #2563eb; }
  .delete-all-btn { background-color: #dc2626; }
  .delete-all-btn:hover { background-color: #b91c1c; }

  .history-entry { border-bottom-color: rgba(161, 161, 170, 0.18); }
  .entry-stats > div { color: #e5e7eb; }
  .delete-btn { background: #dc2626; }

  .no-history {
    color: #cbd5e1;
    .title { color: #e5e7eb; }
    .hint { color: #a1a1aa; }
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
  background: linear-gradient(to right,
      #00b4d8, // cold
      #90e0ef, // cool
      #caf0f8, // perfect
      #ffba08, // warm
      #dc2f02 // hot
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