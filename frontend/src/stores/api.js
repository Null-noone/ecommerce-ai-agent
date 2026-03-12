import { defineStore } from 'pinia'
import axios from 'axios'

const API_BASE = import.meta.env.VITE_API_BASE || 'http://localhost:8080'

const api = axios.create({
  baseURL: API_BASE,
  timeout: 10000
})

// Add token to requests
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

export const useApiStore = defineStore('api', {
  state: () => ({
    products: [],
    searchResults: [],
    orders: [],
    user: null,
    token: localStorage.getItem('token') || ''
  }),

  actions: {
    // Auth
    async login(username, password) {
      const res = await api.post('/api/v1/auth/login', { username, password })
      this.token = res.data.token
      this.user = res.data.user
      localStorage.setItem('token', this.token)
      return res.data
    },

    async register(username, password, email) {
      const res = await api.post('/api/v1/auth/register', { username, password, email })
      return res.data
    },

    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('token')
    },

    // Products
    async fetchProducts() {
      const res = await api.get('/api/v1/products')
      this.products = res.data
      return res.data
    },

    // Semantic Search
    async semanticSearch(query) {
      const res = await api.get('/api/v1/search/semantic', { params: { q: query } })
      this.searchResults = res.data.products
      return res.data
    },

    // Orders
    async createOrder(items) {
      const res = await api.post('/api/v1/orders', { items })
      return res.data
    },

    async fetchOrders() {
      const res = await api.get('/api/v1/orders')
      this.orders = res.data
      return res.data
    }
  }
})
