<template>
  <div class="container">
    <h1>Water Heater Learning System</h1>
    <div class="content">
      <InputForm 
        @calculate="handleCalculate" 
        @submitFeedback="handleSubmit"
        :latestHeatingTime="latestHeatingTime"
      />
      <LatestResult :heatingTime="latestHeatingTime" />
      <div class="history-section">
        <HistoryList 
          :history="history"
          @delete="handleDelete"
          @deleteAll="handleDeleteAll"
        />
      </div>
    </div>
  </div>
</template>

<script>
import InputForm from './components/InputForm.vue'
import LatestResult from './components/LatestResult.vue'
import HistoryList from './components/HistoryList.vue'

export default {
  name: 'App',
  components: {
    InputForm,
    LatestResult,
    HistoryList
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
        const response = await this.$api.post('/calculate', data);
        this.latestHeatingTime = response.data.heatingTime;
      } catch (error) {
        console.error('Error:', error);
        alert('An error occurred while calculating. Please try again.');
      }
    },
    async handleSubmit(data) {
      try {
        const response = await this.$api.post('/feedback', data);
        if (response.status === 200) {
          await this.loadHistory();
          this.latestHeatingTime = null;
        } else {
          throw new Error('Failed to submit feedback');
        }
      } catch (error) {
        console.error('Error:', error);
        alert('An error occurred while saving feedback. Please try again.');
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
.container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;

  h1 {
    text-align: center;
    color: #2c3e50;
  }

  .content {
    display: grid;
    gap: 20px;
    margin-top: 20px;
  }
}
</style> 