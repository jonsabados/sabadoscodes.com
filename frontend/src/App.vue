<template>
  <div id="app">
    <main-nav />
    <div role="main">
      <router-view/>
    </div>
    <b-modal v-model="hasRemoteError" role="alert" ok-only id="remoteErrorDialog" title="Something went wrong">
      <p id="remoteErrorMessage">Please try again.</p>
    </b-modal>
  </div>
</template>

<script lang="ts">
import { Vue, Component } from 'vue-property-decorator'
import MainNav from './navigation/MainNav.vue'
import { AppStore } from '@/app/AppStore'

@Component({
  components: {
    MainNav
  }
})
export default class App extends Vue {
  get hasRemoteError(): boolean {
    return this.$store.state.app.errorToAck != null
  }

  set hasRemoteError(error) {
    if (!error) {
      this.$store.dispatch(AppStore.ACTION_ACK_REMOTE_ERROR)
    }
  }
}
</script>

<style lang="scss">
$dark: #223e5d;
@import "~bootstrap/scss/bootstrap";
</style>
