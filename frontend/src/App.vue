<template>
  <div id="app">
    <main-nav />
    <div>
      <router-view/>
    </div>
    <footer class="footer bg-light" id="footer">
        <nav role="navigation" class="navbar navbar-expand" id="footerNav">
          <ul class="navbar-nav">
            <li class="nav-item">
              <router-link :to="{name: 'privacy'}" class="nav-link">Privacy Policy</router-link>
            </li>
          </ul>
        </nav>
    </footer>
    <b-modal v-model="hasRemoteError" role="alert" ok-only id="remoteErrorDialog" title="Something went wrong">
      <p id="remoteErrorMessage">Please try again.</p>
    </b-modal>
    <tracking-consent />
  </div>
</template>

<script lang="ts">
import { Vue, Component } from 'vue-property-decorator'
import MainNav from './navigation/MainNav.vue'
import { AppStore } from '@/app/AppStore'
import TrackingConsent from '@/legal/TrackingConsent.vue'
import { UserStore } from '@/user/UserStore'

@Component({
  components: {
    MainNav,
    TrackingConsent
  }
})
export default class App extends Vue {
  created() {
    this.$store.dispatch(UserStore.ACTION_INITIALIZE)
  }

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
html {
  position: relative;
  min-height: 100%;
}
body {
  margin-bottom: 60px;
}
.footer {
  position: absolute;
  bottom: 0;
  width: 100%;
  height: 60px;
  line-height: 30px;
}

#footerNav {
  height: 60px;
  padding: 0;

  .navbar-nav {
    margin-left: auto;
    margin-right: auto;
  }
}
</style>
