<template>
  <div>
    <nav aria-label="breadcrumb">
      <ol class="breadcrumb">
        <li class="breadcrumb-item active" aria-current="page" v-for="bc in breadcrumbs" :key="bc.route">
          <router-link :to="bc.path" class="nav-link">{{ bc.name }}</router-link>
        </li>
      </ol>
    </nav>
    <main role="main">
      <loading v-if="!isReady" />
      <router-view v-else />
    </main>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Loading from '@/app/Loading.vue'

interface BreadCrumb {
  name: string
  path: string
}

@Component({
  components: {
    Loading
  }
})
export default class Admin extends Vue {
  get isReady(): boolean {
    return this.$store.state.user.authToken !== 'anonymous'
  }

  get breadcrumbs(): Array<BreadCrumb> {
    return this.$route.matched.filter(r => r.meta && r.meta.breadCrumb).map(r => {
      return {
        name: r.meta.breadCrumb as string,
        path: r.path as string
      }
    })
  }
}
</script>

<style lang="scss">

</style>
