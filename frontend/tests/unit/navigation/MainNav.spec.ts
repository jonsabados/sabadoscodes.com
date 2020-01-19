import { createLocalVue, shallowMount } from '@vue/test-utils'
import MainNav from '@/navigation/MainNav.vue'
import { executeSearch } from '@/search/search'
import VueRouter from 'vue-router'

jest.mock('@/search/search')

const executeSearchMock = executeSearch as unknown as jest.Mock

describe('MainNav', () => {
  beforeEach(() => {
    executeSearchMock.mockClear()
  })

  it('disables the search button when no search string has been entered', () => {
    const localVue = createLocalVue()
    localVue.use(VueRouter)
    const router = new VueRouter()

    const wrapper = shallowMount(MainNav, {
      attachToDocument: true,
      localVue,
      router
    })

    expect(wrapper.find('#searchButton').is(':disabled')).toBeTruthy()

    wrapper.destroy()
  })

  it('disables the search button when only whitespace has been entered', () => {
    const localVue = createLocalVue()
    localVue.use(VueRouter)
    const router = new VueRouter()

    const wrapper = shallowMount(MainNav, {
      attachToDocument: true,
      localVue,
      router
    })

    wrapper.find('#searchInput ').setValue('   ')
    expect(wrapper.find('#searchButton').is(':disabled')).toBeTruthy()

    wrapper.destroy()
  })

  it('enables the search button when a string has been entered', () => {
    const localVue = createLocalVue()
    localVue.use(VueRouter)
    const router = new VueRouter()

    const wrapper = shallowMount(MainNav, {
      attachToDocument: true,
      localVue,
      router
    })

    wrapper.find('#searchInput ').setValue('test')
    expect(wrapper.find('#searchButton').is(':disabled')).toBeFalsy()

    wrapper.destroy()
  })

  it('executes a search when the search form is submitted', () => {
    const localVue = createLocalVue()
    localVue.use(VueRouter)
    const router = new VueRouter()

    const wrapper = shallowMount(MainNav, {
      attachToDocument: true,
      localVue,
      router
    })

    const searchString = 'test'

    let searchCalled:boolean = false
    executeSearchMock.mockImplementation((input) => {
      searchCalled = true
      expect(input).toEqual(searchString)
    })

    wrapper.find('#searchInput ').setValue(searchString)
    wrapper.find('#searchButton').element.click()
    expect(searchCalled).toBeTruthy()

    wrapper.destroy()
  })
})
