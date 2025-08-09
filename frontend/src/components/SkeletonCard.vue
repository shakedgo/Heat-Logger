<template>
  <div class="skeleton card">
    <div class="line w-40"></div>
    <div class="line w-80"></div>
    <div class="line w-60"></div>
    <div class="line w-70"></div>
  </div>
</template>

<style scoped lang="scss">
.skeleton {
  background: linear-gradient(90deg, rgba(0,0,0,0.04), rgba(0,0,0,0.06), rgba(0,0,0,0.04));
  background-size: 200% 100%;
  animation: shimmer 1.2s infinite;
}
.line { height: 12px; background: rgba(255,255,255,0.6); border-radius: 6px; margin: 10px 0; }
.w-40 { width: 40%; }
.w-60 { width: 60%; }
.w-70 { width: 70%; }
.w-80 { width: 80%; }
@keyframes shimmer { 0% { background-position: 0% 0; } 100% { background-position: -200% 0; } }

[data-theme='dark'] .skeleton {
  background: linear-gradient(90deg, rgba(255,255,255,0.06), rgba(255,255,255,0.09), rgba(255,255,255,0.06));
}
.line { background: rgba(0,0,0,0.06); }
</style>


