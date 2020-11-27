import { apiBase } from '@/api/api'
import axios from 'axios'

export interface Asset {
  path: string,
  mimeType: string,
  content: string
}

export async function uploadAsset(authToken: string, asset: Asset, handleProgress?: (progressEvent: ProgressEvent) => void) :Promise<string> {
  const endpoint = `${apiBase()}/article/asset`
  const res = await axios.post(endpoint, JSON.stringify(asset), {
    headers: {
      'Authorization': authToken
    },
    onUploadProgress: handleProgress
  })
  return res.headers['location']
}
