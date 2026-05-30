export interface ApiResponse<T> {
  code: number
  data: T
  msg: string
}

export interface ListResponse<T> {
  list: T[]
  total: number
}
