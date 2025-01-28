<template>
  <div class="container">
    <h1>Water Heater Learning System</h1>
    <div class="content">
      <InputForm 
        @calculate="handleCalculate" 
        @submit="handleSubmit"
        :latestHeatingTime="latestHeatingTime"
      />
      <LatestResult :heatingTime="latestHeatingTime" />
      <HistoryList :history="history" />
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
        await fetch('http://localhost:8080/api/feedback', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(data)
        });

        // Only refresh history after submitting feedback
        await this.loadHistory();
        this.latestHeatingTime = null;
      } catch (error) {
        console.error('Error:', error);
        alert('An error occurred while saving feedback. Please try again.');
      }
    },
    async loadHistory() {
      try {
        const response = await fetch('http://localhost:8080/api/history');
        this.history = await response.json();
      } catch (error) {
        console.error('Error loading history:', error);
      }
    }
  },
  async created() {
    await this.loadHistory();
  }
}
</script>

<style>
.container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.content {
  display: grid;
  gap: 20px;
  margin-top: 20px;
}

h1 {
  text-align: center;
  color: #2c3e50;
}
</style> 