<template>
  <div class="input-form">
    <h2>Enter Today's Data</h2>
    <form @submit.prevent="handleCalculate" v-if="!currentEntry">
      <div class="form-group">
        <label for="userId">User ID:</label>
        <input 
          type="text" 
          id="userId" 
          v-model="formData.userId" 
          required 
          placeholder="Enter your user ID (e.g., user1)"
        >
        <small>This helps personalize predictions for you</small>
      </div>
      
      <div class="form-group">
        <label for="temperature">Average Temperature (°C):</label>
        <input 
          type="number" 
          id="temperature" 
          v-model="formData.averageTemperature" 
          required 
          step="0.1"
        >
      </div>
      
      <div class="form-group">
        <label for="duration">Shower Duration (minutes):</label>
        <input 
          type="number" 
          id="duration" 
          v-model="formData.showerDuration" 
          required 
          step="0.5"
        >
      </div>
      
      <button type="submit">Calculate Heating Time</button>
    </form>

    <form @submit.prevent="handleFeedback" v-if="currentEntry" class="feedback-form">
      <div class="current-values">
        <p>Temperature: {{ currentEntry.averageTemperature }}°C</p>
        <p>Duration: {{ currentEntry.showerDuration }} minutes</p>
        <p>Suggested Heating Time: {{ currentEntry.heatingTime.toFixed(1) }} minutes</p>
      </div>

      <div class="form-group">
        <label for="satisfaction">How was your shower? (1-100):</label>
        <input 
          type="number" 
          id="satisfaction" 
          v-model="formData.satisfaction" 
          required
          step="1"
          min="1" 
          max="100"
        >
        <small>1 = too cold, 50 = perfect, 100 = too hot</small>
      </div>
      
      <div class="button-group">
        <button type="submit" class="submit-btn">Submit Feedback</button>
        <button type="button" class="cancel-btn" @click="resetForm">Cancel</button>
      </div>
    </form>
  </div>
</template>

<script>
export default {
  name: 'InputForm',
  props: {
    latestHeatingTime: {
      type: Number,
      default: null
    }
  },
  data() {
    return {
      formData: {
        userId: localStorage.getItem('heatLogger_userId') || '',
        averageTemperature: '',
        showerDuration: '',
        satisfaction: ''
      },
      currentEntry: null
    }
  },
  methods: {
    handleCalculate() {
      // Save userId to localStorage for future use
      localStorage.setItem('heatLogger_userId', this.formData.userId);
      
      const data = {
        userId: this.formData.userId,
        duration: parseFloat(this.formData.showerDuration),
        temperature: parseFloat(this.formData.averageTemperature)
      };

      this.$emit('calculate', data);
      this.currentEntry = {
        userId: this.formData.userId,
        date: new Date().toISOString(),
        averageTemperature: parseFloat(this.formData.averageTemperature),
        showerDuration: parseFloat(this.formData.showerDuration)
      };
    },
    handleFeedback() {
      if (!this.currentEntry || this.latestHeatingTime === null) {
        console.warn('Cannot submit feedback without a valid entry and heating time');
        return;
      }

      const feedbackData = {
        ...this.currentEntry,
        heatingTime: this.latestHeatingTime,
        satisfaction: parseFloat(this.formData.satisfaction)
      };

      this.$emit('submitFeedback', feedbackData);
      this.resetForm();
    },
    resetForm() {
      // Preserve userId when resetting form
      const savedUserId = this.formData.userId;
      this.formData = {
        userId: savedUserId,
        averageTemperature: '',
        showerDuration: '',
        satisfaction: ''
      };
      this.currentEntry = null;
    }
  },
  watch: {
    latestHeatingTime: {
      handler(newValue) {
        if (this.currentEntry && newValue !== null) {
          this.currentEntry = { ...this.currentEntry, heatingTime: newValue };
        }
      },
      immediate: true
    }
  }
}
</script>

<style lang="scss" scoped>
.input-form {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);

  h2 {
    margin-top: 0;
    color: #2c3e50;
  }
}

.form-group {
  margin-bottom: 15px;

  label {
    display: block;
    margin-bottom: 5px;
    color: #34495e;
  }

  input {
    width: 100%;
    padding: 8px;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 16px;
  }

  small {
    display: block;
    color: #666;
    margin-top: 5px;
  }
}

button {
  background-color: #42b983;
  color: white;
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  width: 100%;

  &:hover {
    background-color: #3aa876;
  }
}

.current-values {
  background: #f8f9fa;
  padding: 15px;
  border-radius: 4px;
  margin-bottom: 20px;

  p {
    margin: 5px 0;
    color: #2c3e50;
  }
}

.button-group {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;

  .cancel-btn {
    background-color: #dc3545;

    &:hover {
      background-color: #c82333;
    }
  }
}

.feedback-form {
  margin-top: 20px;
}
</style> 