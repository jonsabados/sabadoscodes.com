import Vue from 'vue'
import App from './App.vue'
import router from './navigation/router'
import Vuex from 'vuex'
import { SearchStore } from '@/search/SearchStore'
import { AppStore } from '@/app/AppStore'
import VueRouter from 'vue-router'
import Bootstrap from 'bootstrap-vue'

Vue.config.productionTip = false
Vue.use(Vuex)
Vue.use(VueRouter)
Vue.use(Bootstrap)

interface RootState {}

const store = new Vuex.Store<RootState>({
  state: {},
  modules: {
    app: AppStore,
    search: SearchStore
  }
})

new Vue({
  router,
  store,
  render: h => h(App)
}).$mount('#app')
