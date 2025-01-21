<template>
  <div class="history">
    <h2>History</h2>
    <div v-for="entry in entries" :key="entry.date" class="history-entry">
      <p>{{ entry.date }}: {{ entry.actualHeatingTime.toFixed(2) }} minutes 
         ({{ entry.satisfaction }}/10)</p>
    </div>
  </div>
</template>

<script setup>
defineProps({
  entries: {
    type: Array,
    required: true
  }
});
</script>

<style lang="scss" scoped>
.history-entry {
  padding: 10px;
  border-bottom: 1px solid $border-color;

  &:last-child {
    border-bottom: none;
  }
}
</style> 