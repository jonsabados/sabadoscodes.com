import { Action, Module, Mutation, VuexModule } from 'vuex-module-decorators'
import { currentUser, GoogleUser, isSignedIn, listenForUser } from '@/user/google'
import { fetchSelf as apiFetchSelf, Self } from '@/user/self'

export interface UserState {
  isReady: boolean
  signedIn: boolean
  idToken: string | null
}

@Module
export class UserStore extends VuexModule<UserState> {
  static ACTION_INITIALIZE = 'initialize'
  static ACTION_FETCH_SELF = 'fetchSelf'
  static MUTATION_SET_GOOGLE_USER = 'setGoogleUser'
  static MUTATION_SET_SELF = 'setSelf'
  static MUTATION_MARK_READY = 'markReady'

  isReady: boolean = false

  signedIn: boolean = false

  authToken: string = 'anonymous'

  self: Self | null = null

  @Mutation
  markReady() {
    this.isReady = true
  }

  @Mutation
  setGoogleUser(user: GoogleUser) {
    if (user.isSignedIn()) {
      this.signedIn = true
      this.authToken = `Bearer ${user.getAuthResponse().id_token}`
    } else {
      this.signedIn = false
      this.authToken = 'anonymous'
    }
  }

  @Mutation
  setSelf(self: Self | null) {
    this.self = self
  }

  @Action
  async initialize() {
    // missing await is very intentional, don't wanna block
    listenForUser((user) => {
      this.context.commit(UserStore.MUTATION_SET_GOOGLE_USER, user)
      this.context.dispatch(UserStore.ACTION_FETCH_SELF)
    })
    const loggedIn = await isSignedIn()
    this.context.commit(UserStore.MUTATION_MARK_READY)
    if (loggedIn) {
      const user = await currentUser()
      this.context.commit(UserStore.MUTATION_SET_GOOGLE_USER, user)
      this.context.dispatch(UserStore.ACTION_FETCH_SELF)
    }
  }

  @Action
  async fetchSelf() {
    const self = await apiFetchSelf(this.authToken)
    this.context.commit(UserStore.MUTATION_SET_SELF, self)
  }
}
