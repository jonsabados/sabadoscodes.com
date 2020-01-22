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
    </ul>
    <form class="form-inline my-2 my-lg-0" v-on:submit.prevent="executeSearch">
      <input v-model="query" class="form-control mr-sm-2" type="search" placeholder="Search" aria-label="Search" id="searchInput">
      <button class="btn my-2 my-sm-0 btn-light" type="submit" :disabled="disableSearch" id="searchButton">Search</button>
    </form>
  </nav>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { readQuery } from '@/search/search'
import { SEARCH_ROUTE_NAME } from '@/navigation/router'

@Component
export default class MainNav extends Vue {
  query: string = ''

  get disableSearch(): boolean {
    return this.query.trim() === '' || this.query === this.$store.state.search.query
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
}
</script>

<style lang="scss">
.btn:disabled{
  cursor: default;
}

#mainNav .router-link-active {
  font-weight: bold;
}
</style>
