import MockStoreContext, { ActionType } from '../MockStoreContext'
import { SearchStore } from '@/search/SearchStore'
import { executeSearch } from '@/search/search'

jest.mock('@/search/search')

const executeSearchMock = executeSearch as unknown as jest.Mock

describe('SearchStore', () => {
  describe('doSearch', () => {
    beforeEach(() => {
      executeSearchMock.mockClear()
    })

    it('does happy path searches correctly', async() => {
      const mockContext = new MockStoreContext()

      const testInstance = new SearchStore({})
      mockContext.attachToComponent(testInstance)

      const search = 'testing for fun and profit'
      const expectedResult = [
        {
          id: 1,
          title: 'foo'
        }
      ]

      executeSearchMock.mockImplementation(async(query:string) => {
        expect(query).toEqual(search)
        expect(mockContext.actionsSent).toEqual([
          {
            type: ActionType.Commit,
            name: 'setQuery',
            payload: search
          },
          {
            type: ActionType.Commit,
            name: 'setSearchResults',
            payload: null
          }
        ])
        return expectedResult
      })

      const res = await testInstance.doSearch(search)

      expect(res).toEqual(expectedResult)
      expect(mockContext.actionsSent).toEqual([
        {
          type: ActionType.Commit,
          name: 'setQuery',
          payload: search
        },
        {
          type: ActionType.Commit,
          name: 'setSearchResults',
          payload: null
        },
        {
          type: ActionType.Commit,
          name: 'setSearchResults',
          payload: expectedResult
        }
      ])
    })

    it('gracefully handles errors', async() => {
      const mockContext = new MockStoreContext()

      const testInstance = new SearchStore({})
      mockContext.attachToComponent(testInstance)

      const search = 'testing for fun and profit'
      const expectedResult = new Error('BaNG!')

      executeSearchMock.mockImplementation(async(query:string) => {
        expect(query).toEqual(search)
        expect(mockContext.actionsSent).toEqual([
          {
            type: ActionType.Commit,
            name: 'setQuery',
            payload: search
          },
          {
            type: ActionType.Commit,
            name: 'setSearchResults',
            payload: null
          }
        ])
        throw expectedResult
      })

      const res = await testInstance.doSearch(search)

      expect(res).toEqual(expectedResult)
      expect(mockContext.actionsSent).toEqual([
        {
          type: ActionType.Commit,
          name: 'setQuery',
          payload: search
        },
        {
          type: ActionType.Commit,
          name: 'setSearchResults',
          payload: null
        },
        {
          type: ActionType.Dispatch,
          name: 'registerRemoteError',
          payload: expectedResult
        },
        {
          type: ActionType.Commit,
          name: 'setSearchResults',
          payload: []
        }
      ])
    })
  })

  describe('clearSearchResults', () => {
    it('it clears the query and results', async() => {
      const mockContext = new MockStoreContext()

      const testInstance = new SearchStore({})
      mockContext.attachToComponent(testInstance)

      await testInstance.clearSearchResults()
      expect(mockContext.actionsSent).toEqual([
        {
          type: ActionType.Commit,
          name: 'setQuery',
          payload: ''
        },
        {
          type: ActionType.Commit,
          name: 'setSearchResults',
          payload: null
        }
      ])
    })
  })
})
