<template>
  <div class="product-detail-page">
    <el-page-header @back="$router.back()" content="商品详情" />
    
    <div v-loading="loading" class="product-content">
      <el-empty v-if="!loading && !product" description="商品不存在" />
      
      <div v-else-if="product" class="product-main">
        <div class="product-image">
          <img :src="product.image_url || 'https://via.placeholder.com/400'" :alt="product.name" />
        </div>
        
        <div class="product-info">
          <h1>{{ product.name }}</h1>
          <p class="description">{{ product.description }}</p>
          
          <div class="price-section">
            <span class="price">¥{{ product.price }}</span>
            <span class="stock">库存: {{ product.stock }}</span>
          </div>
          
          <div class="quantity-section">
            <span>数量:</span>
            <el-input-number v-model="quantity" :min="1" :max="product.stock" />
          </div>
          
          <div class="action-buttons">
            <el-button type="primary" size="large" @click="handleBuy">
              立即购买
            </el-button>
            <el-button size="large" @click="handleAddToCart">
              加入购物车
            </el-button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useApiStore } from '../stores/api'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const apiStore = useApiStore()

const product = ref(null)
const loading = ref(true)
const quantity = ref(1)

onMounted(async () => {
  const productId = route.params.id
  try {
    const products = await apiStore.fetchProducts()
    product.value = products.find(p => p.id === parseInt(productId))
  } catch (err) {
    ElMessage.error('获取商品失败')
  }
  loading.value = false
})

const handleBuy = async () => {
  if (!apiStore.token) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }

  try {
    await apiStore.createOrder([{ 
      product_id: product.value.id, 
      quantity: quantity.value 
    }])
    ElMessage.success('购买成功！')
    router.push('/orders')
  } catch (err) {
    ElMessage.error(err.message || '购买失败')
  }
}

const handleAddToCart = () => {
  ElMessage.info('购物车功能开发中')
}
</script>

<style scoped>
.product-detail-page {
  min-height: 100vh;
  background: #f5f7fa;
  padding: 20px;
}

.product-content {
  max-width: 1200px;
  margin: 20px auto;
}

.product-main {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 40px;
  background: white;
  padding: 40px;
  border-radius: 12px;
}

.product-image img {
  width: 100%;
  border-radius: 8px;
}

.product-info h1 {
  font-size: 28px;
  margin-bottom: 16px;
}

.product-info .description {
  color: #666;
  line-height: 1.8;
  margin-bottom: 24px;
}

.price-section {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 24px;
}

.price-section .price {
  font-size: 32px;
  color: #f56c6c;
  font-weight: bold;
}

.price-section .stock {
  color: #999;
}

.quantity-section {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
}

.action-buttons {
  display: flex;
  gap: 12px;
}
</style>
