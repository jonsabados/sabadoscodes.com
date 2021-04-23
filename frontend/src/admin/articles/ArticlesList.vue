<template>
  <div>
    <b-form-group>
      <b-form-radio-group id="typeToShow" v-model="published">
        <b-form-radio :value="false">Unpublished</b-form-radio>
        <b-form-radio :value="true">Published</b-form-radio>
      </b-form-radio-group>
    </b-form-group>
    <loading v-if="!articles"/>
    <table v-else class="table table-striped">
      <thead>
      <tr>
        <th scope="col">Slug</th>
        <th scope="col">Title</th>
        <th v-if="published" scope="col">Publish Date</th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="article in articles" :key="article.slug">
        <th scope="row">
          <router-link :to="{name: 'adminArticleEdit', params: {slug: article.slug}}">{{ article.slug }}</router-link>
        </th>
        <td>{{ article.title }}</td>
        <td v-if="published">{{ article.publishDate }}</td>
      </tr>
      </tbody>
    </table>
    <br/>
    <router-link :to="{name: 'adminNewArticle'}">Draft new article</router-link>
  </div>
</template>

<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import { Article, listArticles } from '@/admin/articles/articles'
import Loading from '@/app/Loading.vue'

@Component({
  components: {
    Loading
  }
})
export default class ArticlesList extends Vue {
  published: boolean = false

  articles: Array<Article> | null = null

  @Watch('published')
  async fetchArticles() {
    this.articles = null
    this.articles = await listArticles(this.$store.state.user.authToken, this.published)
  }

  mounted() {
    this.fetchArticles()
  }
}
</script>

<style lang="scss">

</style>
