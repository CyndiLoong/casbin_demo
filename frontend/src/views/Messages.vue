<template>
  <div class="messages-page animate-fade-in-up">
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title-text">消息中心</h2>
        <p class="page-desc">查看系统通知、审核结果等消息</p>
      </div>
      <div class="header-right">
        <div v-if="unreadTotal > 0" class="unread-summary">
          <span class="unread-badge">{{ unreadTotal }}</span>
          <span>条未读</span>
        </div>
        <button v-if="unreadTotal > 0" class="btn-mark-all" @click="handleMarkAll">
          <el-icon><Check /></el-icon>
          全部已读
        </button>
      </div>
    </div>

    <div class="stats-bar glass-card">
      <div class="stat-item" :class="{ active: filterUnread === undefined }" @click="filterUnread = undefined; handleFilter()">
        <span class="stat-num">{{ total }}</span>
        <span class="stat-label">全部消息</span>
      </div>
      <div class="stat-divider"></div>
      <div class="stat-item" :class="{ active: filterUnread === true }" @click="filterUnread = true; handleFilter()">
        <span class="stat-num unread">{{ unreadTotal }}</span>
        <span class="stat-label">未读消息</span>
      </div>
      <div class="stat-divider"></div>
      <div class="stat-item" :class="{ active: filterUnread === false }" @click="filterUnread = false; handleFilter()">
        <span class="stat-num">{{ total - unreadTotal >= 0 ? total - unreadTotal : 0 }}</span>
        <span class="stat-label">已读消息</span>
      </div>
    </div>

    <div v-if="messages.length === 0 && !loading" class="empty-state glass-card">
      <div class="empty-icon">
        <el-icon :size="56"><Bell /></el-icon>
      </div>
      <p class="empty-title">暂无{{ filterUnread === true ? '未读' : '' }}消息</p>
      <p class="empty-text">{{ filterUnread === true ? '所有消息都已读完' : '新消息将在这里显示' }}</p>
    </div>

    <div v-else class="message-list" v-loading="loading">
      <div
        v-for="(msg, index) in messages"
        :key="msg.id"
        :class="['message-item', { unread: !msg.is_read }]"
      >
        <div class="message-timeline">
          <div class="timeline-dot" :class="getTypeClass(msg.type)">
            <el-icon :size="16"><component :is="getTypeIcon(msg.type)" /></el-icon>
          </div>
          <div v-if="index < messages.length - 1" class="timeline-line"></div>
        </div>
        <div class="message-card glass-card" @click="toggleExpand(msg)">
          <div class="card-main">
            <div class="card-header">
              <div class="header-left">
                <span v-if="!msg.is_read" class="unread-dot"></span>
                <h4 class="card-title">{{ msg.title }}</h4>
                <el-tag :type="getBusinessTagType(msg.business_type)" size="small" effect="dark" class="biz-tag">
                  {{ getBusinessLabel(msg.business_type) }}
                </el-tag>
              </div>
              <div class="header-right">
                <span class="card-time">{{ formatRelativeTime(msg.created_at) }}</span>
                <el-icon class="expand-icon" :class="{ expanded: msg._expanded }"><ArrowDown /></el-icon>
              </div>
            </div>
            <p class="card-content" :class="{ expanded: msg._expanded }">{{ msg.content }}</p>
            <div class="card-actions" v-if="msg.business_type === 'audit_application'">
              <button class="action-link" @click.stop="goToBusiness(msg)">
                查看相关业务
                <el-icon><ArrowRight /></el-icon>
              </button>
            </div>
            <div v-if="msg._expanded" class="card-detail">
              <div class="detail-meta">
                <div class="meta-item">
                  <span class="meta-label">消息类型</span>
                  <span class="meta-value">{{ getTypeName(msg.type) }}</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">发送时间</span>
                  <span class="meta-value">{{ formatTime(msg.created_at) }}</span>
                </div>
                <div class="meta-item">
                  <span class="meta-label">消息状态</span>
                  <span class="meta-value" :class="msg.is_read ? 'read' : 'unread'">{{ msg.is_read ? '已读' : '未读' }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="total > pageSize" class="pagination-wrapper glass-card">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, prev, pager, next"
        @size-change="loadMessages"
        @current-change="loadMessages"
        background
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { Bell, Check, ArrowRight, BellFilled, Warning, CircleCheck, ArrowDown } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getMessages, markMessageRead, markAllRead, type SysMessage } from '@/api/audit'
import { wsService } from '@/utils/websocket'
import { useUserStore } from '@/store/user'

interface MessageWithExpand extends SysMessage {
  _expanded?: boolean
}

const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const messages = ref<MessageWithExpand[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const filterUnread = ref<boolean | undefined>(undefined)
const unreadTotal = ref(0)

const isAdmin = computed(() => userStore.hasRole('admin'))

const loadMessages = async () => {
  loading.value = true
  try {
    const res = await getMessages({
      page: page.value,
      page_size: pageSize.value,
      unread: filterUnread.value
    })
    messages.value = res.data.list.map(m => ({ ...m, _expanded: false }))
    total.value = res.data.total
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const loadUnreadCount = async () => {
  try {
    const res = await getMessages({ page: 1, page_size: 1, unread: true })
    unreadTotal.value = res.data.total
  } catch (e) {
    console.error(e)
  }
}

const handleFilter = () => {
  page.value = 1
  loadMessages()
}

const toggleExpand = async (msg: MessageWithExpand) => {
  if (!msg.is_read) {
    try {
      await markMessageRead(msg.id)
      msg.is_read = true
      unreadTotal.value = Math.max(0, unreadTotal.value - 1)
      if (wsService['onUnreadCountChange']) {
        wsService['onUnreadCountChange'](unreadTotal.value)
      }
    } catch (e) {
      console.error(e)
    }
  }
  msg._expanded = !msg._expanded
}

const handleMarkAll = async () => {
  try {
    await markAllRead()
    ElMessage.success('已全部标记为已读')
    unreadTotal.value = 0
    messages.value.forEach(m => m.is_read = true)
    if (wsService['onUnreadCountChange']) {
      wsService['onUnreadCountChange'](0)
    }
    loadMessages()
  } catch (e) {
    console.error(e)
  }
}

const goToBusiness = (msg: MessageWithExpand) => {
  if (!msg.is_read) {
    markMessageRead(msg.id).then(() => {
      msg.is_read = true
      unreadTotal.value = Math.max(0, unreadTotal.value - 1)
    }).catch(() => {})
  }
  msg._expanded = false
  if (msg.business_type === 'audit_application') {
    if (isAdmin.value && (msg.type === 'new_application' || msg.type === 'application_withdrawn')) {
      router.push('/audit')
    } else {
      router.push('/my-applications')
    }
  }
}

const getTypeClass = (type: string) => {
  if (type === 'new_application') return 'info'
  if (type === 'review_result') return 'success'
  if (type === 'application_withdrawn') return 'warning'
  return 'info'
}

const getTypeIcon = (type: string) => {
  if (type === 'new_application') return BellFilled
  if (type === 'review_result') return CircleCheck
  if (type === 'application_withdrawn') return Warning
  return Bell
}

const getTypeName = (type: string) => {
  const map: Record<string, string> = {
    new_application: '新申请通知',
    review_result: '审核结果通知',
    application_withdrawn: '撤回通知'
  }
  return map[type] || '系统通知'
}

const getBusinessLabel = (type: string) => {
  const map: Record<string, string> = {
    audit_application: '资源审核'
  }
  return map[type] || '系统'
}

const getBusinessTagType = (type: string) => {
  if (type === 'audit_application') return 'primary'
  return 'info'
}

const formatTime = (t: string) => {
  const d = new Date(t)
  return `${d.getFullYear()}-${(d.getMonth() + 1).toString().padStart(2, '0')}-${d.getDate().toString().padStart(2, '0')} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
}

const formatRelativeTime = (t: string) => {
  const d = new Date(t)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 7) return `${days}天前`
  return formatTime(t).slice(5)
}

const handleNewMessage = () => {
  loadMessages()
  loadUnreadCount()
}

onMounted(() => {
  loadMessages()
  loadUnreadCount()
  wsService.on('new_application', handleNewMessage)
  wsService.on('review_result', handleNewMessage)
  wsService.on('application_withdrawn', handleNewMessage)
})

onUnmounted(() => {
  wsService.off('new_application', handleNewMessage)
  wsService.off('review_result', handleNewMessage)
  wsService.off('application_withdrawn', handleNewMessage)
})
</script>

<style scoped>
.messages-page {
  max-width: 860px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.header-left {
  flex: 1;
}

.page-title-text {
  font-size: 26px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--text-primary), var(--primary-300));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin-bottom: 6px;
}

.page-desc {
  color: var(--text-secondary);
  font-size: 14px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 14px;
}

.unread-summary {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-secondary);
}

.unread-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 22px;
  height: 22px;
  padding: 0 7px;
  background: linear-gradient(135deg, var(--danger), #dc2626);
  color: white;
  font-size: 12px;
  font-weight: 600;
  border-radius: 11px;
}

.btn-mark-all {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  background: linear-gradient(135deg, var(--primary-600), var(--primary-500));
  border: none;
  border-radius: 10px;
  color: white;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 14px;
}

.btn-mark-all:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 15px rgba(99, 102, 241, 0.4);
}

/* 统计条 */
.stats-bar {
  display: flex;
  align-items: center;
  padding: 18px 28px;
  margin-bottom: 24px;
  gap: 0;
}

.stat-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  padding: 8px 16px;
  border-radius: 12px;
  transition: all 0.2s;
}

.stat-item:hover {
  background: var(--bg-glass-hover);
}

.stat-item.active {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.12), rgba(6, 182, 212, 0.06));
}

.stat-item.active .stat-num {
  background: linear-gradient(135deg, var(--primary-400), var(--accent-cyan));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.stat-num {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  transition: all 0.2s;
}

.stat-num.unread {
  color: var(--danger);
}

.stat-label {
  font-size: 12px;
  color: var(--text-muted);
}

.stat-divider {
  width: 1px;
  height: 36px;
  background: var(--border-glass);
}

/* 空状态 */
.empty-state {
  padding: 80px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.empty-icon {
  width: 90px;
  height: 90px;
  border-radius: 50%;
  background: var(--bg-glass);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  margin-bottom: 8px;
}

.empty-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.empty-text {
  color: var(--text-muted);
  font-size: 13px;
  margin: 0;
}

/* 消息列表 */
.message-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.message-item {
  display: flex;
  gap: 16px;
  position: relative;
}

.message-item:not(:last-child) {
  margin-bottom: 4px;
}

/* 时间轴 */
.message-timeline {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex-shrink: 0;
  padding-top: 18px;
  width: 32px;
}

.timeline-dot {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  z-index: 1;
  transition: all 0.2s;
}

.timeline-dot.info {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.2), rgba(99, 102, 241, 0.1));
  color: var(--primary-400);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.timeline-dot.success {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.2), rgba(16, 185, 129, 0.1));
  color: var(--success);
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
}

.timeline-dot.warning {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.2), rgba(245, 158, 11, 0.1));
  color: var(--warning);
  box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.1);
}

.message-item.unread .timeline-dot.info {
  box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.2);
}

.message-item.unread .timeline-dot.success {
  box-shadow: 0 0 0 4px rgba(16, 185, 129, 0.2);
}

.message-item.unread .timeline-dot.warning {
  box-shadow: 0 0 0 4px rgba(245, 158, 11, 0.2);
}

.timeline-line {
  flex: 1;
  width: 2px;
  background: linear-gradient(to bottom, var(--border-glass), transparent);
  margin: 6px 0;
  min-height: 20px;
}

/* 消息卡片 */
.message-card {
  flex: 1;
  padding: 18px 22px;
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}

.message-card::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: transparent;
  transition: all 0.2s;
}

.message-card:hover {
  transform: translateX(4px);
}

.message-item.unread .message-card {
  background: rgba(99, 102, 241, 0.04);
}

.message-item.unread .message-card::before {
  background: linear-gradient(to bottom, var(--primary-500), var(--primary-400));
}

.card-main {
  position: relative;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
  gap: 12px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 0;
}

.unread-dot {
  width: 8px;
  height: 8px;
  background: var(--danger);
  border-radius: 50%;
  flex-shrink: 0;
  box-shadow: 0 0 8px rgba(239, 68, 68, 0.5);
  animation: pulse-dot 2s ease-in-out infinite;
}

@keyframes pulse-dot {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.card-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.4;
}

.message-item:not(.unread) .card-title {
  font-weight: 500;
  color: var(--text-secondary);
}

.biz-tag {
  flex-shrink: 0;
  border-radius: 6px !important;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.card-time {
  font-size: 12px;
  color: var(--text-muted);
  white-space: nowrap;
}

.expand-icon {
  font-size: 14px;
  color: var(--text-muted);
  transition: transform 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.expand-icon.expanded {
  transform: rotate(180deg);
}

.card-content {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.7;
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  transition: all 0.3s ease;
}

.card-content.expanded {
  -webkit-line-clamp: unset;
  color: var(--text-primary);
}

.card-actions {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--border-glass);
  display: flex;
  justify-content: flex-start;
}

.action-link {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 14px;
  background: rgba(99, 102, 241, 0.1);
  border: none;
  border-radius: 8px;
  color: var(--primary-300);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.action-link:hover {
  background: rgba(99, 102, 241, 0.2);
  transform: translateX(2px);
}

.card-detail {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px dashed var(--border-glass);
  animation: slideDown 0.25s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.detail-meta {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 12px;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.meta-label {
  font-size: 11px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.meta-value {
  font-size: 13px;
  color: var(--text-primary);
}

.meta-value.read {
  color: var(--text-muted);
}

.meta-value.unread {
  color: var(--danger);
  font-weight: 500;
}

/* 分页 */
.pagination-wrapper {
  padding: 20px;
  margin-top: 24px;
  display: flex;
  justify-content: center;
}

:deep(.el-radio-button__inner) {
  background: var(--bg-glass) !important;
  border-color: var(--border-glass) !important;
  color: var(--text-secondary) !important;
  box-shadow: none !important;
}

:deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
  background: linear-gradient(135deg, var(--primary-600), var(--primary-500)) !important;
  border-color: transparent !important;
  color: white !important;
}

:deep(.el-loading-mask) {
  background: rgba(10, 10, 30, 0.5) !important;
  backdrop-filter: blur(4px);
  border-radius: 16px;
}
</style>
