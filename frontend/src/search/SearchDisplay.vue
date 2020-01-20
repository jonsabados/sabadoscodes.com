<template>
  <div  class="container-fluid">
    <h1>Search Results</h1>
    <div v-if="isLoading">
      <loading id="searchResultLoadingIndicator" />
    </div>
    <div v-else-if="resultsPresent">
      <ol id="searchResults">
        <li v-for="item in searchResults" :key="item.id" class="searchItem" :id="`searchItem${item.id}`">
          <span class="searchTitle">{{ item.title }}</span>
          <p class="searchMatch">{{ item.match }}</p>
        </li>
      </ol>
    </div>
    <div v-else>
      <h3 id="noResultsFound">No results found.</h3>
    </div>
  </div>
</template>

<script lang="ts">
import { Vue, Component, Watch } from 'vue-property-decorator'
import { readQuery, SearchResult } from '@/search/search'
import Loading from '@/app/Loading.vue'
import { SearchStore } from '@/search/SearchStore'

@Component({
  components: {
    Loading
  }
})
export default class Search extends Vue {
  get isLoading(): boolean {
    return this.searchResults == null
  }

  get resultsPresent(): boolean {
    return this.searchResults != null && this.searchResults.length > 0
  }

  get searchResults(): Array<SearchResult> | null {
    return this.$store.state.search.searchResults
  }

  mounted() {
    this.queryUpdated()
  }

  destroyed() {
    this.$store.dispatch(SearchStore.ACTION_CLEAR_SEARCH_RESULTS)
  }

  @Watch('$route')
  queryUpdated() {
    this.$store.dispatch(SearchStore.ACTION_DO_SEARCH, readQuery(this.$route))
  }
}
</script>

<style lang="scss">

</style>
