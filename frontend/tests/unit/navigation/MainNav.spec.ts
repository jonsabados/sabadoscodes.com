import { createLocalVue, shallowMount } from '@vue/test-utils'
import MainNav from '@/navigation/MainNav.vue'
import VueRouter, { Route } from 'vue-router'
import Vuex from 'vuex'

describe('MainNav', () => {
  it('disables the search button when no search string has been entered', () => {
    const localVue = createLocalVue()
    localVue.use(VueRouter)

    const router = new VueRouter()

    const state = {
      search: {
        query: ''
      }
    }

    localVue.use(Vuex)
    const store = new Vuex.Store({
      state
    })

    const wrapper = shallowMount(MainNav, {
      attachToDocument: true,
      localVue,
      router,
      store
    })

    expect(wrapper.find('#searchButton').is(':disabled')).toBeTruthy()

    wrapper.destroy()
  })

  it('disables the search button when only whitespace has been entered', () => {
    const localVue = createLocalVue()
    localVue.use(VueRouter)

    const router = new VueRouter()

    const state = {
      search: {
        query: ''
      }
    }

    localVue.use(Vuex)
    const store = new Vuex.Store({
      state
    })

    const wrapper = shallowMount(MainNav, {
      attachToDocument: true,
      localVue,
      router,
      store
    })

    wrapper.find('#searchInput').setValue('   ')
    expect(wrapper.find('#searchButton').is(':disabled')).toBeTruthy()

    wrapper.destroy()
  })

  it('enables the search button when a string has been entered', () => {
    const localVue = createLocalVue()
    localVue.use(VueRouter)

    const router = new VueRouter()

    const state = {
      search: {
        query: ''
      }
    }

    localVue.use(Vuex)
    const store = new Vuex.Store({
      state
    })

    const wrapper = shallowMount(MainNav, {
      attachToDocument: true,
      localVue,
      router,
      store
    })

    wrapper.find('#searchInput').setValue('test')
    expect(wrapper.find('#searchButton').is(':disabled')).toBeFalsy()

    wrapper.destroy()
  })

  it('executes a search when the search form is submitted', () => {
    const localVue = createLocalVue()
    localVue.use(VueRouter)

    const router = new VueRouter()

    const state = {
      search: {
        query: ''
      }
    }

    localVue.use(Vuex)
    const store = new Vuex.Store({
      state
    })

    const wrapper = shallowMount(MainNav, {
      attachToDocument: true,
      router,
      localVue,
      store
    })

    const searchString = 'test'

    let callCount = 0
    let routePassed: any = null

    router.push = async(r:any):Promise<Route> => {
      callCount++
      routePassed = r
      return r
    }

    wrapper.find('#searchInput').setValue(searchString)
    wrapper.find('#searchButton').element.click()
    expect(callCount).toEqual(1)
    expect(routePassed).toEqual({
      name: 'search',
      query: {
        query: searchString
      }
    })

    wrapper.destroy()
  })
})
