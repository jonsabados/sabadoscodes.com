<template>
  <nav role="navigation" class="navbar navbar-expand-lg navbar-light bg-light">
    <ul class="navbar-nav mr-auto">
      <li class="nav-item">
        <router-link to="/" class="nav-link">Home</router-link>
      </li>
      <li class="nav-item">
        <router-link to="/stuff" class="nav-link">Stuff</router-link>
      </li>
      <li class="nav-item">
        <router-link to="/about" class="nav-link">About</router-link>
      </li>
    </ul>
    <form class="form-inline my-2 my-lg-0" v-on:submit.prevent="executeSearch">
      <input v-model="searchQuery" class="form-control mr-sm-2" type="search" placeholder="Search" aria-label="Search" id="searchInput">
      <button class="btn btn-outline-dark my-2 my-sm-0" type="submit" :disabled="disableSearch" id="searchButton">Search</button>
    </form>
  </nav>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { executeSearch } from '@/search/search'

@Component
export default class MainNav extends Vue {
  searchQuery: string = ''

  get disableSearch(): boolean {
    return this.searchQuery.trim() === ''
  }

  executeSearch() {
    executeSearch(this.searchQuery)
  }
}
</script>

<style lang="scss">

</style>
