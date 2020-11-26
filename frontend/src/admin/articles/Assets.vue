<template>
  <div>
    <h2>Assets</h2>
    <router-view />
    <router-link :to="{name: 'adminAssetUpload'}">Upload</router-link>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component({})
export default class Assets extends Vue {

}
</script>

<style lang="scss">

</style>
