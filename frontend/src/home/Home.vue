<template>
  <main class="container-fluid" role="main">
    <h1>sabadoscodes.com</h1>
    <p>
      <span v-if="isSignedIn">Hello {{ name }},</span>
      <span v-else>Hello,</span>
      and welcome to sabadoscodes.com, my personal development playground.
    </p>
    <p>
      Thanks to all the various AWS goodies available its a place where I can mess with both front end<sup>*</sup> and back end
      code on the cheap. The front end is a <a href="https://vuejs.org/" target="_blank">Vue.js</a> application pushed
      to an S3 bucket sitting behind CloudFront. For the backend I'm currently exploring what a restful API might look
      like when its composed entirely of lambdas exposed through
      <a href="https://aws.amazon.com/api-gateway/" target="_blank">API Gateway</a>.
    </p>
    <p>
      If you want to know more about me check out the <router-link :to="{name: 'about'}">about page</router-link>.
    </p>
    <p>
      For those that are curious all the code behind this site & the terraform for the infrastructure is available in
      GitHub <a href="https://github.com/jonsabados/sabadoscodes.com" target="_blank">here</a>.
    </p>
    <sup>
      * Please note that I am not a designer, nor do I have any aspirations to be one. So if this UI makes your eyes
      bleed that is probably why.
    </sup>
  </main>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component({
  components: {
  }
})
export default class SignIn extends Vue {
  get isSignedIn(): boolean {
    return this.$store.state.user.signedIn && this.$store.state.user.self
  }

  get name(): string | null {
    if (this.$store.state.user.self) {
      return this.$store.state.user.self.name
    } else {
      return null
    }
  }
}
</script>

<style lang="scss">

</style>
