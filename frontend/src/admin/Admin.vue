<template>
  <div>
    <h1>Admin</h1>
    <main class="col-md-9 order-0 order-md-1" role="main">
      <router-view />
    </main>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component({})
export default class Admin extends Vue {

}
</script>

<style lang="scss">

</style>
