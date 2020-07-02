import { Action, Module, Mutation, VuexModule } from 'vuex-module-decorators'
import { currentUser, GoogleUser, isSignedIn, listenForUser } from '@/user/google'

export interface UserState {
  isReady: boolean
  signedIn: boolean
  idToken: string | null
}

@Module
export class UserStore extends VuexModule<UserState> {
  static ACTION_INITIALIZE = 'initialize'
  static MUTATION_SET_USER = 'setUser'
  static MUTATION_MARK_READY = 'markReady'

  isReady: boolean = false

  signedIn: boolean = false

  idToken: string | null = null

  @Mutation
  markReady() {
    this.isReady = true
  }

  @Mutation
  setUser(user: GoogleUser) {
    if (user.isSignedIn()) {
      this.signedIn = true
      this.idToken = user.getAuthResponse().id_token
      // eslint-disable-next-line
      console.log(`id token: ${this.idToken}`)
    } else {
      this.signedIn = false
      this.idToken = null
    }
  }

  @Action
  async initialize() {
    // missing await is very intentional, don't wanna block
    listenForUser((user) => {
      this.context.commit(UserStore.MUTATION_SET_USER, user)
    })
    const loggedIn = await isSignedIn()
    this.context.commit(UserStore.MUTATION_MARK_READY)
    if (loggedIn) {
      const user = await currentUser()
      this.context.commit(UserStore.MUTATION_SET_USER, user)
    }
  }
}
