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

export type progressHandler = (progressEvent: ProgressEvent) => void

export type uploadCompleteCallback = (location: string) => void

export function uploadAsset(authToken: string, file: File, path: string, onComplete: uploadCompleteCallback, handleProgress?: progressHandler) : void {
  const endpoint = `${apiBase()}/article/asset`
  const reader = new FileReader()

  reader.onload = async function() {
    const content = (this.result as string).split(',')[1]
    const payload = {
      content: content,
      mimeType: file.type,
      path: path
    }
    const res = await axios.post(endpoint, JSON.stringify(payload), {
      headers: {
        'Authorization': authToken
      },
      onUploadProgress: handleProgress
    })
    onComplete(res.headers['location'])
  }
  reader.readAsDataURL(file)
}
