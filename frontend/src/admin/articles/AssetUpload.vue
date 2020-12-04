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
          :disabled="uploading"
        ></b-form-file>
      </div>
      <div class="form-group">
        <label for="assetPath">Target Path</label>
        <b-form-input type="text" class="form-control" id="assetPath" placeholder="some/file.png" v-model="assetPath" :disabled="uploading"></b-form-input>
      </div>
      <button v-if="!uploading" class="btn btn-primary" type="submit" :disabled="disableUpload" id="uploadButton">Upload</button>
      <b-alert v-else-if="newAssetUrl" show dismissible variant="success" role="alert" @dismissed="completeUpload">
        Asset uploaded to <a :href="newAssetUrl" target="_blank">{{ newAssetUrl }}</a>
      </b-alert>
      <b-progress v-else :value="uploadProgress" :max="uploadSize" show-progress animated></b-progress>
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
  uploading: boolean = false
  uploadSize: number = 1
  uploadProgress: number = 0
  newAssetUrl: string | null = null

  get hasFile(): boolean {
    return this.asset !== null
  }

  get disableUpload(): boolean {
    return !this.hasFile || this.assetPath === ''
  }

  handleUploadProgress(progressEvent: ProgressEvent) {
    this.uploadSize = progressEvent.total
    this.uploadProgress = progressEvent.loaded
  }

  completeUpload() {
    this.uploading = false
    this.asset = null
    this.assetPath = ''
    this.uploadSize = 1
    this.uploadProgress = 0
    this.newAssetUrl = null
  }

  uploadFinished(assetURL: string) {
    this.newAssetUrl = assetURL
  }

  uploadAsset() {
    if (!this.asset) {
      throw Error('asset not set on upload')
    }
    this.uploading = true
    uploadAsset(this.$store.state.user.authToken, this.asset, this.assetPath, this.uploadFinished, this.handleUploadProgress)
  }
}
</script>

<style lang="scss">

</style>
