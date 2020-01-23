<template>
  <div class="container-fluid">
    <div class="row">
      <nav id="articlesNav" class="col-md-3 bg-light order-1 order-md-0">
        <div class="sidebar-sticky">
          <h3>Articles</h3>
          <ul class="nav flex-column">
            <li class="nav-item">
              <router-link :to="{name: 'article', params: {id: 1}}" class="nav-link">Setting up a support email address</router-link>
              <router-link :to="{name: 'article', params: {id: 0}}" class="nav-link">Spinning up sabadoscodes.com</router-link>
            </li>
          </ul>
        </div>
      </nav>
      <main class="col-md-9 order-0 order-md-1" role="main">
        <router-view />
      </main>
    </div>
  </div>
</template>

<style lang="scss">
#articlesNav .router-link-active {
  font-style: italic;
  font-weight: bold;
}
</style>
