<template>
  <div class="container">
    <h1>Water Heater Learning System</h1>
    
    <!-- Input Form -->
    <form @submit.prevent="handleSubmit" class="input-form">
      <div class="form-group">
        <label for="temperature">Average Temperature (°C):</label>
        <input 
          type="number" 
          id="temperature" 
          v-model="formData.temperature" 
          required 
          step="0.1"
        >
      </div>
      
      <div class="form-group">
        <label for="duration">Shower Duration (minutes):</label>
        <input 
          type="number" 
          id="duration" 
          v-model="formData.duration" 
          required 
          step="0.5"
        >
      </div>
      
      <div class="form-group">
        <label for="satisfaction">Satisfaction (1-10):</label>
        <input 
          type="number" 
          id="satisfaction" 
          v-model="formData.satisfaction" 
          required 
          min="1" 
          max="10"
        >
        <small>1 = too cold, 5 = perfect, 10 = too hot</small>
      </div>
      
      <button type="submit">Submit</button>
    </form>

    <!-- Latest Result -->
    <div v-if="latestResult" class="result">
      <h2>Latest Entry</h2>
      <p>Date: {{ latestResult.date }}</p>
      <p>Conditions: Temp={{ latestResult.averageTemperature }}°C, 
         Duration={{ latestResult.showerDuration }}min</p>
      <p>Suggested heating time: {{ latestResult.actualHeatingTime.toFixed(2) }} minutes</p>
      <p>Feedback: {{ latestResult.satisfaction }}/10 
         ({{ getSatisfactionText(latestResult.satisfaction) }})</p>
    </div>

    <!-- History -->
    <div class="history">
      <h2>History</h2>
      <div v-for="entry in recentHistory" :key="entry.date" class="history-entry">
        <p>{{ entry.date }}: {{ entry.actualHeatingTime.toFixed(2) }} minutes 
           ({{ entry.satisfaction }}/10)</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';

// State
const history = ref(JSON.parse(localStorage.getItem('heatingHistory')) || []);
const latestResult = ref(null);
const formData = ref({
  temperature: '',
  duration: '',
  satisfaction: ''
});

// Computed
const recentHistory = computed(() => {
  return [...history.value].reverse().slice(0, 7);
});

// Methods
const calculateHeatingTime = (day) => {
  // Base heating time calculation
  let baseHeatingTime = day.showerDuration + 10;
  let temperatureFactor = (30 - day.averageTemperature) / 2;
  let heatingTime = baseHeatingTime + temperatureFactor;

  // Learn from previous feedback
  if (history.value.length > 0) {
    const recentHistory = history.value.slice(-5);
    let totalAdjustment = 0;
    let contributingEntries = 0;

    recentHistory.forEach(entry => {
      if (entry.satisfaction !== 5) {
        contributingEntries++;
        if (entry.satisfaction < 5) {
          totalAdjustment += (5 - entry.satisfaction) * 2;
        } else if (entry.satisfaction > 5) {
          totalAdjustment -= (entry.satisfaction - 5) * 2;
        }
        
        const tempDiff = Math.abs(entry.averageTemperature - day.averageTemperature);
        if (tempDiff < 5) {
          totalAdjustment *= 1.5;
        }
      }
    });

    heatingTime += contributingEntries > 0 ? totalAdjustment / contributingEntries : 0;
  }

  return Math.max(20, Math.min(80, heatingTime));
};

const saveHistory = () => {
  localStorage.setItem('heatingHistory', JSON.stringify(history.value));
};

const getSatisfactionText = (satisfaction) => {
  if (satisfaction < 5) return 'too cold';
  if (satisfaction > 5) return 'too hot';
  return 'perfect';
};

const handleSubmit = () => {
  const day = {
    averageTemperature: parseFloat(formData.value.temperature),
    showerDuration: parseFloat(formData.value.duration),
    satisfaction: parseFloat(formData.value.satisfaction),
    date: new Date().toISOString().split('T')[0]
  };

  if (day.satisfaction < 1 || day.satisfaction > 10) {
    alert("Please provide feedback between 1 (too cold) and 10 (too hot)");
    return;
  }

  day.actualHeatingTime = calculateHeatingTime(day);
  history.value.push(day);
  saveHistory();
  latestResult.value = day;

  // Reset form
  formData.value = {
    temperature: '',
    duration: '',
    satisfaction: ''
  };
};

// Initialize
onMounted(() => {
  if (history.value.length > 0) {
    latestResult.value = history.value[history.value.length - 1];
  }
});
</script>

<style scoped>
.container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
  font-family: Arial, sans-serif;
}

.input-form {
  margin-bottom: 30px;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
}

.form-group input {
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  width: 200px;
}

.form-group small {
  display: block;
  color: #666;
  margin-top: 5px;
}

button {
  background-color: #4CAF50;
  color: white;
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

button:hover {
  background-color: #45a049;
}

.result {
  background-color: #f5f5f5;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 30px;
}

.history-entry {
  padding: 10px;
  border-bottom: 1px solid #eee;
}

.history-entry:last-child {
  border-bottom: none;
}
</style> 