// @ts-ignore
import { bootstrap } from 'vue-gtag'

let _callGTag = (opts?: any) => {}

export function routeChanged(opts?: any) {
  _callGTag(opts)
}

// jest has issues mocking bootstrap, so lets just wrap the thing
export default function startAnalytics() {
  bootstrap().then(gtag => {
    _callGTag = (opts?: any) => {
      gtag('config', process.env.VUE_APP_GOOGLE_ANALYTICS_ID, opts)
    }
    _callGTag()
  })
}
