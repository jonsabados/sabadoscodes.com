import { apiBase } from '@/api/api'
import axios from 'axios'

export interface Article {
  slug: string
  title: string
  content: string
  publishDate?: Date
}

export enum ArticleSaveEventType {
  CREATED = 0,
  UPDATED = 1
}

export interface ArticleCreationResult {
  eventType: ArticleSaveEventType,
  location?: string
}

export async function getArticle(authToken: string, slug: string): Promise<Article> {
  const endpoint = `${apiBase()}/article/slug/${slug}`
  const res = await axios.get(endpoint, {
    headers: {
      'Authorization': authToken
    }
  })
  return res.data
}

export async function listArticles(authToken: string, published: boolean): Promise<Array<Article>> {
  const endpoint = `${apiBase()}/article/?published=${published}`
  const res = await axios.get(endpoint, {
    headers: {
      'Authorization': authToken
    }
  })
  return res.data.results
}

export async function saveArticle(authToken: string, article: Article): Promise<ArticleCreationResult> {
  const endpoint = `${apiBase()}/article/slug/${article.slug}`
  const data = {
    title: article.title,
    content: article.content,
    publishDate: article.publishDate
  }
  const res = await axios.put(endpoint, data, {
    headers: {
      'Authorization': authToken
    }
  })
  if (res.status === 201) {
    return {
      eventType: ArticleSaveEventType.CREATED,
      location: res.headers['location']
    }
  } else {
    return { eventType: ArticleSaveEventType.UPDATED }
  }
}
