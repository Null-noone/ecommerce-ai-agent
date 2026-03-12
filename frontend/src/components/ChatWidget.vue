<template>
  <div class="chat-widget">
    <!-- Toggle Button -->
    <div v-if="!isOpen" class="chat-toggle" @click="isOpen = true">
      <el-icon size="24"><ChatDotRound /></el-icon>
      <span>AI 助手</span>
    </div>

    <!-- Chat Window -->
    <div v-else class="chat-window">
      <div class="chat-header">
        <span>🧠 AI 购物助手</span>
        <el-button type="danger" size="small" text @click="isOpen = false">
          <el-icon><Close /></el-icon>
        </el-button>
      </div>

      <div class="chat-messages" ref="messagesContainer">
        <div
          v-for="(msg, index) in messages"
          :key="index"
          class="message"
          :class="msg.role"
        >
          <div class="avatar">{{ msg.role === 'user' ? '👤' : '🤖' }}</div>
          <div class="content">
            <div v-if="msg.isStreaming" class="typing">
              <span v-for="(dot, i) in 3" :key="i" class="dot">.</span>
            </div>
            <div v-else v-html="formatMessage(msg.content)"></div>
          </div>
        </div>
      </div>

      <div class="chat-input">
        <el-input
          v-model="inputMessage"
          placeholder="问我任何关于商品的问题..."
          @keyup.enter="sendMessage"
          :disabled="isLoading"
        >
          <template #append>
            <el-button :loading="isLoading" @click="sendMessage">
              <el-icon><Promotion /></el-icon>
            </el-button>
          </template>
        </el-input>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, nextTick, onMounted } from 'vue'
import { useApiStore } from '../stores/api'

const apiStore = useApiStore()
const isOpen = ref(false)
const inputMessage = ref('')
const messages = ref([
  { role: 'assistant', content: '你好！我是你的AI购物助手，有什么可以帮你的？' }
])
const isLoading = ref(false)
const messagesContainer = ref(null)

const scrollToBottom = async () => {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

const formatMessage = (content) => {
  // Simple markdown-like formatting
  return content
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\n/g, '<br>')
    .replace(/- /g, '<br>• ')
}

const sendMessage = async () => {
  if (!inputMessage.value.trim() || isLoading.value) return

  const userMessage = inputMessage.value.trim()
  inputMessage.value = ''

  // Add user message
  messages.value.push({ role: 'user', content: userMessage })
  await scrollToBottom()

  // Add streaming assistant message placeholder
  messages.value.push({ role: 'assistant', content: '', isStreaming: true })
  isLoading.value = true
  await scrollToBottom()

  try {
    // Call AI chat API (simulated for now - needs backend SSE support)
    const sessionId = localStorage.getItem('chat_session') || 
      'session_' + Date.now()
    localStorage.setItem('chat_session', sessionId)

    // Simulate AI response (replace with real SSE call)
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    const responses = [
      '根据您的需求，我推荐以下商品：\n\n• **iPhone 15 Pro** - 适合科技爱好者\n• **Dior 999 口红** - 经典正红色，送礼首选',
      '这款产品非常适合您！它有以下特点：\n\n• 高品质\n• 性价比高\n• 好评如潮',
      '关于您的问题，这款商品支持7天无理由退换，质量有保障哦！'
    ]
    
    const randomResponse = responses[Math.floor(Math.random() * responses.length)]
    
    // Remove streaming indicator and set content
    messages.value[messages.value.length - 1] = { 
      role: 'assistant', 
      content: randomResponse 
    }
    
  } catch (error) {
    messages.value[messages.value.length - 1] = { 
      role: 'assistant', 
      content: '抱歉，请稍后重试。' 
    }
  }

  isLoading.value = false
  await scrollToBottom()
}
</script>

<style scoped>
.chat-widget {
  position: fixed;
  bottom: 20px;
  right: 20px;
  z-index: 1000;
}

.chat-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 30px;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  font-weight: bold;
  transition: transform 0.2s;
}

.chat-toggle:hover {
  transform: scale(1.05);
}

.chat-window {
  width: 380px;
  height: 500px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-weight: bold;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.message {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}

.message.user {
  flex-direction: row-reverse;
}

.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #f0f2f5;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
}

.content {
  max-width: 75%;
  padding: 10px 14px;
  border-radius: 12px;
  font-size: 14px;
  line-height: 1.5;
}

.message.user .content {
  background: #667eea;
  color: white;
}

.message.assistant .content {
  background: #f0f2f5;
  color: #333;
}

.typing {
  color: #999;
}

.dot {
  animation: bounce 1.4s infinite ease-in-out;
}

.dot:nth-child(1) { animation-delay: 0s; }
.dot:nth-child(2) { animation-delay: 0.2s; }
.dot:nth-child(3) { animation-delay: 0.4s; }

@keyframes bounce {
  0%, 80%, 100% { transform: translateY(0); }
  40% { transform: translateY(-8px); }
}

.chat-input {
  padding: 12px;
  border-top: 1px solid #eee;
}
</style>
