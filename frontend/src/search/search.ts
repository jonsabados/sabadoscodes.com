import { Route } from 'vue-router/types/router'

export interface SearchResult {
  id: number
  title: string
  match: string
}

export function readQuery(route: Route): string {
  const queryValue = route.query['query']
  return queryValue ? queryValue as string : ''
}

export function executeSearch(query: string): Promise<Array<SearchResult>> {
  return new Promise<Array<SearchResult>>((resolve, reject) => {
    setTimeout(() => {
      const roll = Math.floor(Math.random() * 10)
      if (roll === 1) {
        reject(new Error('Simulating a failure'))
      } else {
        const ret = []
        for (let i = 0; i < roll; i++) {
          ret.push(
            {
              id: i,
              title: `would search for ${query}. Not yet implemented`,
              match: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum'
            }
          )
        }
        resolve(ret)
      }
    }, 300)
  })
}
