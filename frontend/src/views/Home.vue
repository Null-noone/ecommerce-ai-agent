<template>
  <div class="home">
    <!-- Header -->
    <header class="header">
      <div class="logo">🛒 电商AI</div>
      <div class="header-actions">
        <el-button v-if="!apiStore.token" type="primary" @click="$router.push('/login')">登录</el-button>
        <el-button v-if="!apiStore.token" @click="$router.push('/register')">注册</el-button>
        <el-dropdown v-if="apiStore.token" trigger="click">
          <el-button type="primary">
            我的 <el-icon><arrow-down /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="$router.push('/orders')">我的订单</el-dropdown-item>
              <el-dropdown-item divided @click="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </header>

    <!-- Hero Search -->
    <section class="hero">
      <h1>智能购物助手</h1>
      <p>描述你的需求，AI帮你推荐最合适的商品</p>
      <div class="search-box">
        <el-input
          v-model="searchQuery"
          placeholder="给我推荐适合送女生的口红"
          size="large"
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
          <template #append>
            <el-button type="primary" @click="handleSearch">搜索</el-button>
          </template>
        </el-input>
      </div>
    </section>

    <!-- Products Grid -->
    <section class="products">
      <h2>{{ searchQuery ? '搜索结果' : '热门商品' }}</h2>
      <div v-loading="loading" class="product-grid">
        <el-empty v-if="!loading && displayProducts.length === 0" description="暂无商品" />
        <el-card
          v-for="product in displayProducts"
          :key="product.id"
          class="product-card"
          shadow="hover"
          @click="showProductDetail(product)"
        >
          <div class="product-image">
            <img :src="product.image_url || 'https://via.placeholder.com/200'" :alt="product.name" />
          </div>
          <div class="product-info">
            <h3>{{ product.name }}</h3>
            <p class="description">{{ product.description }}</p>
            <div class="price-row">
              <span class="price">¥{{ product.price }}</span>
              <span class="stock">库存: {{ product.stock }}</span>
            </div>
          </div>
        </el-card>
      </div>
    </section>

    <!-- AI Chat Widget -->
    <ChatWidget />

    <!-- Product Detail Dialog -->
    <el-dialog v-model="dialogVisible" title="商品详情" width="500px">
      <div v-if="selectedProduct" class="product-detail">
        <img :src="selectedProduct.image_url || 'https://via.placeholder.com/300'" />
        <h2>{{ selectedProduct.name }}</h2>
        <p>{{ selectedProduct.description }}</p>
        <div class="detail-row">
          <span class="price">¥{{ selectedProduct.price }}</span>
          <span class="stock">库存: {{ selectedProduct.stock }}</span>
        </div>
        <el-input-number v-model="buyQuantity" :min="1" :max="selectedProduct.stock" />
        <el-button type="primary" size="large" @click="handleBuy">立即购买</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useApiStore } from '../stores/api'
import { ElMessage } from 'element-plus'
import ChatWidget from '../components/ChatWidget.vue'

const router = useRouter()
const apiStore = useApiStore()

const searchQuery = ref('')
const loading = ref(false)
const dialogVisible = ref(false)
const selectedProduct = ref(null)
const buyQuantity = ref(1)

const displayProducts = computed(() => {
  return searchQuery.value ? apiStore.searchResults : apiStore.products
})

onMounted(async () => {
  loading.value = true
  await apiStore.fetchProducts()
  loading.value = false
})

const handleSearch = async () => {
  if (!searchQuery.value.trim()) return
  
  loading.value = true
  try {
    await apiStore.semanticSearch(searchQuery.value)
  } catch (err) {
    ElMessage.error('搜索失败')
  }
  loading.value = false
}

const showProductDetail = (product) => {
  selectedProduct.value = product
  buyQuantity.value = 1
  dialogVisible.value = true
}

const handleBuy = async () => {
  if (!apiStore.token) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }

  try {
    await apiStore.createOrder([{ product_id: selectedProduct.value.id, quantity: buyQuantity.value }])
    ElMessage.success('下单成功！')
    dialogVisible.value = false
    // Refresh stock
    await apiStore.fetchProducts()
  } catch (err) {
    ElMessage.error(err.message || '下单失败')
  }
}

const logout = () => {
  apiStore.logout()
  ElMessage.success('已退出登录')
}
</script>

<style scoped>
.home {
  min-height: 100vh;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 40px;
  background: #fff;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

.logo {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
}

.hero {
  text-align: center;
  padding: 80px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.hero h1 {
  font-size: 48px;
  margin-bottom: 16px;
}

.hero p {
  font-size: 18px;
  margin-bottom: 40px;
  opacity: 0.9;
}

.search-box {
  max-width: 600px;
  margin: 0 auto;
}

.search-box .el-input {
  font-size: 18px;
}

.products {
  padding: 40px;
  max-width: 1200px;
  margin: 0 auto;
}

.products h2 {
  margin-bottom: 24px;
  font-size: 24px;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 20px;
}

.product-card {
  cursor: pointer;
  transition: transform 0.2s;
}

.product-card:hover {
  transform: translateY(-4px);
}

.product-image {
  width: 100%;
  height: 200px;
  overflow: hidden;
  border-radius: 8px;
  background: #f5f7fa;
}

.product-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.product-info {
  padding: 12px 0;
}

.product-info h3 {
  font-size: 16px;
  margin-bottom: 8px;
}

.description {
  font-size: 14px;
  color: #666;
  margin-bottom: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.price-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.price {
  font-size: 20px;
  color: #f56c6c;
  font-weight: bold;
}

.stock {
  font-size: 12px;
  color: #999;
}

.product-detail {
  text-align: center;
}

.product-detail img {
  width: 200px;
  height: 200px;
  object-fit: cover;
  border-radius: 8px;
  margin-bottom: 20px;
}

.product-detail h2 {
  margin-bottom: 12px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  margin: 20px 0;
}
</style>
