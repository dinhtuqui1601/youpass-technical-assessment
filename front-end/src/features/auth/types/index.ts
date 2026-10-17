export interface AuthUser {
  id: string
  email: string
  name: string
  role: string
  createdAt: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  name: string
}

export interface LoginResponse {
  accessToken: string
  expiresIn: number
  user: AuthUser
}
