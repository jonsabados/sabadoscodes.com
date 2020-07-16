import { apiBase } from '@/api/api'
import axios from 'axios'

export interface Self {
  userId: string
  email: string
  name: string
  roles: null|Array<string>
}

export async function fetchSelf(authToken: string):Promise<Self> {
  const endpoint = `${apiBase()}/self`
  const res = await axios.get<Self>(endpoint, {
    headers: {
      'Authorization': authToken
    }
  })
  return res.data
}
