// @ts-ignore
import { bootstrap } from 'vue-gtag'

// jest has issues mocking bootstrap, so lets just wrap the thing
export default function startAnalytics() {
  bootstrap()
}
