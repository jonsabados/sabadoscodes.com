<template>
  <main role="main">
    <article-zero v-if="articleId == 0"/>
    <article-one v-else-if="articleId == 1"/>
    <div v-else>
      <h1>Article Not Found</h1>
      <p>No article having an id of {{ articleId }} is present in the system.</p>
    </div>
  </main>
</template>

<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import ArticleZero from './ArticleZero.vue'
import ArticleOne from './ArticleOne.vue'

@Component({
  components: { ArticleZero, ArticleOne }
})
export default class Article extends Vue {
  articleId: number | null = null

  mounted() {
    this.routeUpdated()
  }

  @Watch('$route')
  routeUpdated() {
    try {
      this.articleId = parseInt(this.$route.params.id)
    } catch {
      // don't care
    }
  }
}
</script>

<style lang="scss">
.code-block {
  color: #e83e8c;
}
</style>
