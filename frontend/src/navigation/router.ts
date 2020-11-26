import VueRouter from 'vue-router'
import Home from '@/home/Home.vue'
import About from '@/about/About.vue'
import Search from '@/search/SearchDisplay.vue'
import Articles from '@/articles/Articles.vue'
import ArticlesHome from '@/articles/ArticlesHome.vue'
import Article from '@/articles/Article.vue'
import Privacy from '@/legal/Privacy.vue'
import Admin from '@/admin/Admin.vue'
import AdminHome from '@/admin/AdminHome.vue'
import AdminArticles from '@/admin/articles/Articles.vue'
import AdminAssets from '@/admin/articles/Assets.vue'
import AdminAssetUpload from '@/admin/articles/AssetUpload.vue'

export const HOME_ROUTE_NAME = 'home'
export const SEARCH_ROUTE_NAME = 'search'
export const ABOUT_ROUTE_NAME = 'about'
export const ARTICLES_HOME_ROUTE_NAME = 'articles'
export const ARTICLE_ROUTE_NAME = 'article'
export const PRIVACY_ROUTE_NAME = 'privacy'
export const ADMIN_ROUTE_NAME = 'admin'
export const ADMIN_ARTICLES_ROUTE_NAME = 'adminArticles'
export const ADMIN_ARTICLES_ASSETS_NAME = 'adminAssets'
export const ADMIN_ARTICLES_ASSET_UPLOAD_NAME = 'adminAssetUpload'

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
  },
  {
    path: '/privacy',
    name: PRIVACY_ROUTE_NAME,
    component: Privacy
  },
  {
    path: '/admin',
    component: Admin,
    children: [
      {
        path: '/',
        name: ADMIN_ROUTE_NAME,
        component: AdminHome
      },
      {
        path: 'articles',
        name: ADMIN_ARTICLES_ROUTE_NAME,
        component: AdminArticles
      },
      {
        path: 'assets',
        name: ADMIN_ARTICLES_ASSETS_NAME,
        component: AdminAssets,
        children: [
          {
            path: 'upload',
            name: ADMIN_ARTICLES_ASSET_UPLOAD_NAME,
            component: AdminAssetUpload
          }
        ]
      }
    ]
  }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
})

export default router
