<template>
  <div class="layout" :class="{ 'admin-theme': isAdmin, 'user-theme': !isAdmin }">
    <aside class="sidebar" :class="{ collapsed }">
      <div class="sidebar-header">
        <div class="logo">
          <div class="logo-icon-sm" :class="{ 'admin-logo': isAdmin }">
            <el-icon :size="22"><Key /></el-icon>
          </div>
          <span v-show="!collapsed" class="logo-text">{{ isAdmin ? '管理控制台' : '资源平台' }}</span>
        </div>
        <el-tag v-show="!collapsed" :type="isAdmin ? 'warning' : 'info'" effect="dark" size="small" class="version-tag">
          {{ isAdmin ? 'Admin' : 'v1.0' }}
        </el-tag>
      </div>
      <nav class="sidebar-nav">
        <template v-for="group in menuGroups" :key="group.label">
          <div v-show="!collapsed && group.items.length > 0" class="nav-group-label">{{ group.label }}</div>
          <router-link
            v-for="item in group.items"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: isActive(item.path) }"
            :title="collapsed ? item.title : ''"
          >
            <el-icon :size="20"><component :is="item.icon" /></el-icon>
            <span v-show="!collapsed">{{ item.title }}</span>
            <el-badge v-if="item.badge && !collapsed" :value="item.badge" :max="99" class="nav-badge" />
            <div v-if="isActive(item.path) && !collapsed" class="active-indicator"></div>
          </router-link>
        </template>
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
          <el-tag v-if="isAdmin" type="warning" effect="dark" size="small" class="role-tag">
            <el-icon><UserFilled /></el-icon>
            管理员
          </el-tag>
          <el-tag v-else type="info" effect="plain" size="small" class="role-tag">
            <el-icon><User /></el-icon>
            普通用户
          </el-tag>
        </div>
        <div class="header-right">
          <NotificationBell />
          <el-dropdown trigger="click" @command="handleCommand">
            <div class="user-profile">
              <div class="avatar" :class="{ admin: isAdmin }">
                <el-icon :size="20"><UserFilled /></el-icon>
              </div>
              <div class="user-info">
                <span class="username">{{ userStore.userInfo?.nickname || userStore.userInfo?.username }}</span>
                <span class="user-role">{{ isAdmin ? '系统管理员' : '普通用户' }}</span>
              </div>
              <el-icon :size="14"><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile" disabled>
                  <el-icon><User /></el-icon>
                  {{ userStore.userInfo?.username }}
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
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
import { Key, DataBoard, User, UserFilled, Fold, Expand, ArrowDown, SwitchButton, Document, Checked, List, Bell, Cpu } from '@element-plus/icons-vue'
import { useUserStore } from '@/store/user'
import { ElMessage } from 'element-plus'
import { wsService } from '@/utils/websocket'
import NotificationBell from '@/components/NotificationBell.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const collapsed = ref(false)

const isAdmin = computed(() => userStore.hasRole('admin'))

interface MenuItem {
  path: string
  title: string
  icon: any
  adminOnly: boolean
  userOnly?: boolean
  badge?: () => number
}

interface MenuGroup {
  label: string
  items: MenuItem[]
}

const menuGroups = computed<MenuGroup[]>(() => {
  const groups: MenuGroup[] = [
    {
      label: '工作台',
      items: [
        { path: '/dashboard', title: '仪表盘', icon: DataBoard, adminOnly: false }
      ]
    },
    {
      label: '资源中心',
      items: [
        { path: '/resources', title: isAdmin.value ? '资源管理' : '资源清单', icon: Cpu, adminOnly: false },
        { path: '/apply', title: '申请资源', icon: Document, adminOnly: false, userOnly: true },
        { path: '/my-applications', title: '我的申请', icon: List, adminOnly: false, userOnly: true }
      ]
    },
    {
      label: '审核管理',
      items: [
        { path: '/audit', title: '资源审核', icon: Checked, adminOnly: true }
      ]
    },
    {
      label: '消息中心',
      items: [
        { path: '/messages', title: '消息通知', icon: Bell, adminOnly: false }
      ]
    },
    {
      label: '系统管理',
      items: [
        { path: '/users', title: '用户管理', icon: User, adminOnly: true },
        { path: '/roles', title: '角色管理', icon: UserFilled, adminOnly: true },
        { path: '/permissions', title: '权限管理', icon: Key, adminOnly: true }
      ]
    }
  ]

  return groups.map(group => ({
    label: group.label,
    items: group.items.filter(item => {
      if (item.adminOnly && !isAdmin.value) return false
      if (item.userOnly && isAdmin.value) return false
      return true
    })
  })).filter(group => group.items.length > 0)
})

const currentTitle = computed(() => {
  if (route.path === '/resources') {
    return isAdmin.value ? '资源管理' : '资源清单'
  }
  const matched = route.matched.find(r => r.meta.title)
  return matched?.meta?.title || ''
})

const isActive = (path: string) => {
  if (path === '/dashboard') return route.path === '/' || route.path === '/dashboard'
  return route.path.startsWith(path)
}

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

.layout.admin-theme .sidebar {
  background: linear-gradient(180deg, rgba(30, 20, 10, 0.98), rgba(20, 15, 10, 0.95));
  border-right: 1px solid rgba(245, 158, 11, 0.25);
}

.layout.admin-theme .nav-item:hover {
  background: rgba(245, 158, 11, 0.08);
  color: #fbbf24;
}

.layout.admin-theme .nav-item.active {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.25), rgba(249, 115, 22, 0.12));
  color: #fbbf24;
}

.layout.admin-theme .active-indicator {
  background: linear-gradient(180deg, #f59e0b, #f97316);
  box-shadow: 0 0 12px rgba(245, 158, 11, 0.5);
}

.layout.admin-theme .logo-icon-sm.admin-logo {
  background: linear-gradient(135deg, #f59e0b, #f97316);
  box-shadow: 0 4px 15px rgba(245, 158, 11, 0.3);
}

.layout.admin-theme .logo-text {
  background: linear-gradient(135deg, #fbbf24, #f59e0b);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.layout.admin-theme .nav-group-label {
  color: rgba(245, 158, 11, 0.7);
}

.layout.admin-theme .version-tag {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.2), rgba(249, 115, 22, 0.1)) !important;
  border: 1px solid rgba(245, 158, 11, 0.3);
  color: #fbbf24 !important;
}

.layout.user-theme .sidebar {
  background: linear-gradient(180deg, rgba(15, 20, 40, 0.98), rgba(10, 15, 35, 0.95));
  border-right: 1px solid rgba(99, 102, 241, 0.25);
}

.layout.user-theme .nav-item:hover {
  background: rgba(99, 102, 241, 0.08);
  color: #818cf8;
}

.layout.user-theme .logo-icon-sm:not(.admin-logo) {
  background: linear-gradient(135deg, #6366f1, #06b6d4);
  box-shadow: 0 4px 15px rgba(99, 102, 241, 0.3);
}

.layout.user-theme .nav-group-label {
  color: rgba(99, 102, 241, 0.7);
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
  margin-bottom: 12px;
}

.version-tag {
  margin-left: 52px;
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
  padding: 12px 12px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  overflow-y: auto;
  overflow-x: hidden;
}

.nav-group-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 1px;
  padding: 16px 16px 8px;
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
  flex: 1;
}

.nav-badge {
  flex-shrink: 0;
}

.nav-badge :deep(.el-badge__content) {
  background: var(--danger);
  border: none;
  font-size: 10px;
  height: 16px;
  line-height: 16px;
  padding: 0 4px;
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

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
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

.role-tag {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
}

.user-profile {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 14px 6px 6px;
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
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, var(--primary-500), var(--accent-purple));
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

.avatar.admin {
  background: linear-gradient(135deg, var(--warning), #f97316);
}

.user-info {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.username {
  font-size: 14px;
  font-weight: 500;
  line-height: 1.2;
}

.user-role {
  font-size: 11px;
  color: var(--text-muted);
  line-height: 1.2;
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
</style>

<style>
/* el-dropdown-menu 使用 Teleport 渲染到 body，必须用全局样式 */
.el-dropdown-menu {
  background: rgba(20, 20, 40, 0.92) !important;
  backdrop-filter: blur(24px) saturate(150%) !important;
  -webkit-backdrop-filter: blur(24px) saturate(150%) !important;
  border: 1px solid rgba(255, 255, 255, 0.1) !important;
  border-radius: 12px !important;
  padding: 6px !important;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.5) !important;
  min-width: 180px !important;
}

.el-dropdown-menu__item {
  color: rgba(255, 255, 255, 0.85) !important;
  border-radius: 8px !important;
  margin: 2px 0 !important;
  padding: 10px 14px !important;
  font-size: 14px !important;
  transition: all 0.15s ease !important;
}

.el-dropdown-menu__item:hover {
  background: rgba(99, 102, 241, 0.15) !important;
  color: #fff !important;
}

.el-dropdown-menu__item.is-disabled {
  color: rgba(255, 255, 255, 0.4) !important;
  cursor: default !important;
}

.el-dropdown-menu__item.is-disabled:hover {
  background: transparent !important;
}

.el-dropdown-menu__item--divided {
  border-top: 1px solid rgba(255, 255, 255, 0.08) !important;
  margin-top: 6px !important;
  padding-top: 10px !important;
}

.el-dropdown-menu__item .el-icon {
  margin-right: 8px !important;
  vertical-align: middle;
}
</style>
