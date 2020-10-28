<template>
  <div v-if="gatherConsent" class="container-fluid" id="cookieConsentGatherer">
    <b-alert v-model="gatherConsent" class="alert-warning" dismissible>
      In order to gain insights into how it is used this site makes use of tracking cookies. Dismissing this alert will
      prevent this and will prevent the storing of any tracking cookies. If you choose <i>Do Not Track Me, Ever</i> a
      non-identifying cookie will be set to retain this preference. If you are OK with the tracking you may choose
      <i>OK, track away!</i> For more details see the
      <router-link :to="{name: 'privacy', query: {skipTracking: 'true'}}">Privacy Policy</router-link>
      <div class="container-fluid text-center" aria-label="Tracking Options" id="trackingOptionButtons">
        <button type="button" class="btn btn-primary" @click="consentGathered" id="allowTrackingButton">OK, track away!</button>&nbsp;
        <button type="button" class="btn btn-primary" @click="doNotTrack" id="rejectTrackingButton">Do Not Track Me, Ever</button>
      </div>
    </b-alert>
  </div>
</template>

<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import Cookies from 'js-cookie'
import startAnalytics from '@/legal/analytics'

const consentCookie = 'okToTrack'
const okToTrackValue = 'yes'
const doNotTrackValue = 'no'

@Component
export default class TrackingConsent extends Vue {
  gatherConsent: boolean = true

  created() {
    this.loadInitialState()
  }

  @Watch('$route')
  routeUpdated() {
    this.loadInitialState()
  }

  loadInitialState() {
    if (this.$route.query.skipTracking || isBot()) {
      this.gatherConsent = false
      return
    }
    const okToTrack = Cookies.get(consentCookie)
    if (okToTrack) {
      if (okToTrack === okToTrackValue) {
        this.gatherConsent = false
        startAnalytics()
      } else {
        this.gatherConsent = false
      }
    }
  }

  doNotTrack() {
    this.gatherConsent = false
    Cookies.set(consentCookie, doNotTrackValue)
  }

  consentGathered() {
    this.gatherConsent = false
    Cookies.set(consentCookie, okToTrackValue)
    startAnalytics()
  }
}

function isBot():boolean {
  return /bot|googlebot|crawler|spider|robot|crawling/i.test(navigator.userAgent)
}
</script>

<style lang="scss">
#cookieConsentGatherer {
  position: absolute;
  top: 0;
  z-index: 1070;
  #trackingOptionButtons button {
    margin-top: .5em
  }
}
</style>
