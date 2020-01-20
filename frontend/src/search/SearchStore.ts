import { Module, VuexModule, Action, Mutation } from 'vuex-module-decorators'
import { AppStore } from '@/app/AppStore'
import { executeSearch, SearchResult } from '@/search/search'

export interface SearchState {
  query: string
  searchResults: Array<SearchResult> | null
}

@Module
export class SearchStore extends VuexModule<SearchState> {
  static ACTION_DO_SEARCH: string = 'doSearch'
  static ACTION_CLEAR_SEARCH_RESULTS: string = 'clearSearchResults'
  static MUTATION_SET_SEARCH_RESULTS = 'setSearchResults'
  static MUTATION_SET_QUERY = 'setQuery'

  query: string = ''

  searchResults: Array<SearchResult> | null = null

  @Mutation
  setQuery(query: string) {
    this.query = query
  }

  @Mutation
  setSearchResults(searchResults: Array<SearchResult>) {
    this.searchResults = searchResults
  }

  @Action
  async doSearch(query: string): Promise<Array<SearchResult> | Error> {
    this.context.commit(SearchStore.MUTATION_SET_QUERY, query)
    this.context.commit(SearchStore.MUTATION_SET_SEARCH_RESULTS, null)
    try {
      const searchResults = await executeSearch(query)
      this.context.commit(SearchStore.MUTATION_SET_SEARCH_RESULTS, searchResults)
      return searchResults
    } catch (e) {
      await this.context.dispatch(AppStore.ACTION_REGISTER_REMOTE_ERROR, e)
      this.context.commit(SearchStore.MUTATION_SET_SEARCH_RESULTS, [])
      return e
    }
  }

  @Action
  async clearSearchResults() {
    this.context.commit(SearchStore.MUTATION_SET_QUERY, '')
    this.context.commit(SearchStore.MUTATION_SET_SEARCH_RESULTS, null)
  }
}
