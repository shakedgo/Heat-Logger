<template>
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
</template>

<script setup>
import { ref } from 'vue';

const emit = defineEmits(['submit']);

const formData = ref({
  temperature: '',
  duration: '',
  satisfaction: ''
});

const handleSubmit = () => {
  emit('submit', { ...formData.value });
  formData.value = {
    temperature: '',
    duration: '',
    satisfaction: ''
  };
};
</script>

<style lang="scss" scoped>
.input-form {
  margin-bottom: 30px;
}
</style> 