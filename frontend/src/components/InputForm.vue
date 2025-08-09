<template>
  <div class="input-form card">
    <h2>{{ currentEntry ? 'Enter Feedback' : "Enter Today's Data" }}</h2>
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
          step="1"
        >
      </div>
      
      <div class="form-group">
        <label for="duration">Shower Duration (minutes):</label>
        <input 
          type="number" 
          id="duration" 
          v-model="formData.showerDuration" 
          required 
          step="1"
        >
      </div>
      
      <button type="submit">Calculate Heating Time</button>
    </form>

    <form @submit.prevent="handleFeedback" v-if="currentEntry" class="feedback-form">
      <div class="current-values">
        <div class="cv-stats">
          <p>Temperature: {{ currentEntry.averageTemperature }}°C</p>
          <p>Duration: {{ currentEntry.showerDuration }} minutes</p>
        </div>
        <div class="cv-suggestion" v-if="currentEntry && currentEntry.heatingTime != null">
          <div class="label">Suggested Heating Time</div>
          <div class="value">{{ currentEntry.heatingTime.toFixed(1) }} minutes</div>
        </div>
        <div class="cv-suggestion pending" v-else>
          <div class="label">Suggested Heating Time</div>
          <div class="value">Calculating…</div>
        </div>
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
      this.$toast('Calculating heating time...', { type: 'info', duration: 1200 })
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
  padding: 24px;
  border-radius: 16px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);

  h2 { margin-top: 0; color: var(--heading); }
}

.form-group {
  margin-bottom: 14px;

  label { display: block; margin-bottom: 6px; color: var(--text); opacity: 0.9; }

  input {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    font-size: 16px;
  }

  small { display: block; color: var(--muted); margin-top: 5px; }
}

button {
  background-color: #10b981;
  color: white;
  padding: 12px 16px;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  font-size: 16px;
  width: 100%;

  &:hover {
    background-color: #059669;
  }
}

.current-values {
  background: #f8f9fa;
  padding: 16px;
  border-radius: 12px;
  margin-bottom: 20px;
  border: 1px solid #e5e7eb;

  .cv-stats p { margin: 4px 0; color: #334155; }

  .cv-suggestion {
    margin-top: 10px;
    background: #ecfdf5; /* light emerald tint */
    border: 1px solid #a7f3d0;
    border-radius: 10px;
    padding: 10px 12px;
    display: flex;
    align-items: baseline;
    justify-content: space-between;

    .label { color: #065f46; font-weight: 600; }
    .value { color: #065f46; font-weight: 700; font-size: 1.25rem; }
  }

  .cv-suggestion.pending {
    background: #f1f5f9;
    border-color: #cbd5e1;
    .label { color: #475569; }
    .value { color: #475569; font-weight: 600; }
  }
}

.button-group {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;

  .cancel-btn {
    background-color: #ef4444;

    &:hover {
      background-color: #dc2626;
    }
  }
}

.feedback-form {
  margin-top: 20px;
}

[data-theme='dark'] .input-form { background: #1f1f1f; }
[data-theme='dark'] .input-form h2 { color: #f3f4f6; }
[data-theme='dark'] .input-form label { color: #e5e7eb; }
[data-theme='dark'] .input-form small { color: #a1a1aa; }
[data-theme='dark'] .form-group label { color: #cbd5e1; }
[data-theme='dark'] .form-group input {
  background: #2a2a2a;
  color: #f3f4f6;
  border-color: #3a3a3a;
}
[data-theme='dark'] .form-group input::placeholder { color: #bdbdbd; }
[data-theme='dark'] .current-values {
  background: #2a2a2a; /* same as dark input bg */
  border-color: #3a3a3a;
}
[data-theme='dark'] .current-values .cv-stats p { color: #e5e7eb; }
[data-theme='dark'] .current-values .cv-suggestion {
  background: #052e2b; /* dark emerald tint */
  border-color: #0b4d45;
  .label { color: #a7f3d0; }
  .value { color: #34d399; }
}
[data-theme='dark'] .current-values .cv-suggestion.pending {
  background: #2f3136;
  border-color: #3a3a3a;
  .label, .value { color: #cbd5e1; }
}
[data-theme='dark'] .button-group .cancel-btn { background-color: #dc2626; }
</style> 