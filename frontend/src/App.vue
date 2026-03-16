<template>
  <div class="container app-shell">
    <header class="app-header">
      <div class="brand">
        <span class="logo">🔥</span>
        <h1>Heat Logger</h1>
      </div>
      <p class="subtitle">Smarter water heater timings, based on you</p>
    </header>
    <div class="content">
      <InputForm 
        @calculate="handleCalculate" 
        @submitFeedback="handleSubmit"
        :latestHeatingTime="latestHeatingTime"
      />
      <div class="history-section">
        <HistoryList 
          :history="history"
          @delete="handleDelete"
          @deleteAll="handleDeleteAll"
        />
      </div>
      <UiToaster />
    </div>
  </div>
</template>

<script>
import InputForm from './components/InputForm.vue'
import HistoryList from './components/HistoryList.vue'
import UiToaster from './components/UiToaster.vue'

export default {
  name: 'App',
  components: {
    InputForm,
    HistoryList,
    UiToaster
  },
  data() {
    return {
      history: [],
      latestHeatingTime: null
    }
  },
  methods: {
    async handleCalculate(data) {
      try {
        console.log('Sending prediction request:', data);
        const response = await this.$api.post('/calculate', data);
        console.log('Received prediction response:', response.data);
        this.latestHeatingTime = response.data.heatingTime;
      } catch (error) {
        console.error('Error:', error);
        const msg = error.response?.data?.error || 'An error occurred while calculating. Please try again.';
        this.$toast(msg, { type: 'error' });
      }
    },
    async handleSubmit(data) {
      try {
        console.log('Sending feedback:', data);
        const response = await this.$api.post('/feedback', data);
        console.log('Feedback response:', response.data);
        if (response.status === 200) {
          await this.loadHistory();
          this.latestHeatingTime = null;
        } else {
          throw new Error('Failed to submit feedback');
        }
      } catch (error) {
        console.error('Error:', error);
        const msg = error.response?.data?.error || 'An error occurred while saving feedback. Please try again.';
        this.$toast(msg, { type: 'error' });
      }
    },
    async loadHistory() {
      try {
        const response = await this.$api.get('/history');
        console.log('Loaded history:', response.data);
        this.history = response.data.history;
      } catch (error) {
        console.error('Error loading history:', error);
        this.$toast('Failed to load history. Is the server running?', { type: 'error' });
      }
    },
    async handleDelete(id) {
      try {
        const response = await this.$api.post('/history/delete', { id });
        if (response.status !== 200) throw new Error('Failed to delete record');
        await this.loadHistory();
      } catch (error) {
        console.error('Error deleting record:', error);
        this.$toast('Failed to delete record. Please try again.', { type: 'error' });
      }
    },
    async handleDeleteAll() {
      try {
        const response = await this.$api.post('/history/deleteall');
        if (response.status !== 200) throw new Error('Failed to delete all records');
        await this.loadHistory();
      } catch (error) {
        console.error('Error deleting all records:', error);
        this.$toast('Failed to delete all records. Please try again.', { type: 'error' });
      }
    }
  },
  async created() {
    await this.loadHistory();
  }
}
</script>

<style lang="scss">
.app-shell {
  max-width: 1100px;
}

.app-header {
  & {
    position: sticky;
    top: 0;
    z-index: 50;
    text-align: center;
    margin-top: 12px;
    margin-bottom: 8px;
    padding-bottom: 8px;
    backdrop-filter: blur(6px);
  }

  .brand {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
  }

  .logo { font-size: 28px; }

  h1 { margin: 0; font-size: 28px; letter-spacing: 0.2px; color: var(--heading); }

  .subtitle { margin: 6px 0 0 0; color: var(--muted); }

  .brand > *:last-child { margin-left: 6px; }
}

.content {
  display: grid;
  gap: 20px;
  margin-top: 12px;
}

@media (min-width: 940px) {
  .content {
    grid-template-columns: 0.9fr 1.1fr; // form slightly narrower than history
    align-items: start;
  }
}

[data-theme='dark'] {
  body { background: #303134; color: #e4e4e7; }
  .app-header h1 { color: #e7e7ea; }
  .subtitle { color: #a1a1aa; }
}
</style> 