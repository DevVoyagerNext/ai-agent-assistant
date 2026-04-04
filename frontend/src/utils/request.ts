import axios, { AxiosHeaders } from 'axios'

const request = axios.create({
  baseURL: 'http://localhost:8080/v1',
  timeout: 10000
})

// 可以添加拦截器
request.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      const headers = AxiosHeaders.from(config.headers)
      headers.set('x-token', token)
      config.headers = headers
    }
    return config
  },
  error => Promise.reject(error)
)

request.interceptors.response.use(
  response => response,
  error => Promise.reject(error)
)

export default request
