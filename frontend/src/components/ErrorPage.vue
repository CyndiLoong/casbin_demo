<template>
  <div class="error-page">
    <div class="bg-orb" :style="{ background: errorConfig.color }"></div>
    <div class="bg-orb bg-orb-secondary" :style="{ background: errorConfig.secondaryColor }"></div>

    <div class="error-container">
      <div class="error-code-wrapper">
        <div class="error-code-glow" :style="{ background: `radial-gradient(circle, ${errorConfig.glowColor}, transparent 70%)` }"></div>
        <span class="error-code" :style="{ color: errorConfig.color }">{{ code }}</span>
      </div>

      <div class="error-icon" :style="{ color: errorConfig.color }">
        <component :is="errorConfig.icon" :size="64" />
      </div>

      <h1 class="error-title">{{ errorConfig.title }}</h1>
      <p class="error-desc">{{ errorConfig.description }}</p>

      <div class="error-actions">
        <template v-if="code === 401">
          <button class="btn-gradient btn-home" @click="goLogin">
            <el-icon><ArrowRight /></el-icon>
            去登录
          </button>
          <button class="btn-glass" @click="goBack">
            <el-icon><Back /></el-icon>
            返回上页
          </button>
        </template>
        <template v-else>
          <button class="btn-gradient btn-home" @click="goHome">
            <el-icon><House /></el-icon>
            返回首页
          </button>
          <button class="btn-glass" @click="goBack">
            <el-icon><Back /></el-icon>
            返回上页
          </button>
        </template>
      </div>

      <div class="error-footnote">
        <span v-if="code === 401">请登录后再试</span>
        <span v-else-if="code === 403">如需访问权限，请联系管理员</span>
        <span v-else-if="code === 404">页面可能已被移动或删除</span>
        <span v-else>请稍后重试，或联系技术支持</span>
      </div>
    </div>

    <div class="grid-lines"></div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Lock, Warning, QuestionFilled, CircleClose, House, Back, ArrowRight } from '@element-plus/icons-vue'

const props = defineProps<{
  code: 401 | 403 | 404 | 500
}>()

const router = useRouter()

interface ErrorConfig {
  title: string
  description: string
  color: string
  secondaryColor: string
  glowColor: string
  icon: any
}

const errorConfigs: Record<number, ErrorConfig> = {
  401: {
    title: '身份未认证',
    description: '您需要提供有效的身份凭证才能访问此资源',
    color: '#f59e0b',
    secondaryColor: '#fbbf24',
    glowColor: 'rgba(245, 158, 11, 0.15)',
    icon: Lock
  },
  403: {
    title: '访问被拒绝',
    description: '您的账户没有权限访问此页面或资源',
    color: '#f97316',
    secondaryColor: '#fb923c',
    glowColor: 'rgba(249, 115, 22, 0.15)',
    icon: Warning
  },
  404: {
    title: '页面未找到',
    description: '您访问的页面不存在，或已被移至其他位置',
    color: '#06b6d4',
    secondaryColor: '#22d3ee',
    glowColor: 'rgba(6, 182, 212, 0.15)',
    icon: QuestionFilled
  },
  500: {
    title: '服务器内部错误',
    description: '服务器遇到了意外情况，无法完成您的请求',
    color: '#ef4444',
    secondaryColor: '#f87171',
    glowColor: 'rgba(239, 68, 68, 0.15)',
    icon: CircleClose
  }
}

const errorConfig = computed(() => errorConfigs[props.code])

const goHome = () => {
  router.push('/')
}

const goLogin = () => {
  router.push('/login')
}

const goBack = () => {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/')
  }
}
</script>

<style scoped lang="scss">
.error-page {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-dark);
  overflow: hidden;
  z-index: 9999;
}

.bg-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(100px);
  opacity: 0.35;
  pointer-events: none;
  width: 500px;
  height: 500px;
  top: -150px;
  right: -100px;
  animation: float 10s ease-in-out infinite;
}

.bg-orb-secondary {
  width: 350px;
  height: 350px;
  bottom: -100px;
  left: -80px;
  top: auto;
  right: auto;
  animation-delay: -5s;
}

.grid-lines {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.02) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.02) 1px, transparent 1px);
  background-size: 50px 50px;
  mask-image: radial-gradient(ellipse at center, black 20%, transparent 70%);
  -webkit-mask-image: radial-gradient(ellipse at center, black 20%, transparent 70%);
  pointer-events: none;
}

.error-container {
  position: relative;
  z-index: 1;
  text-align: center;
  padding: 60px 80px;
  background: var(--bg-glass);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border: 1px solid var(--border-glass);
  border-radius: 24px;
  max-width: 560px;
  animation: errorFadeIn 0.8s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

.error-code-wrapper {
  position: relative;
  margin-bottom: 8px;
  height: 160px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.error-code-glow {
  position: absolute;
  width: 280px;
  height: 280px;
  border-radius: 50%;
  animation: pulse-glow 3s ease-in-out infinite;
}

.error-code {
  position: relative;
  font-size: 140px;
  font-weight: 800;
  line-height: 1;
  letter-spacing: -8px;
  font-family: 'Inter', -apple-system, sans-serif;
  background: linear-gradient(135deg, currentColor 0%, rgba(255,255,255,0.8) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  text-shadow: none;
  animation: codeReveal 1s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

.error-icon {
  margin: 8px 0 20px;
  animation: iconFloat 3s ease-in-out infinite;
}

.error-title {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 12px;
  letter-spacing: -0.5px;
}

.error-desc {
  font-size: 15px;
  color: var(--text-secondary);
  margin: 0 0 32px;
  line-height: 1.6;
}

.error-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  margin-bottom: 20px;
}

.btn-home {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 28px;
  font-size: 14px;
}

.btn-glass {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 28px;
  background: var(--bg-glass);
  border: 1px solid var(--border-glass);
  border-radius: 12px;
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.btn-glass:hover {
  background: var(--bg-glass-hover);
  border-color: rgba(255, 255, 255, 0.2);
  transform: translateY(-2px);
}

.error-footnote {
  font-size: 13px;
  color: var(--text-muted);
  padding-top: 16px;
  border-top: 1px solid var(--border-glass);
}

@keyframes errorFadeIn {
  from {
    opacity: 0;
    transform: translateY(30px) scale(0.96);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes codeReveal {
  from {
    opacity: 0;
    transform: scale(0.5) translateY(20px);
    filter: blur(10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
    filter: blur(0);
  }
}

@keyframes iconFloat {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-8px); }
}

@keyframes pulse-glow {
  0%, 100% { opacity: 0.5; transform: scale(1); }
  50% { opacity: 1; transform: scale(1.1); }
}

@keyframes float {
  0%, 100% { transform: translate(0, 0) scale(1); }
  33% { transform: translate(30px, -40px) scale(1.08); }
  66% { transform: translate(-20px, 20px) scale(0.95); }
}

@media (max-width: 640px) {
  .error-container {
    padding: 40px 24px;
    margin: 16px;
    border-radius: 20px;
  }

  .error-code {
    font-size: 100px;
    letter-spacing: -4px;
  }

  .error-code-wrapper {
    height: 120px;
  }

  .error-code-glow {
    width: 200px;
    height: 200px;
  }

  .error-title {
    font-size: 22px;
  }

  .error-actions {
    flex-direction: column;
  }

  .btn-home, .btn-glass {
    justify-content: center;
    width: 100%;
  }
}
</style>
