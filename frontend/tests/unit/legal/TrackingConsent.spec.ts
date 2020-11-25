import { get, set } from 'js-cookie'
import { createLocalVue, shallowMount } from '@vue/test-utils'
import TrackingConsent from '@/legal/TrackingConsent.vue'
import VueRouter from 'vue-router'
import BootstrapVue from 'bootstrap-vue'
import startAnalytics from '@/legal/analytics'

jest.mock('js-cookie')
jest.mock('@/legal/analytics')

const getMock = get as unknown as jest.Mock
const setMock = set as unknown as jest.Mock
const startAnalyticsMock = startAnalytics as jest.Mock

describe('TrackingConsent', () => {
  beforeEach(() => {
    getMock.mockClear()
    setMock.mockClear()
  })

  it('does not show if a skipTracking parameter is present', () => {
    const localVue = createLocalVue()
    localVue.use(BootstrapVue)

    let startAnalyticsCalled = false
    startAnalyticsMock.mockImplementation(() => {
      startAnalyticsCalled = true
    })

    const wrapper = shallowMount(TrackingConsent, {
      localVue,
      mocks: {
        $route: {
          query: {
            skipTracking: 'true'
          }
        }
      }
    })

    expect(wrapper.find('#cookieConsentGatherer').exists()).toBeFalsy()
    expect(startAnalyticsCalled).toBeFalsy()
  })

  it('does not show if consent was rejected', () => {
    const localVue = createLocalVue()
    localVue.use(VueRouter)
    localVue.use(BootstrapVue)

    let startAnalyticsCalled = false
    startAnalyticsMock.mockImplementation(() => {
      startAnalyticsCalled = true
    })

    getMock.mockImplementation((cookie:string) => {
      expect(cookie).toEqual('okToTrack')
      return 'no'
    })

    const router = new VueRouter()

    const wrapper = shallowMount(TrackingConsent, {
      localVue,
      router
    })

    expect(wrapper.find('#cookieConsentGatherer').exists()).toBeFalsy()
    expect(startAnalyticsCalled).toBeFalsy()
  })

  it('does not show and enables google analytics if consent was given', () => {
    const localVue = createLocalVue()
    localVue.use(VueRouter)
    localVue.use(BootstrapVue)

    let startAnalyticsCalled = false
    startAnalyticsMock.mockImplementation(() => {
      startAnalyticsCalled = true
    })

    getMock.mockImplementation((cookie:string) => {
      expect(cookie).toEqual('okToTrack')
      return 'yes'
    })

    const router = new VueRouter()

    const wrapper = shallowMount(TrackingConsent, {
      localVue,
      router
    })

    expect(wrapper.find('#cookieConsentGatherer').exists()).toBeFalsy()
    expect(startAnalyticsCalled).toBeTruthy()
  })

  it('properly handles consent', () => {
    const localVue = createLocalVue()
    localVue.use(VueRouter)
    localVue.use(BootstrapVue)

    let startAnalyticsCalled = false
    startAnalyticsMock.mockImplementation(() => {
      startAnalyticsCalled = true
    })

    const setCookies:Map<string, string> = new Map()

    getMock.mockImplementation((cookie:string) => {
      expect(cookie).toEqual('okToTrack')
      return undefined
    })

    setMock.mockImplementation((cookie:string, value:string) => {
      setCookies.set(cookie, value)
    })

    const router = new VueRouter()

    const wrapper = shallowMount(TrackingConsent, {
      attachToDocument: true,
      localVue,
      router
    })

    expect(wrapper.find('#cookieConsentGatherer').exists()).toBeTruthy()
    expect(startAnalyticsCalled).toBeFalsy()
    wrapper.find('#allowTrackingButton').element.click()
    expect(wrapper.find('#cookieConsentGatherer').exists()).toBeFalsy()
    expect(startAnalyticsCalled).toBeTruthy()
    expect(setCookies.get('okToTrack')).toEqual('yes')

    wrapper.destroy()
  })

  it('properly handles the declining of consent', () => {
    const localVue = createLocalVue()
    localVue.use(VueRouter)
    localVue.use(BootstrapVue)

    let startAnalyticsCalled = false
    startAnalyticsMock.mockImplementation(() => {
      startAnalyticsCalled = true
    })

    const setCookies:Map<string, string> = new Map()

    getMock.mockImplementation((cookie:string) => {
      expect(cookie).toEqual('okToTrack')
      return undefined
    })

    setMock.mockImplementation((cookie:string, value:string) => {
      setCookies.set(cookie, value)
    })

    const router = new VueRouter()

    const wrapper = shallowMount(TrackingConsent, {
      attachToDocument: true,
      localVue,
      router
    })

    expect(wrapper.find('#cookieConsentGatherer').exists()).toBeTruthy()
    expect(startAnalyticsCalled).toBeFalsy()
    wrapper.find('#rejectTrackingButton').element.click()
    expect(wrapper.find('#cookieConsentGatherer').exists()).toBeFalsy()
    expect(startAnalyticsCalled).toBeFalsy()
    expect(setCookies.get('okToTrack')).toEqual('no')

    wrapper.destroy()
  })
})
