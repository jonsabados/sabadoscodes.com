<template>
  <div v-if="isReady" id="signInWidget">
    <google-login v-if="!isSignedIn" :params="googleLoginParams">Sign In With Google</google-login>
    <google-login v-else :params="googleLoginParams" :logoutButton="true">Logout</google-login>
  </div>
</template>

<script lang="ts">
// @ts-ignore
import GoogleLogin from 'vue-google-login'
import { Component, Vue } from 'vue-property-decorator'
import Loading from '@/app/Loading.vue'

@Component({
  components: {
    GoogleLogin,
    Loading
  }
})
export default class SignIn extends Vue {
  get isReady(): boolean {
    return this.$store.state.user.isReady
  }

  get isSignedIn(): boolean {
    return this.$store.state.user.signedIn
  }

  get idToken(): string | null {
    return this.$store.state.user.idToken
  }

  googleLoginParams = {
    client_id: process.env.VUE_APP_GOOGLE_OAUTH_CLIENT_ID
  }
}
</script>

<style lang="scss">
@import "~bootstrap/scss/bootstrap";

#signInWidget  {
  display: inline;
  white-space: nowrap;
  button {
    border: none;
    background: none;
    color: $navbar-dark-color;
    font-weight: bold;
  }
  button:hover {
    color: $navbar-dark-hover-color;
  }
}
</style>
