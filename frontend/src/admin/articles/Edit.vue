<template>
  <div>
    <div v-if="loading">
      <loading/>
    </div>
    <div v-else>
      <h3>Edit Article</h3>
      <b-form class="articleEditor" v-on:submit.prevent="save">
        <div>
          <label for="articleSlug">Slug:</label>
          <b-form-input id="articleSug" v-model="slug" :disabled="slugLocked"></b-form-input>
        </div>
        <div>
          <label for="articleTitle">Title:</label>
          <b-form-input id="articleTitle" v-model="title"></b-form-input>
        </div>
        <div class="form-check-inline">
          <b-form-checkbox id="publish" v-model="publish"/>
          <label class="form-check-label" for="publish">Publish</label>
        </div>
        <div v-if="publish" class="form-group">
          <label for="publishDate">Publication Date:</label>
          <b-form-input id="publishDate" v-model="publishDateStr"></b-form-input>
        </div>
        <div class="row">
          <div class="col-sm" id="articleEdit">
            <div class="row">
              <div class="col-sm"><h4>Article Content</h4></div>
              <div class="col-sm align-right" v-if="!showPreview"><a href="#" v-on:click="showPreview=true">Show Preview
                &gt;&gt;</a></div>
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
  </div>
</template>

<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import Render from '@/admin/articles/Render.vue'
import { ArticleSaveEventType, getArticle, saveArticle } from '@/admin/articles/articles'
import Loading from '@/app/Loading.vue'

@Component({
  components: {
    Loading,
    Render
  }
})
export default class Edit extends Vue {
  title: string = ''
  slug: string = ''
  article: string = ''
  publish: boolean = false
  publishDateStr: string = ''
  showPreview: boolean = true
  dirty: boolean = false
  slugLocked: boolean = false
  loading: boolean = false

  get preventSave(): boolean {
    return !this.dirty || this.title === '' || this.slug === '' || this.article === '' || (this.publish && !this.publishDate)
  }

  get publishDate(): Date | undefined {
    const seconds = Date.parse(this.publishDateStr)
    if (seconds) {
      return new Date(seconds)
    }
  }

  @Watch('title')
  flagDirtyTitle() {
    this.dirty = true
  }

  @Watch('slug')
  flagDirtySlug() {
    this.dirty = true
  }

  @Watch('publishDate')
  flagDirtyPubDate() {
    this.dirty = true
  }

  @Watch('article')
  flagDirtyArticle() {
    this.dirty = true
  }

  mounted() {
    this.routeParamsChanged()
  }

  @Watch('$route')
  async routeParamsChanged() {
    const slug = this.$route.params.slug
    if (!slug) {
      return
    }
    this.loading = true
    const article = await getArticle(this.$store.state.user.authToken, slug)
    this.slug = slug
    this.slugLocked = true
    this.title = article.title
    this.article = article.content
    if (article.publishDate) {
      this.publish = true
      this.publishDateStr = article.publishDate as any as string
    }
    this.loading = false
  }

  async save() {
    this.loading = true
    this.dirty = false
    this.slugLocked = true
    const res = await saveArticle(this.$store.state.user.authToken, {
      slug: this.slug,
      title: this.title,
      content: this.article,
      publishDate: this.publishDate
    })
    if (res.eventType === ArticleSaveEventType.CREATED) {
      console.log(`new article created @ ${res.location}`)
    } else {
      console.log(`article updated`)
    }
    this.loading = false
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
