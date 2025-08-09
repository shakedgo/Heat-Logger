<template>
  <div class="ui-layer">
    <div class="toaster" aria-live="polite" aria-atomic="true">
      <transition-group name="toast" tag="div">
        <div
          v-for="t in ui.state.toasts"
          :key="t.id"
          class="toast"
          :class="t.type"
          role="status"
        >
          <span class="msg">{{ t.message }}</span>
        </div>
      </transition-group>
    </div>
    <div v-if="ui.state.confirm.visible" class="confirm-backdrop" @click="close(false)">
      <div class="confirm" @click.stop>
        <h3>{{ ui.state.confirm.title }}</h3>
        <p>{{ ui.state.confirm.message }}</p>
        <div class="actions">
          <button class="secondary" @click="close(false)">{{ ui.state.confirm.cancelText }}</button>
          <button class="primary" @click="close(true)">{{ ui.state.confirm.confirmText }}</button>
        </div>
      </div>
    </div>
  </div>
  
</template>

<script setup>
import { inject } from 'vue'
const ui = inject('ui')
function close(result) {
  ui.resolveConfirm(result)
}
</script>

<style scoped lang="scss">
.ui-layer { position: relative; }
.toaster {
  position: fixed;
  top: 12px;
  right: 12px;
  z-index: 1000;
  display: grid;
  gap: 8px;
}

.toast-enter-from, .toast-leave-to { opacity: 0; transform: translateY(-6px); }
.toast-enter-active, .toast-leave-active { transition: all .16s ease; }

.toast {
  background: #111827;
  color: white;
  padding: 14px 18px;
  border-radius: 12px;
  box-shadow: 0 12px 22px rgba(0,0,0,0.14);
  font-size: 15px;
  line-height: 1.25;
}
.toast.success { background: #065f46; }
.toast.error { background: #7f1d1d; }
.toast.info { background: #1f2937; }

.confirm-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.45);
  // backdrop-filter: blur(0.5px);
  display: grid;
  place-items: center;
  z-index: 900;
}
.confirm {
  background: white;
  color: #0f172a;
  width: min(92vw, 420px);
  padding: 18px 20px;
  border-radius: 14px;
  box-shadow: 0 10px 30px rgba(0,0,0,0.15);
}
.actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 14px;
}
.actions .secondary { background: #e5e7eb; color: #111827; }
.actions .primary { background: #10b981; }
.actions button {
  padding: 10px 14px;
  border: none;
  border-radius: 10px;
  font-weight: 600;
  cursor: pointer;
}

[data-theme='dark'] .confirm { background: #1f1f1f; color: #e5e7eb; }
[data-theme='dark'] .actions .secondary { background: #1f2937; color: #e5e7eb; }
</style>


