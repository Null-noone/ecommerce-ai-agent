<template>
  <div class="orders-page">
    <el-page-header @back="$router.push('/')" content="我的订单" />
    
    <div class="orders-content">
      <el-empty v-if="loading" description="加载中..." />
      <el-empty v-else-if="orders.length === 0" description="暂无订单" />
      
      <el-card v-for="order in orders" :key="order.id" class="order-card">
        <template #header>
          <div class="order-header">
            <span>订单号: {{ order.id }}</span>
            <el-tag :type="getStatusType(order.status)">{{ order.status }}</el-tag>
          </div>
        </template>
        
        <div class="order-items">
          <div v-for="item in order.items" :key="item.id" class="order-item">
            <span>商品ID: {{ item.product_id }} x {{ item.quantity }}</span>
            <span>¥{{ item.price }}</span>
          </div>
        </div>
        
        <div class="order-footer">
          <span class="total">合计: ¥{{ order.total_amount }}</span>
          <span class="date">{{ formatDate(order.created_at) }}</span>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useApiStore } from '../stores/api'

const apiStore = useApiStore()
const orders = ref([])
const loading = ref(true)

onMounted(async () => {
  try {
    orders.value = await apiStore.fetchOrders()
  } catch (err) {
    console.error(err)
  }
  loading.value = false
})

const getStatusType = (status) => {
  const map = { pending: 'warning', paid: 'success', shipped: 'info' }
  return map[status] || ''
}

const formatDate = (date) => {
  return new Date(date).toLocaleString('zh-CN')
}
</script>

<style scoped>
.orders-page {
  min-height: 100vh;
  background: #f5f7fa;
  padding: 20px;
}

.orders-content {
  max-width: 800px;
  margin: 20px auto;
}

.order-card {
  margin-bottom: 16px;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.order-item {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #eee;
}

.order-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 2px solid #eee;
}

.total {
  font-size: 18px;
  font-weight: bold;
  color: #f56c6c;
}

.date {
  color: #999;
}
</style>
