<template>
  <div class="notification-wrapper">
    <el-badge :value="totalCount" :hidden="totalCount === 0" :max="99" class="notification-badge">
      <button class="bell-btn" @click="showPanel = !showPanel">
        <el-icon :size="20"><Bell /></el-icon>
      </button>
    </el-badge>

    <transition name="dropdown">
      <div v-if="showPanel" class="notification-panel glass-card">
        <div class="panel-header">
          <span class="panel-title">消息通知</span>
          <button v-if="unreadCount > 0" class="mark-all-btn" @click="handleMarkAllRead">全部已读</button>
        </div>
        <div class="panel-tabs">
          <button :class="['tab-btn', { active: activeTab === 'notifications' }]" @click="activeTab = 'notifications'">
            <el-icon><Bell /></el-icon>
            通知
            <span v-if="unreadCount > 0" class="tab-badge">{{ unreadCount }}</span>
          </button>
          <button v-if="isAdmin" :class="['tab-btn', { active: activeTab === 'pending' }]" @click="activeTab = 'pending'">
            <el-icon><Warning /></el-icon>
            待审核
            <span v-if="pendingCount > 0" class="tab-badge pending">{{ pendingCount }}</span>
          </button>
        </div>
        <div class="panel-body" ref="panelBody">
          <div v-if="loading" class="panel-empty">
            <el-icon class="is-loading"><Loading /></el-icon>
            <span>加载中...</span>
          </div>
          <template v-else-if="activeTab === 'notifications'">
            <div v-if="messages.length === 0" class="panel-empty">
              <el-icon :size="32" color="var(--text-muted)"><Bell /></el-icon>
              <span>暂无消息</span>
            </div>
            <div v-else class="message-list">
              <div
                v-for="msg in messages"
                :key="msg.id"
                :class="['message-item', { unread: !msg.is_read }]"
                @click="handleMessageClick(msg)"
              >
                <div class="message-icon" :class="getMsgTypeClass(msg.type)">
                  <el-icon :size="16"><component :is="getMsgIcon(msg.type)" /></el-icon>
                </div>
                <div class="message-content">
                  <div class="message-title">{{ msg.title }}</div>
                  <div class="message-text">{{ msg.content }}</div>
                  <div class="message-time">{{ formatTime(msg.created_at) }}</div>
                </div>
                <div v-if="!msg.is_read" class="unread-dot"></div>
              </div>
            </div>
          </template>
          <template v-else-if="activeTab === 'pending'">
            <div v-if="pendingApps.length === 0" class="panel-empty">
              <el-icon :size="32" color="var(--success)"><CircleCheck /></el-icon>
              <span>暂无待审核申请</span>
            </div>
            <div v-else class="message-list">
              <div
                v-for="app in pendingApps"
                :key="app.id"
                class="message-item"
                @click="goToAudit(app.id)"
              >
                <div class="message-icon pending">
                  <el-icon :size="16"><Document /></el-icon>
                </div>
                <div class="message-content">
                  <div class="message-title">{{ app.applicant_name }} 提交了API申请</div>
                  <div class="message-text">资源: {{ app.resource_name }}</div>
                  <div class="message-time">{{ formatTime(app.created_at) }}</div>
                </div>
              </div>
            </div>
          </template>
        </div>
        <div v-if="messages.length > 0" class="panel-footer">
          <button class="view-all-btn" @click="goToMessages">查看全部消息</button>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { Bell, Warning, Loading, CircleCheck, Document, BellFilled, Check } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getMessages, getPendingCount, markAllRead, markMessageRead, getAllApplications, type SysMessage, type AuditApplication } from '@/api/audit'
import { wsService } from '@/utils/websocket'
import { useUserStore } from '@/store/user'

const router = useRouter()
const userStore = useUserStore()

const showPanel = ref(false)
const activeTab = ref<'notifications' | 'pending'>('notifications')
const loading = ref(false)
const messages = ref<SysMessage[]>([])
const pendingApps = ref<AuditApplication[]>([])
const unreadCount = ref(0)
const pendingCount = ref(0)
const panelBody = ref<HTMLElement | null>(null)

const isAdmin = computed(() => {
  return userStore.userInfo?.roles?.includes('admin') ?? false
})

const totalCount = computed(() => unreadCount.value + (isAdmin.value ? pendingCount.value : 0))

const fetchMessages = async () => {
  loading.value = true
  try {
    const res = await getMessages({ page: 1, page_size: 10, unread: true })
    messages.value = res.data.list
    unreadCount.value = res.data.total
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const fetchPending = async () => {
  if (!isAdmin.value) return
  try {
    const [pendingRes, appsRes] = await Promise.all([
      getPendingCount(),
      getAllApplications({ page: 1, page_size: 5, status: 0 })
    ])
    pendingCount.value = pendingRes.data.count
    pendingApps.value = appsRes.data.list
  } catch (e) {
    console.error(e)
  }
}

const fetchCounts = async () => {
  wsService.fetchCounts()
}

const handleMarkAllRead = async () => {
  try {
    await markAllRead()
    ElMessage.success('已全部标记为已读')
    unreadCount.value = 0
    messages.value.forEach(m => m.is_read = true)
    fetchMessages()
  } catch (e) {
    console.error(e)
  }
}

const handleMessageClick = async (msg: SysMessage) => {
  if (!msg.is_read) {
    try {
      await markMessageRead(msg.id)
      msg.is_read = true
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    } catch (e) {
      console.error(e)
    }
  }
  if (msg.business_type === 'audit_application') {
    if (msg.type === 'new_application') {
      router.push(`/audit`)
    } else {
      router.push(`/my-applications`)
    }
    showPanel.value = false
  }
}

const goToAudit = (id: number) => {
  router.push(`/audit`)
  showPanel.value = false
}

const goToMessages = () => {
  router.push('/messages')
  showPanel.value = false
}

const getMsgTypeClass = (type: string) => {
  if (type === 'new_application') return 'info'
  if (type === 'review_result') return 'success'
  return 'info'
}

const getMsgIcon = (type: string) => {
  if (type === 'new_application') return BellFilled
  if (type === 'review_result') return Check
  return Bell
}

const formatTime = (t: string) => {
  const d = new Date(t)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  if (diff < 604800000) return `${Math.floor(diff / 86400000)}天前`
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
}

const handleClickOutside = (e: MouseEvent) => {
  const wrapper = document.querySelector('.notification-wrapper')
  if (wrapper && !wrapper.contains(e.target as Node)) {
    showPanel.value = false
  }
}

onMounted(() => {
  fetchMessages()
  fetchPending()
  wsService.setUnreadCountCallback((c) => { unreadCount.value = c })
  wsService.setPendingCountCallback((c) => { pendingCount.value = c; if (showPanel.value) fetchPending() })
  wsService.on('new_application', () => { fetchMessages(); fetchPending() })
  wsService.on('review_result', () => { fetchMessages() })
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.notification-wrapper {
  position: relative;
}

.bell-btn {
  width: 40px;
  height: 40px;
  background: var(--bg-glass);
  border: 1px solid var(--border-glass);
  border-radius: 12px;
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.bell-btn:hover {
  background: var(--bg-glass-hover);
  color: var(--text-primary);
}

.notification-badge :deep(.el-badge__content) {
  background: linear-gradient(135deg, var(--danger), var(--accent-pink));
  border: none;
  font-size: 11px;
  padding: 0 5px;
  height: 18px;
  line-height: 18px;
}

.notification-panel {
  position: absolute;
  top: calc(100% + 12px);
  right: 0;
  width: 380px;
  max-height: 520px;
  z-index: 1000;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-glass);
}

.panel-title {
  font-size: 16px;
  font-weight: 600;
}

.mark-all-btn {
  background: none;
  border: none;
  color: var(--primary-400);
  font-size: 13px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: all 0.2s;
}

.mark-all-btn:hover {
  background: rgba(99, 102, 241, 0.1);
}

.panel-tabs {
  display: flex;
  gap: 4px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-glass);
}

.tab-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 12px;
  background: none;
  border: none;
  border-radius: 8px;
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
}

.tab-btn:hover {
  background: var(--bg-glass);
  color: var(--text-primary);
}

.tab-btn.active {
  background: rgba(99, 102, 241, 0.15);
  color: var(--primary-300);
}

.tab-badge {
  background: var(--danger);
  color: white;
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 10px;
  min-width: 18px;
  text-align: center;
}

.tab-badge.pending {
  background: var(--warning);
}

.panel-body {
  flex: 1;
  overflow-y: auto;
  max-height: 340px;
}

.panel-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px 20px;
  color: var(--text-muted);
  font-size: 14px;
}

.message-list {
  padding: 8px;
}

.message-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
}

.message-item:hover {
  background: var(--bg-glass);
}

.message-item.unread {
  background: rgba(99, 102, 241, 0.05);
}

.message-icon {
  width: 32px;
  height: 32px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: rgba(99, 102, 241, 0.15);
  color: var(--primary-400);
}

.message-icon.success {
  background: rgba(16, 185, 129, 0.15);
  color: var(--success);
}

.message-icon.pending {
  background: rgba(245, 158, 11, 0.15);
  color: var(--warning);
}

.message-content {
  flex: 1;
  min-width: 0;
}

.message-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.message-text {
  font-size: 12px;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 4px;
}

.message-time {
  font-size: 11px;
  color: var(--text-muted);
}

.unread-dot {
  width: 8px;
  height: 8px;
  background: var(--primary-500);
  border-radius: 50%;
  flex-shrink: 0;
  margin-top: 6px;
  box-shadow: 0 0 8px rgba(99, 102, 241, 0.5);
}

.panel-footer {
  padding: 12px 20px;
  border-top: 1px solid var(--border-glass);
}

.view-all-btn {
  width: 100%;
  padding: 8px;
  background: var(--bg-glass);
  border: 1px solid var(--border-glass);
  border-radius: 10px;
  color: var(--text-primary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.view-all-btn:hover {
  background: var(--bg-glass-hover);
  border-color: var(--primary-500);
}
</style>
