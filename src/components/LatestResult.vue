<template>
  <div v-if="result" class="result">
    <h2>Latest Entry</h2>
    <p>Date: {{ result.date }}</p>
    <p>Conditions: Temp={{ result.averageTemperature }}°C, 
       Duration={{ result.showerDuration }}min</p>
    <p>Suggested heating time: {{ result.actualHeatingTime.toFixed(2) }} minutes</p>
    <p>Feedback: {{ result.satisfaction }}/10 
       ({{ getSatisfactionText(result.satisfaction) }})</p>
  </div>
</template>

<script setup>
defineProps({
  result: {
    type: Object,
    required: true
  }
});

const getSatisfactionText = (satisfaction) => {
  if (satisfaction < 5) return 'too cold';
  if (satisfaction > 5) return 'too hot';
  return 'perfect';
};
</script>

<style lang="scss" scoped>
.result {
  background-color: $light-bg;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 30px;
}
</style> 