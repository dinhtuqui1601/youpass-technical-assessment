export function getShortUrl(code: string): string {
  return `${window.location.origin}/s/${code}`
}
