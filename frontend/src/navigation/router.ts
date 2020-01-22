import VueRouter from 'vue-router'
import Home from '@/home/Home.vue'
import About from '@/about/About.vue'
import Search from '@/search/SearchDisplay.vue'
import Articles from '@/articles/Articles.vue'
import ArticlesHome from '@/articles/ArticlesHome.vue'
import Article from '@/articles/Article.vue'

export const HOME_ROUTE_NAME = 'home'
export const SEARCH_ROUTE_NAME = 'search'
export const ABOUT_ROUTE_NAME = 'about'
export const ARTICLES_HOME_ROUTE_NAME = 'articles'
export const ARTICLE_ROUTE_NAME = 'article'

const routes = [
  {
    path: '/',
    name: HOME_ROUTE_NAME,
    component: Home
  },
  {
    path: '/articles',
    component: Articles,
    children: [
      {
        path: '/',
        name: ARTICLES_HOME_ROUTE_NAME,
        component: ArticlesHome
      },
      {
        path: 'article/:id',
        name: ARTICLE_ROUTE_NAME,
        component: Article
      }
    ]
  },
  {
    path: '/about',
    name: ABOUT_ROUTE_NAME,
    component: About
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
