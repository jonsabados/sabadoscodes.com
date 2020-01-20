import { ActionContext } from 'vuex'
import { VuexModule } from 'vuex-module-decorators'

export enum ActionType {
  Commit,
  Dispatch
}

export interface ContextAction {
  type: ActionType,
  name: string,
  payload: any
}

export default class MockStoreContext<A, B> {
  actionsSent:Array<ContextAction> = []

  commit(name: string, payload: any) {
    this.actionsSent.push({
      type: ActionType.Commit,
      name: name,
      payload: payload
    })
  }

  async dispatch(name: string, payload: any) {
    this.actionsSent.push({
      type: ActionType.Dispatch,
      name: name,
      payload: payload
    })
    return null
  }

  attachToComponent(c: any) {
    c.context = this
  }
}
