import { defineStore } from 'pinia'
import axios from 'axios'

const API_BASE = import.meta.env.VITE_API_BASE || 'http://localhost:8080'

const api = axios.create({
  baseURL: API_BASE,
  timeout: 10000
})

// Request interceptor - add token
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => Promise.reject(error)
)

// Response interceptor - handle errors
api.interceptors.response.use(
  response => response.data,
  error => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error.response?.data || error)
  }
)

export const useApiStore = defineStore('api', {
  state: () => ({
    products: [],
    searchResults: [],
    orders: [],
    cart: [],
    user: null,
    token: localStorage.getItem('token') || '',
    loading: false
  }),

  getters: {
    isLoggedIn: state => !!state.token,
    cartTotal: state => state.cart.reduce((sum, item) => sum + item.price * item.quantity, 0),
    cartCount: state => state.cart.reduce((sum, item) => sum + item.quantity, 0)
  },

  actions: {
    // ==================== Auth ====================
    async login(username, password) {
      const res = await api.post('/api/v1/auth/login', { username, password })
      this.token = res.data.token
      this.user = res.data.user
      localStorage.setItem('token', this.token)
      return res.data
    },

    async register(username, password, email) {
      return await api.post('/api/v1/auth/register', { username, password, email })
    },

    logout() {
      this.token = ''
      this.user = null
      this.cart = []
      localStorage.removeItem('token')
    },

    // ==================== Products ====================
    async fetchProducts(page = 1, pageSize = 10) {
      this.loading = true
      try {
        const res = await api.get('/api/v1/products', { params: { page, page_size: pageSize } })
        this.products = res.data.products || res.data
        return this.products
      } finally {
        this.loading = false
      }
    },

    async getProduct(id) {
      const res = await api.get(`/api/v1/products/${id}`)
      return res.data
    },

    // ==================== Search ====================
    async semanticSearch(query) {
      this.loading = true
      try {
        const res = await api.get('/api/v1/search/semantic', { params: { q: query } })
        this.searchResults = res.data.products || res.data
        return this.searchResults
      } finally {
        this.loading = false
      }
    },

    // ==================== Orders ====================
    async createOrder(items) {
      const res = await api.post('/api/v1/orders', { items })
      // Clear cart after successful order
      this.cart = []
      return res.data
    },

    async fetchOrders(page = 1, pageSize = 10) {
      this.loading = true
      try {
        const res = await api.get('/api/v1/orders', { params: { page, page_size: pageSize } })
        this.orders = res.data.orders || res.data
        return this.orders
      } finally {
        this.loading = false
      }
    },

    // ==================== Cart ====================
    addToCart(product, quantity = 1) {
      const existingItem = this.cart.find(item => item.id === product.id)
      
      if (existingItem) {
        existingItem.quantity += quantity
      } else {
        this.cart.push({
          id: product.id,
          name: product.name,
          price: product.price,
          image_url: product.image_url,
          quantity
        })
      }
    },

    removeFromCart(productId) {
      this.cart = this.cart.filter(item => item.id !== productId)
    },

    updateCartQuantity(productId, quantity) {
      const item = this.cart.find(item => item.id === productId)
      if (item) {
        item.quantity = quantity
        if (item.quantity <= 0) {
          this.removeFromCart(productId)
        }
      }
    },

    clearCart() {
      this.cart = []
    }
  }
})
