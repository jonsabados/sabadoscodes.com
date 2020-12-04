<template>
  <div>
    <h1>Admin</h1>
    <main class="col-md-9 order-0 order-md-1" role="main">
      <loading v-if="!isReady" />
      <router-view v-else />
    </main>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Loading from '@/app/Loading.vue'

@Component({
  components: {
    Loading
  }
})
export default class Admin extends Vue {
  get isReady(): boolean {
    return this.$store.state.user.authToken !== 'anonymous'
  }
}
</script>

<style lang="scss">

</style>
