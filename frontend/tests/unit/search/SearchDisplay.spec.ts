import { createLocalVue, shallowMount } from '@vue/test-utils'
import SearchDisplay from '@/search/SearchDisplay.vue'
import Vuex from 'vuex'
import sinon from 'sinon'

describe('SearchDisplay', () => {
  it('kicks off a search when mounted', () => {
    const localVue = createLocalVue()

    const state = {
      search: {
        searchResults: null
      }
    }

    const actions = {
      doSearch: sinon.spy()
    }

    localVue.use(Vuex)
    const store = new Vuex.Store({
      state,
      actions
    })

    const inputQuery = 'testing rulez!'
    const mockRoute = {
      query: {
        query: inputQuery
      }
    }

    shallowMount(SearchDisplay, {
      localVue,
      store,
      mocks: {
        $route: mockRoute
      }
    })

    expect(actions.doSearch.calledOnce).toBeTruthy()
    expect(actions.doSearch.args[0][1]).toEqual(inputQuery)
  })

  it('shows the loading indicator when search results have not been loaded', () => {
    const localVue = createLocalVue()

    const state = {
      search: {
        searchResults: null
      }
    }

    const actions = {
      doSearch: sinon.spy()
    }

    localVue.use(Vuex)
    const store = new Vuex.Store({
      state,
      actions
    })

    const inputQuery = 'testing rulez!'
    const mockRoute = {
      query: {
        query: inputQuery
      }
    }

    const wrapper = shallowMount(SearchDisplay, {
      localVue,
      store,
      mocks: {
        $route: mockRoute
      }
    })

    expect(wrapper.find('#searchResultLoadingIndicator').exists()).toBeTruthy()
    expect(wrapper.find('#noResultsFound').exists()).toBeFalsy()
    expect(wrapper.find('#searchResults').exists()).toBeFalsy()
  })

  it('shows no results when results are empty', () => {
    const localVue = createLocalVue()

    const state = {
      search: {
        searchResults: []
      }
    }

    const actions = {
      doSearch: sinon.spy()
    }

    localVue.use(Vuex)
    const store = new Vuex.Store({
      state,
      actions
    })

    const inputQuery = 'testing rulez!'
    const mockRoute = {
      query: {
        query: inputQuery
      }
    }

    const wrapper = shallowMount(SearchDisplay, {
      localVue,
      store,
      mocks: {
        $route: mockRoute
      }
    })

    expect(wrapper.find('#searchResultLoadingIndicator').exists()).toBeFalsy()
    expect(wrapper.find('#noResultsFound').exists()).toBeTruthy()
    expect(wrapper.find('#searchResults').exists()).toBeFalsy()
  })

  it('shows results when they are present', () => {
    const localVue = createLocalVue()

    const state = {
      search: {
        searchResults: [
          {
            id: 3,
            title: 'Foo',
            match: 'Bar'
          },
          {
            id: 4,
            title: 'Foo',
            match: 'Bar'
          }
        ]
      }
    }

    const actions = {
      doSearch: sinon.spy()
    }

    localVue.use(Vuex)
    const store = new Vuex.Store({
      state,
      actions
    })

    const inputQuery = 'testing rulez!'
    const mockRoute = {
      query: {
        query: inputQuery
      }
    }

    const wrapper = shallowMount(SearchDisplay, {
      localVue,
      store,
      mocks: {
        $route: mockRoute
      }
    })

    expect(wrapper.find('#searchResultLoadingIndicator').exists()).toBeFalsy()
    expect(wrapper.find('#noResultsFound').exists()).toBeFalsy()
    expect(wrapper.find('#searchResults').exists()).toBeTruthy()
    expect(wrapper.find('#searchItem3').exists()).toBeTruthy()
    expect(wrapper.find('#searchItem4').exists()).toBeTruthy()
  })
})
