<template>
  <div>
    <h3>Edit Article</h3>
    <b-form class="articleEditor" v-on:submit.prevent="save">
      <div>
        <label for="articleTitle">Title:</label>
        <b-form-input id="articleTitle" v-model="title"></b-form-input>
      </div>
      <div>
        <label for="articleSlug">Slug:</label>
        <b-form-input id="articleSug" v-model="slug"></b-form-input>
      </div>
      <div class="row">
        <div class="col-sm" id="articleEdit">
          <div class="row">
            <div class="col-sm"><h4>Article Content</h4></div>
            <div class="col-sm align-right" v-if="!showPreview"><a href="#" v-on:click="showPreview=true">Show Preview &gt;&gt;</a></div>
          </div>
          <b-textarea v-model="article" id="articleInput"></b-textarea>
        </div>
        <div v-if="showPreview" class="col-sm">
          <div class="row">
            <div class="col-sm"><h4>Preview</h4></div>
            <div class="col-sm align-right"><a href="#" v-on:click="showPreview=false">&lt;&lt; Hide Preview</a></div>
          </div>
          <render :template="article" id="articlePreview"/>
        </div>
      </div>
      <div class="row pt-2">
        <div class="col-sm">
          <b-button variant="primary" v-on:click="save" :disabled="preventSave">Save</b-button>
        </div>
      </div>
    </b-form>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Render from '@/admin/articles/Render.vue'

@Component({
  components: {
    Render
  }
})
export default class Edit extends Vue {
  title: String = ''
  slug: String = ''
  article: String = ''
  showPreview: boolean = true

  get preventSave(): boolean {
    return this.title === '' || this.slug === '' || this.article === ''
  }

  save() {
    alert('would save')
  }
}
</script>

<style lang="scss">
#articleInput {
  height: 550px;
  overflow: scroll;
}

#articlePreview {
  height: 550px;
  overflow: auto;
}
</style>
