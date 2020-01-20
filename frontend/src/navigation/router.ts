import VueRouter from 'vue-router'
import Home from '@/home/Home.vue'
import Search from '@/search/SearchDisplay.vue'

export const HOME_ROUTE_NAME = 'home'
export const SEARCH_ROUTE_NAME = 'search'

const routes = [
  {
    path: '/',
    name: HOME_ROUTE_NAME,
    component: Home
  },
  {
    path: '/search',
    name: SEARCH_ROUTE_NAME,
    component: Search
  }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
})

export default router
