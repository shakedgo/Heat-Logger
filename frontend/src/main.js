import { createApp } from 'vue'
// System theme sync (no UI toggle)
const mq = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)')
function applySystemTheme(isDark) {
  const root = document.documentElement
  if (isDark) root.setAttribute('data-theme', 'dark')
  else root.removeAttribute('data-theme')
}
if (mq) {
  applySystemTheme(mq.matches)
  if (typeof mq.addEventListener === 'function') mq.addEventListener('change', e => applySystemTheme(e.matches))
  else mq.onchange = e => applySystemTheme(e.matches)
}
import './styles/main.scss'
import App from './App.vue'
import apiPlugin from './plugins/api'
import uiPlugin from './plugins/ui'

/* Import Font Awesome */
import { library } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { 
    faSnowflake, 
    faFire, 
    faTimes, 
    faTrash,
    faFileExport 
} from '@fortawesome/free-solid-svg-icons'

/* Add icons to the library */
library.add(faSnowflake, faFire, faTimes, faTrash, faFileExport)

/* Create app */
const app = createApp(App)

/* Register Font Awesome component globally */
app.component('font-awesome-icon', FontAwesomeIcon)

/* Register API plugin */
app.use(apiPlugin)
app.use(uiPlugin)

app.mount('#app')
