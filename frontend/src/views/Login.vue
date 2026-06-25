<template>
  <div class="login-page">
    <div class="login-wrapper">
      <div class="login-left">
        <div class="brand">
          <div class="logo-icon">
            <el-icon :size="40"><Key /></el-icon>
          </div>
          <h1>Casbin RBAC</h1>
          <p class="subtitle">企业级权限管理系统</p>
        </div>
        <div class="features">
          <div class="feature-item" v-for="(f, i) in features" :key="i" :style="{ animationDelay: `${i * 0.1}s` }">
            <div class="feature-icon">
              <el-icon :size="20"><component :is="f.icon" /></el-icon>
            </div>
            <div class="feature-text">
              <h4>{{ f.title }}</h4>
              <p>{{ f.desc }}</p>
            </div>
          </div>
        </div>
      </div>
      <div class="login-right">
        <div class="form-container animate-fade-in-up">
          <h2>{{ isRegister ? '创建账户' : '欢迎回来' }}</h2>
          <p class="form-subtitle">{{ isRegister ? '填写信息注册新账号' : '登录您的账号继续' }}</p>
          
          <el-form ref="formRef" :model="form" :rules="rules" @submit.prevent="handleSubmit">
            <el-form-item prop="username">
              <el-input
                v-model="form.username"
                placeholder="用户名"
                size="large"
                class="glass-input"
                :prefix-icon="User"
              />
            </el-form-item>
            <el-form-item v-if="isRegister" prop="nickname">
              <el-input
                v-model="form.nickname"
                placeholder="昵称（选填）"
                size="large"
                class="glass-input"
                :prefix-icon="Avatar"
              />
            </el-form-item>
            <el-form-item v-if="isRegister" prop="email">
              <el-input
                v-model="form.email"
                placeholder="邮箱（选填）"
                size="large"
                class="glass-input"
                :prefix-icon="Message"
              />
            </el-form-item>
            <el-form-item prop="password">
              <el-input
                v-model="form.password"
                type="password"
                placeholder="密码"
                size="large"
                class="glass-input"
                :prefix-icon="Lock"
                show-password
                @keyup.enter="handleSubmit"
              />
            </el-form-item>
            
            <button type="submit" class="btn-gradient w-full" :disabled="loading">
              <span v-if="loading">
                <el-icon class="is-loading"><Loading /></el-icon>
                {{ isRegister ? '注册中...' : '登录中...' }}
              </span>
              <span v-else>{{ isRegister ? '注 册' : '登 录' }}</span>
            </button>
          </el-form>
          
          <div class="switch-mode">
            <span>{{ isRegister ? '已有账号？' : '还没有账号？' }}</span>
            <a @click="toggleMode">{{ isRegister ? '立即登录' : '立即注册' }}</a>
          </div>
          
          <div class="demo-accounts">
            <p class="demo-title">演示账号</p>
            <div class="demo-btns">
              <button class="demo-btn" @click="fillDemo('admin')">管理员 admin</button>
              <button class="demo-btn" @click="fillDemo('user')">普通用户 user</button>
            </div>
            <p class="demo-pwd">密码均为：123456</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, Key, DataBoard, Setting, Avatar, Message, Loading } from '@element-plus/icons-vue'
import { useUserStore } from '@/store/user'

const router = useRouter()
const userStore = useUserStore()

const isRegister = ref(false)
const loading = ref(false)
const formRef = ref()

const form = reactive({
  username: '',
  password: '',
  nickname: '',
  email: ''
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const features = [
  { icon: Lock, title: 'Casbin 授权', desc: '基于RBAC模型的精细权限控制' },
  { icon: DataBoard, title: '仪表盘', desc: '实时数据监控与统计概览' },
  { icon: Setting, title: '灵活配置', desc: '用户、角色、权限全生命周期管理' }
]

const toggleMode = () => {
  isRegister.value = !isRegister.value
  form.username = ''
  form.password = ''
  form.nickname = ''
  form.email = ''
}

const fillDemo = (role: string) => {
  form.username = role
  form.password = '123456'
}

const handleSubmit = async () => {
  await formRef.value?.validate()
  loading.value = true
  try {
    if (isRegister.value) {
      const { register } = await import('@/api/auth')
      await register({
        username: form.username,
        password: form.password,
        nickname: form.nickname || undefined,
        email: form.email || undefined
      })
      ElMessage.success('注册成功，请登录')
      isRegister.value = false
      form.password = ''
    } else {
      await userStore.login({
        username: form.username,
        password: form.password
      })
      ElMessage.success('登录成功')
      router.push('/dashboard')
    }
  } catch (e) {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.login-wrapper {
  display: flex;
  width: 100%;
  max-width: 1000px;
  min-height: 600px;
  border-radius: 24px;
  overflow: hidden;
  position: relative;
  z-index: 1;
}

.login-left {
  flex: 1;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.15), rgba(6, 182, 212, 0.1));
  backdrop-filter: blur(20px);
  border: 1px solid var(--border-glass);
  border-right: none;
  border-radius: 24px 0 0 24px;
  padding: 48px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.login-right {
  flex: 1;
  background: var(--bg-glass);
  backdrop-filter: blur(20px);
  border: 1px solid var(--border-glass);
  border-radius: 0 24px 24px 0;
  padding: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.brand {
  margin-bottom: 40px;
}

.logo-icon {
  width: 64px;
  height: 64px;
  background: linear-gradient(135deg, var(--primary-500), var(--accent-cyan));
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  margin-bottom: 20px;
  animation: pulse-glow 3s ease-in-out infinite;
}

.brand h1 {
  font-size: 32px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--text-primary), var(--primary-300));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.subtitle {
  color: var(--text-secondary);
  margin-top: 8px;
  font-size: 16px;
}

.feature-item {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 24px;
  opacity: 0;
  animation: slideInLeft 0.5s ease forwards;
}

.feature-icon {
  width: 44px;
  height: 44px;
  background: var(--bg-glass);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--primary-400);
  flex-shrink: 0;
}

.feature-text h4 {
  color: var(--text-primary);
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 4px;
}

.feature-text p {
  color: var(--text-muted);
  font-size: 13px;
}

.form-container {
  width: 100%;
  max-width: 360px;
}

.form-container h2 {
  font-size: 28px;
  font-weight: 700;
  margin-bottom: 8px;
}

.form-subtitle {
  color: var(--text-secondary);
  margin-bottom: 32px;
  font-size: 14px;
}

.w-full {
  width: 100%;
}

.switch-mode {
  text-align: center;
  margin-top: 24px;
  color: var(--text-secondary);
  font-size: 14px;
}

.switch-mode a {
  color: var(--primary-400);
  cursor: pointer;
  margin-left: 4px;
  font-weight: 500;
  transition: color 0.2s;
}

.switch-mode a:hover {
  color: var(--primary-300);
}

.demo-accounts {
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid var(--border-glass);
}

.demo-title {
  text-align: center;
  color: var(--text-muted);
  font-size: 12px;
  margin-bottom: 12px;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.demo-btns {
  display: flex;
  gap: 8px;
  justify-content: center;
}

.demo-btn {
  padding: 8px 16px;
  background: var(--bg-glass);
  border: 1px solid var(--border-glass);
  border-radius: 10px;
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.demo-btn:hover {
  background: var(--bg-glass-hover);
  border-color: var(--primary-500);
  color: var(--primary-400);
}

.demo-pwd {
  text-align: center;
  color: var(--text-muted);
  font-size: 12px;
  margin-top: 8px;
}

@media (max-width: 768px) {
  .login-wrapper {
    flex-direction: column;
    max-width: 400px;
  }
  .login-left {
    display: none;
  }
  .login-right {
    border-radius: 24px;
    padding: 32px 24px;
  }
}
</style>
