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
        <h2>History</h2>
        <HistoryList 
          :history="history"
          @delete="handleDelete"
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
        const calcResponse = await fetch('http://localhost:8080/api/calculate', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(data)
        });
        const calcResult = await calcResponse.json();
        this.latestHeatingTime = calcResult.heatingTime;
      } catch (error) {
        console.error('Error:', error);
        alert('An error occurred while calculating. Please try again.');
      }
    },
    async handleSubmit(data) {
      try {
        const response = await fetch('http://localhost:8080/api/feedback', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(data)
        });

        if (response.ok) {
          // Only load history if feedback was successfully submitted
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
        const response = await fetch('http://localhost:8080/api/history');
        const data = await response.json();
        console.log('Loaded history:', data);
        this.history = data;
      } catch (error) {
        console.error('Error loading history:', error);
      }
    },
    async handleDelete(id) {
      try {
        const response = await fetch(`http://localhost:8080/api/history/delete`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ id })
        });
        
        if (!response.ok) {
          throw new Error('Failed to delete record');
        }
        
        // Reload history after successful deletion
        await this.loadHistory();
      } catch (error) {
        console.error('Error deleting record:', error);
        alert('Failed to delete record. Please try again.');
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