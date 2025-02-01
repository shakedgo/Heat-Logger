import { createApp } from 'vue'
import './styles/main.scss'
import App from './App.vue'
import apiPlugin from './plugins/api'

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

app.mount('#app')
