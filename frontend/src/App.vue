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
        if (error.response && error.response.data && error.response.data.error) {
          alert(`Error: ${error.response.data.error}`);
        } else {
          alert('An error occurred while calculating. Please try again.');
        }
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
        if (error.response && error.response.data && error.response.data.error) {
          alert(`Error: ${error.response.data.error}`);
        } else {
          alert('An error occurred while saving feedback. Please try again.');
        }
      }
    },
    async loadHistory() {
      try {
        const response = await this.$api.get('/history');
        console.log('Loaded history:', response.data);
        this.history = response.data.history;
      } catch (error) {
        console.error('Error loading history:', error);
      }
    },
    async handleDelete(id) {
      try {
        const response = await this.$api.post('/history/delete', { id });
        if (response.status === 200) {
          await this.loadHistory();
        } else {
          throw new Error('Failed to delete record');
        }
      } catch (error) {
        console.error('Error deleting record:', error);
        alert('Failed to delete record. Please try again.');
      }
    },
    async handleDeleteAll() {
      try {
        const response = await this.$api.post('/history/deleteall');
        if (response.status === 200) {
          await this.loadHistory();
        } else {
          throw new Error('Failed to delete all records');
        }
      } catch (error) {
        console.error('Error deleting all records:', error);
        alert('Failed to delete all records. Please try again.');
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