import Vue from 'vue'
import App from './App.vue'
import router from './navigation/router'
import Vuex from 'vuex'
import { SearchStore } from '@/search/SearchStore'
import { AppStore } from '@/app/AppStore'
import { UserStore } from '@/user/UserStore'
import VueRouter from 'vue-router'
import Bootstrap from 'bootstrap-vue'
// @ts-ignore
import { LoaderPlugin } from 'vue-google-login'

Vue.config.productionTip = false
Vue.use(Vuex)
Vue.use(VueRouter)
Vue.use(Bootstrap)

interface RootState {
}

Vue.use(LoaderPlugin, {
  client_id: process.env.VUE_APP_GOOGLE_OAUTH_CLIENT_ID
})

const store = new Vuex.Store<RootState>({
  state: {},
  modules: {
    app: AppStore,
    search: SearchStore,
    user: UserStore
  }
})

new Vue({
  router,
  store,
  render: h => h(App)
}).$mount('#app')
