<template>
  <div>
    <h3>Asset Upload</h3>
    <form v-on:submit.prevent="uploadAsset">
      <div class="form-group">
        <b-form-file
          v-model="asset"
          :state="hasFile"
          placeholder="Choose a file or drop it here..."
          drop-placeholder="Drop file here..."
        ></b-form-file>
      </div>
      <div class="form-group">
        <label for="assetPath">Target Path</label>
        <b-form-input type="text" class="form-control" id="assetPath" placeholder="some/file.png" v-model="assetPath"></b-form-input>
      </div>
      <button class="btn btn-primary" type="submit" :disabled="disableUpload" id="uploadButton">Upload</button>
    </form>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { uploadAsset } from '@/admin/articles/assets'

@Component({})
export default class AssetUpload extends Vue {
  asset: File | null = null
  assetPath: string = ''

  get hasFile(): boolean {
    return this.asset !== null
  }

  get disableUpload(): boolean {
    return !this.hasFile || this.assetPath === ''
  }

  uploadAsset() {
    if (!this.asset) {
      throw Error('asset not set on upload')
    }
    const authToken = this.$store.state.user.authToken
    const assetType = this.asset.type
    const assetPath = this.assetPath
    const reader = new FileReader()
    reader.onload = async function() {
      const content = btoa(String.fromCharCode(...new Uint8Array(this.result as ArrayBuffer)))
      const location = await uploadAsset(authToken, {
        content: content,
        mimeType: assetType,
        path: assetPath
      })
      alert(`asset uploaded to ${location}`)
    }
    reader.readAsArrayBuffer(this.asset)
  }
}
</script>

<style lang="scss">

</style>
