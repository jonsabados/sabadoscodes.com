<template>
  <div>
    <loading v-if="!assets"/>
    <table v-else class="table table-striped">
      <thead>
        <tr>
          <th scope="col">Asset</th>
          <th scope="col">Size</th>
          <th scope="col">URL</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="asset in assets" :key="asset.path">
          <th scope="row">{{ asset.path }}</th>
          <td>{{ asset.size }}</td>
          <td><a :href="asset.url" target="_blank">{{ asset.url }}</a></td>
        </tr>
      </tbody>
    </table>
    <br/>
    <router-link :to="{name: 'adminAssetUpload'}">Upload New Asset</router-link>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { AssetListDto, listAssets } from '@/admin/articles/assets'
import Loading from '@/app/Loading.vue'

@Component({
  components: {
    Loading
  }
})
export default class AssetsList extends Vue {
  assets: Array<AssetListDto> | null = null

  async mounted() {
    this.assets = await listAssets(this.$store.state.user.authToken)
  }
}
</script>

<style lang="scss">

</style>
