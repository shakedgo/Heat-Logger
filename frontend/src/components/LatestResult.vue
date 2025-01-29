<template>
  <div class="latest-result" v-if="heatingTime !== null">
    <h2>Calculated Heating Time</h2>
    <div class="result-value">
      {{ heatingTime.toFixed(1) }} minutes
    </div>
    <div class="result-note">
      This is the recommended heating time based on your input and previous feedback.
    </div>
  </div>
</template>

<script>
export default {
  name: 'LatestResult',
  props: {
    heatingTime: {
      type: Number,
      default: null
    }
  }
}
</script>

<style lang="scss" scoped>
.latest-result {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  text-align: center;

  h2 {
    margin-top: 0;
    color: #2c3e50;
  }

  .result-value {
    font-size: 2em;
    font-weight: bold;
    color: #42b983;
    margin: 20px 0;
  }

  .result-note {
    color: #666;
    font-size: 0.9em;
  }
}
</style> 