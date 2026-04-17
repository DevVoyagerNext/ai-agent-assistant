import axios, { AxiosHeaders } from 'axios'

const request = axios.create({
  baseURL: 'http://localhost:8080/v1',
  timeout: 60000 // 修改为 60 秒，防止 AI 大模型等请求超时导致 context canceled
})

// 是否正在刷新 Token 的标记
let isRefreshing = false
// 重试队列，在刷新 Token 期间的所有请求都会暂存
let requests: any[] = []

const clearAuthAndRedirect = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('refreshToken')
  localStorage.removeItem('expiresAt')
  localStorage.removeItem('user')
  window.location.href = '/login'
}

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
  async response => {
    const res = response.data
    // 如果 code 为 1005，表示短 Token 过期
    if (res.code === 1005) {
      const config = response.config
      
      if (!isRefreshing) {
        isRefreshing = true
        const refreshToken = localStorage.getItem('refreshToken')
        
        if (!refreshToken) {
          clearAuthAndRedirect()
          return Promise.reject(new Error('No refresh token'))
        }

        try {
          // 调用刷新 Token 接口
          const refreshRes = await axios.post('http://localhost:8080/v1/user/refresh-token', {
            refreshToken
          })
          
          if (refreshRes.data?.code === 200 && refreshRes.data.data) {
            const { token, expiresAt } = refreshRes.data.data
            localStorage.setItem('token', token)
            localStorage.setItem('expiresAt', String(expiresAt))
            
            // 执行队列中的请求
            requests.forEach(cb => cb(token))
            requests = []
            
            // 重试当前请求
            const headers = AxiosHeaders.from(config.headers)
            headers.set('x-token', token)
            config.headers = headers
            return request(config)
          } else {
            // 长 Token 可能也过期了
            clearAuthAndRedirect()
            return Promise.reject(new Error('Refresh token failed'))
          }
        } catch (err) {
          clearAuthAndRedirect()
          return Promise.reject(err)
        } finally {
          isRefreshing = false
        }
      } else {
        // 正在刷新中，将请求放入队列并返回 Promise
        return new Promise(resolve => {
          requests.push((newToken: string) => {
            const headers = AxiosHeaders.from(config.headers)
            headers.set('x-token', newToken)
            config.headers = headers
            resolve(request(config))
          })
        })
      }
    }
    return response
  },
  error => {
    // 处理网络错误或 401 等其他错误
    if (error.response?.status === 401) {
      clearAuthAndRedirect()
    }
    return Promise.reject(error)
  }
)

export default request
