import { reactive, readonly } from 'vue'

function createUiStore() {
  const state = reactive({
    toasts: [],
    confirm: {
      visible: false,
      title: '',
      message: '',
      confirmText: 'Confirm',
      cancelText: 'Cancel',
      resolve: null,
      reject: null,
      variant: 'default',
    },
  })

  function showToast(message, options = {}) {
    const id = crypto.randomUUID ? crypto.randomUUID() : Math.random().toString(36).slice(2)
    const toast = {
      id,
      message,
      type: options.type || 'info',
      duration: options.duration || 3000,
    }
    state.toasts.push(toast)
    setTimeout(() => {
      const index = state.toasts.findIndex(t => t.id === id)
      if (index !== -1) state.toasts.splice(index, 1)
    }, toast.duration)
  }

  function openConfirm({ title, message, confirmText = 'Confirm', cancelText = 'Cancel', variant = 'default' }) {
    return new Promise((resolve, reject) => {
      state.confirm = {
        visible: true,
        title,
        message,
        confirmText,
        cancelText,
        resolve,
        reject,
        variant,
      }
    })
  }

  function resolveConfirm(confirmed) {
    const { resolve, reject } = state.confirm
    state.confirm.visible = false
    if (confirmed) resolve && resolve(true)
    else reject && reject(new Error('Cancelled'))
  }

  return {
    state: readonly(state),
    _state: state,
    showToast,
    openConfirm,
    resolveConfirm,
  }
}

export default {
  install(app) {
    const ui = createUiStore()
    app.config.globalProperties.$toast = (message, options) => ui.showToast(message, options)
    app.config.globalProperties.$confirm = (options) => ui.openConfirm(options)
    app.provide('ui', ui)
  }
}


