import { apiBase } from '@/api/api'
import axios from 'axios'

export interface AssetListDto {
  path: string
  size: number
  url: string
}

export async function listAssets(authToken: string): Promise<Array<AssetListDto>> {
  const endpoint = `${apiBase()}/article/asset`
  const res = await axios.get(endpoint, {
    headers: {
      'Authorization': authToken
    }
  })
  return res.data.results
}

export interface AssetCreationDto {
  path: string,
  mimeType: string,
  content: string
}

export async function uploadAsset(authToken: string, asset: AssetCreationDto, handleProgress?: (progressEvent: ProgressEvent) => void) :Promise<string> {
  const endpoint = `${apiBase()}/article/asset`
  const res = await axios.post(endpoint, JSON.stringify(asset), {
    headers: {
      'Authorization': authToken
    },
    onUploadProgress: handleProgress
  })
  return res.headers['location']
}
