<template>
  <div class="article">
    <vue-markdown :source="template"/>
  </div>
</template>

<script lang="ts">
import { Component, Prop, Vue } from 'vue-property-decorator'
import VueMarkdown from 'vue-markdown'

@Component({
  components: {
    VueMarkdown
  }
})
export default class Render extends Vue {
  @Prop({ required: true })
  public template!: String
}
</script>

<style lang="scss">

</style>
