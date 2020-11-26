import { apiBase } from '@/api/api'
import axios from 'axios'

export interface Asset {
  path: string,
  mimeType: string,
  content: string
}

export async function uploadAsset(authToken: string, asset: Asset):Promise<string> {
  const endpoint = `${apiBase()}/article/asset`
  const res = await axios.post(endpoint, JSON.stringify(asset), {
    headers: {
      'Authorization': authToken
    }
  })
  return res.headers['location']
}
