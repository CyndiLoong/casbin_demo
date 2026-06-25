<template>
  <div class="layout">
    <aside class="sidebar" :class="{ collapsed }">
      <div class="sidebar-header">
        <div class="logo">
          <div class="logo-icon-sm">
            <el-icon :size="22"><Key /></el-icon>
          </div>
          <span v-show="!collapsed" class="logo-text">Casbin RBAC</span>
        </div>
      </div>
      <nav class="sidebar-nav">
        <router-link
          v-for="item in menuItems"
          :key="item.path"
          :to="item.path"
          class="nav-item"
          :class="{ active: isActive(item.path) }"
        >
          <el-icon :size="20"><component :is="item.icon" /></el-icon>
          <span v-show="!collapsed">{{ item.title }}</span>
          <div v-if="isActive(item.path) && !collapsed" class="active-indicator"></div>
        </router-link>
      </nav>
      <div class="sidebar-footer">
        <button class="collapse-btn" @click="collapsed = !collapsed">
          <el-icon :size="18">
            <Fold v-if="!collapsed" />
            <Expand v-else />
          </el-icon>
        </button>
      </div>
    </aside>

    <main class="main-content" :class="{ expanded: collapsed }">
      <header class="header">
        <div class="header-left">
          <h2 class="page-title">{{ currentTitle }}</h2>
        </div>
        <div class="header-right">
          <NotificationBell />
          <el-dropdown trigger="click" @command="handleCommand">
            <div class="user-profile">
              <div class="avatar">
                <el-icon :size="20"><UserFilled /></el-icon>
              </div>
              <span class="username">{{ userStore.userInfo?.nickname || userStore.userInfo?.username }}</span>
              <el-icon :size="14"><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <div class="page-content">
        <router-view v-slot="{ Component }">
          <transition name="page" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Key, DataBoard, User, UserFilled, Fold, Expand, ArrowDown, SwitchButton, Document, Checked, List } from '@element-plus/icons-vue'
import { useUserStore } from '@/store/user'
import { ElMessage } from 'element-plus'
import { wsService } from '@/utils/websocket'
import NotificationBell from '@/components/NotificationBell.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const collapsed = ref(false)

const isAdmin = computed(() => userStore.hasRole('admin'))

const allMenuItems = [
  { path: '/dashboard', title: '仪表盘', icon: DataBoard, adminOnly: false },
  { path: '/apply', title: '资源申请', icon: Document, adminOnly: false },
  { path: '/my-applications', title: '我的申请', icon: List, adminOnly: false },
  { path: '/audit', title: '审核管理', icon: Checked, adminOnly: true },
  { path: '/users', title: '用户管理', icon: User, adminOnly: true },
  { path: '/roles', title: '角色管理', icon: UserFilled, adminOnly: true },
  { path: '/permissions', title: '权限管理', icon: Key, adminOnly: true }
]

const menuItems = computed(() => {
  return allMenuItems.filter(item => !item.adminOnly || isAdmin.value)
})

const currentTitle = computed(() => {
  const matched = route.matched.find(r => r.meta.title)
  return matched?.meta?.title || ''
})

const isActive = (path: string) => route.path.startsWith(path)

const handleCommand = (cmd: string) => {
  if (cmd === 'logout') {
    wsService.disconnect()
    userStore.logout()
    ElMessage.success('已退出登录')
    router.push('/login')
  }
}

onMounted(async () => {
  if (!userStore.userInfo) {
    await userStore.fetchUserInfo()
  }
  if (userStore.token) {
    wsService.connect(userStore.token)
  }
})

onUnmounted(() => {
  wsService.disconnect()
})

watch(() => userStore.token, (newToken) => {
  if (newToken) {
    wsService.connect(newToken)
  } else {
    wsService.disconnect()
  }
})
</script>

<style scoped>
.layout {
  display: flex;
  min-height: 100vh;
}

.sidebar {
  width: var(--sidebar-width);
  background: rgba(15, 15, 35, 0.9);
  backdrop-filter: blur(20px);
  border-right: 1px solid var(--border-glass);
  display: flex;
  flex-direction: column;
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  z-index: 100;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.sidebar.collapsed {
  width: 72px;
}

.sidebar-header {
  padding: 20px;
  border-bottom: 1px solid var(--border-glass);
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-icon-sm {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, var(--primary-500), var(--accent-cyan));
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

.logo-text {
  font-size: 18px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--text-primary), var(--primary-300));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  white-space: nowrap;
}

.sidebar-nav {
  flex: 1;
  padding: 16px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 12px;
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 14px;
  font-weight: 500;
  position: relative;
  transition: all 0.2s ease;
}

.nav-item:hover {
  background: var(--bg-glass);
  color: var(--text-primary);
}

.nav-item.active {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.2), rgba(6, 182, 212, 0.1));
  color: var(--primary-300);
}

.nav-item span {
  white-space: nowrap;
}

.active-indicator {
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 24px;
  background: linear-gradient(180deg, var(--primary-500), var(--accent-cyan));
  border-radius: 2px;
}

.sidebar-footer {
  padding: 16px;
  border-top: 1px solid var(--border-glass);
}

.collapse-btn {
  width: 100%;
  padding: 10px;
  background: var(--bg-glass);
  border: 1px solid var(--border-glass);
  border-radius: 10px;
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.collapse-btn:hover {
  background: var(--bg-glass-hover);
  color: var(--text-primary);
}

.main-content {
  flex: 1;
  margin-left: var(--sidebar-width);
  display: flex;
  flex-direction: column;
  transition: margin-left 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.main-content.expanded {
  margin-left: 72px;
}

.header {
  height: var(--header-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 32px;
  background: rgba(15, 15, 35, 0.5);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid var(--border-glass);
  position: sticky;
  top: 0;
  z-index: 50;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
}

.user-profile {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 12px 6px 6px;
  background: var(--bg-glass);
  border: 1px solid var(--border-glass);
  border-radius: 30px;
  cursor: pointer;
  transition: all 0.2s;
}

.user-profile:hover {
  background: var(--bg-glass-hover);
}

.avatar {
  width: 32px;
  height: 32px;
  background: linear-gradient(135deg, var(--primary-500), var(--accent-purple));
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.username {
  font-size: 14px;
  font-weight: 500;
}

.page-content {
  flex: 1;
  padding: 32px;
}

.page-enter-active,
.page-leave-active {
  transition: all 0.3s ease;
}

.page-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.page-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

:deep(.el-dropdown-menu) {
  background: rgba(20, 20, 40, 0.95) !important;
  backdrop-filter: blur(20px);
  border: 1px solid var(--border-glass) !important;
  border-radius: 12px !important;
}

:deep(.el-dropdown-menu__item) {
  color: var(--text-primary) !important;
  border-radius: 8px;
}

:deep(.el-dropdown-menu__item:hover) {
  background: var(--bg-glass-hover) !important;
}
</style>
