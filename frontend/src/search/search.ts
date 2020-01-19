export function executeSearch(query: string) {
  // until we implement a search thing just send the user to google
  window.open('https://google.com/?q=' + encodeURIComponent('site:sabadoscodes.com ' + query), '_blank ')
}
