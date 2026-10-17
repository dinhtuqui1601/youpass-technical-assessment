// In-memory access token holder. Tokens are intentionally NOT persisted to
// LocalStorage / SessionStorage.
let accessToken: string | null = null
let expiresAt: number | null = null

export const tokenManager = {
  get(): string | null {
    return accessToken
  },
  set(token: string, expiresInSec?: number): void {
    accessToken = token
    expiresAt = expiresInSec ? Date.now() + expiresInSec * 1000 : null
  },
  clear(): void {
    accessToken = null
    expiresAt = null
  },
  getExpiresAt(): number | null {
    return expiresAt
  },
}
