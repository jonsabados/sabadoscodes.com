<template>
  <nav role="navigation" class="navbar navbar-expand-lg sticky-top navbar-dark bg-dark" id="mainNav">
    <ul class="navbar-nav mr-auto">
      <li class="nav-item">
        <router-link :to="{name: 'home'}" class="nav-link">Home</router-link>
      </li>
      <li class="nav-item">
        <router-link :to="{name: 'articles'}" class="nav-link">Articles</router-link>
      </li>
      <li class="nav-item">
        <router-link :to="{name: 'about'}" class="nav-link">About</router-link>
      </li>
      <b-nav-item-dropdown v-if="isAdmin" text="Admin" right>
        <b-dropdown-item v-if="hasArticleAssetRole" :to="{name: 'adminAssets'}">Article Assets</b-dropdown-item>
      </b-nav-item-dropdown>
    </ul>
    <div class="my-2 my-lg-0">
      <sign-in />
      <form class="form-inline" v-on:submit.prevent="executeSearch" id="searchForm">
        <input v-model="query" class="form-control mr-sm-2" type="search" placeholder="Search" aria-label="Search" id="searchInput">
        <button class="btn my-2 my-sm-0 btn-light" type="submit" :disabled="disableSearch" id="searchButton">Search</button>
      </form>
    </div>
  </nav>
</template>

<script lang="ts">
import SignIn from '@/user/SignIn.vue'
import { Component, Vue, Watch } from 'vue-property-decorator'
import { readQuery } from '@/search/search'
import { SEARCH_ROUTE_NAME } from '@/navigation/router'
import { routeChanged } from '@/legal/analytics'

@Component({
  components: {
    SignIn
  }
})
export default class MainNav extends Vue {
  query: string = ''

  get disableSearch(): boolean {
    return this.query.trim() === '' || this.query === this.$store.state.search.query
  }

  get isAdmin(): boolean {
    return this.hasArticleAssetRole
  }

  get hasArticleAssetRole(): boolean {
    return this.$store.state.user.self && this.$store.state.user.self.roles && this.$store.state.user.self.roles.includes('article_asset_publish')
  }

  mounted() {
    this.query = readQuery(this.$route)
  }

  executeSearch() {
    this.$router.push({
      name: SEARCH_ROUTE_NAME,
      query: {
        query: this.query
      }
    })
  }

  @Watch('$route')
  routeUpdated(r: any) {
    routeChanged({ 'page_path': r.path })
  }
}
</script>

<style lang="scss">
.btn:disabled{
  cursor: default;
}

#mainNav .router-link-active {
  font-weight: bold;
}

#searchForm {
  display: inline;
}
</style>
