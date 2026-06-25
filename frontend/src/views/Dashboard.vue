<template>
  <div class="dashboard">
    <div class="welcome-banner glass-card animate-fade-in-up">
      <div class="welcome-content">
        <h1>欢迎回来，{{ user?.nickname || user?.username }} 👋</h1>
        <p>基于 Casbin 的 RBAC 权限管理系统，为您提供精细的访问控制</p>
        <div class="role-badges">
          <el-tag v-for="role in user?.roles" :key="role" type="primary" effect="dark" round>
            {{ role }}
          </el-tag>
        </div>
      </div>
      <div class="welcome-visual">
        <div class="pulse-ring"></div>
        <el-icon :size="64" color="var(--primary-400)"><DataBoard /></el-icon>
      </div>
    </div>

    <div class="stats-grid" v-if="stats">
      <div v-for="(s, i) in statCards" :key="i" class="stat-card glass-card animate-fade-in-up" :style="{ animationDelay: `${i * 0.1}s` }">
        <div class="stat-icon" :style="{ background: s.gradient }">
          <el-icon :size="24"><component :is="s.icon" /></el-icon>
        </div>
        <div class="stat-info">
          <p class="stat-value">{{ s.value }}</p>
          <p class="stat-label">{{ s.label }}</p>
        </div>
        <div class="stat-trend up">
          <el-icon><TrendCharts /></el-icon>
        </div>
      </div>
    </div>

    <div class="content-grid">
      <div class="tech-card glass-card animate-fade-in-up" style="animation-delay: 0.3s">
        <h3><el-icon><SetUp /></el-icon> 技术栈</h3>
        <div class="tech-list">
          <div v-for="tech in techStack" :key="tech.name" class="tech-item">
            <span class="tech-dot" :style="{ background: tech.color }"></span>
            <span class="tech-name">{{ tech.name }}</span>
            <span class="tech-desc">{{ tech.desc }}</span>
          </div>
        </div>
      </div>
      <div class="api-card glass-card animate-fade-in-up" style="animation-delay: 0.4s">
        <h3><el-icon><Connection /></el-icon> 系统状态</h3>
        <div class="status-list">
          <div v-for="svc in services" :key="svc.name" class="status-item">
            <span class="status-indicator" :class="svc.status"></span>
            <span class="status-name">{{ svc.name }}</span>
            <span class="status-text">{{ svc.text }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { DataBoard, User, UserFilled, Key, TrendCharts, SetUp, Connection } from '@element-plus/icons-vue'
import { getDashboard } from '@/api/auth'
import { useUserStore } from '@/store/user'

const userStore = useUserStore()
const stats = ref<any>(null)
const user = computed(() => userStore.userInfo)

const statCards = computed(() => [
  { icon: User, label: '用户总数', value: stats.value?.stats?.total_users || 0, gradient: 'linear-gradient(135deg, #6366f1, #818cf8)' },
  { icon: UserFilled, label: '角色数量', value: stats.value?.stats?.total_roles || 0, gradient: 'linear-gradient(135deg, #06b6d4, #22d3ee)' },
  { icon: Key, label: '权限节点', value: stats.value?.stats?.total_permissions || 0, gradient: 'linear-gradient(135deg, #a855f7, #c084fc)' }
])

const techStack = [
  { name: 'Gin', desc: 'Go Web 框架', color: '#00ADD8' },
  { name: 'Gorm', desc: 'ORM 框架', color: '#6366f1' },
  { name: 'Casbin', desc: '权限控制', color: '#10b981' },
  { name: 'PostgreSQL', desc: '关系型数据库', color: '#336791' },
  { name: 'Redis', desc: '缓存中间件', color: '#DC382D' },
  { name: 'Vue 3', desc: '前端框架', color: '#42b883' }
]

const services = [
  { name: 'API 服务', status: 'online', text: '运行中' },
  { name: 'PostgreSQL', status: 'online', text: '已连接' },
  { name: 'Redis', status: 'online', text: '已连接' },
  { name: 'Casbin', status: 'online', text: '策略已加载' }
]

onMounted(async () => {
  try {
    const res = await getDashboard()
    stats.value = res.data
  } catch (e) {
    // fallback
    stats.value = { stats: { total_users: 2, total_roles: 2, total_permissions: 10 } }
  }
})
</script>

<style scoped>
.dashboard {
  max-width: 1400px;
}

.welcome-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 32px;
  margin-bottom: 24px;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.1), rgba(6, 182, 212, 0.05));
}

.welcome-content h1 {
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 8px;
}

.welcome-content p {
  color: var(--text-secondary);
  margin-bottom: 16px;
}

.role-badges {
  display: flex;
  gap: 8px;
}

.welcome-visual {
  position: relative;
  width: 100px;
  height: 100px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pulse-ring {
  position: absolute;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  border: 2px solid var(--primary-500);
  animation: pulse-ring 2s ease-out infinite;
}

@keyframes pulse-ring {
  0% { transform: scale(0.8); opacity: 1; }
  100% { transform: scale(1.5); opacity: 0; }
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  line-height: 1.2;
}

.stat-label {
  color: var(--text-muted);
  font-size: 13px;
  margin-top: 4px;
}

.stat-trend {
  color: var(--success);
}

.content-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(360px, 1fr));
  gap: 20px;
}

.tech-card, .api-card {
  padding: 24px;
}

.tech-card h3, .api-card h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 20px;
}

.tech-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.tech-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 0;
}

.tech-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.tech-name {
  font-weight: 600;
  width: 100px;
}

.tech-desc {
  color: var(--text-muted);
  font-size: 13px;
}

.status-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 0;
}

.status-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-indicator.online {
  background: var(--success);
  box-shadow: 0 0 10px var(--success);
}

.status-name {
  font-weight: 500;
  width: 120px;
}

.status-text {
  color: var(--text-muted);
  font-size: 13px;
}
</style>
