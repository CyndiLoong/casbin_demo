<template>
  <div class="dashboard" :class="{ 'admin-dashboard': isAdmin }">
    <div class="welcome-banner glass-card animate-fade-in-up">
      <div class="welcome-content">
        <div class="welcome-greeting">
          <h1>{{ greetingText }}，{{ user?.nickname || user?.username }} <span class="wave">👋</span></h1>
          <p v-if="isAdmin">您拥有系统管理员权限，可以管理用户、角色、权限和审核资源申请</p>
          <p v-else>您可以在此提交大模型API资源申请，查看申请进度和审核通知</p>
        </div>
        <div class="quick-actions" v-if="!isAdmin">
          <button class="action-btn primary" @click="$router.push('/apply')">
            <el-icon><Document /></el-icon>
            立即申请资源
          </button>
          <button class="action-btn secondary" @click="$router.push('/my-applications')">
            <el-icon><List /></el-icon>
            查看我的申请
          </button>
        </div>
        <div class="quick-actions" v-else>
          <button class="action-btn primary" @click="$router.push('/audit')">
            <el-icon><Checked /></el-icon>
            待审核申请
            <span v-if="pendingCount > 0" class="action-badge">{{ pendingCount }}</span>
          </button>
          <button class="action-btn secondary" @click="$router.push('/users')">
            <el-icon><User /></el-icon>
            用户管理
          </button>
        </div>
      </div>
      <div class="welcome-visual">
        <div class="pulse-ring"></div>
        <div class="pulse-ring delay"></div>
        <el-icon :size="56" :color="isAdmin ? '#f59e0b' : 'var(--primary-400)'">
          <Key v-if="isAdmin" />
          <DataBoard v-else />
        </el-icon>
      </div>
    </div>

    <div v-if="isAdmin" class="stats-grid">
      <div v-for="(s, i) in adminStatCards" :key="i" class="stat-card glass-card animate-fade-in-up" :style="{ animationDelay: `${i * 0.1}s` }">
        <div class="stat-icon" :style="{ background: s.gradient }">
          <el-icon :size="24"><component :is="s.icon" /></el-icon>
        </div>
        <div class="stat-info">
          <p class="stat-value">{{ s.value }}</p>
          <p class="stat-label">{{ s.label }}</p>
        </div>
      </div>
    </div>

    <div v-else class="stats-grid">
      <div v-for="(s, i) in userStatCards" :key="i" class="stat-card glass-card animate-fade-in-up" :style="{ animationDelay: `${i * 0.1}s` }">
        <div class="stat-icon" :style="{ background: s.gradient }">
          <el-icon :size="24"><component :is="s.icon" /></el-icon>
        </div>
        <div class="stat-info">
          <p class="stat-value">{{ s.value }}</p>
          <p class="stat-label">{{ s.label }}</p>
        </div>
      </div>
    </div>

    <div class="content-grid">
      <div v-if="isAdmin" class="panel-card glass-card animate-fade-in-up" style="animation-delay: 0.3s">
        <div class="panel-header">
          <h3><el-icon><Clock /></el-icon> 最近待审核</h3>
          <button class="view-all-btn" @click="$router.push('/audit')">查看全部 <el-icon><ArrowRight /></el-icon></button>
        </div>
        <div v-if="pendingApps.length === 0" class="panel-empty">
          <el-icon :size="36" color="var(--success)"><CircleCheck /></el-icon>
          <p>暂无待审核申请</p>
        </div>
        <div v-else class="list-items">
          <div v-for="app in pendingApps" :key="app.id" class="list-item" @click="$router.push('/audit')">
            <div class="item-avatar pending">
              <el-icon><Document /></el-icon>
            </div>
            <div class="item-content">
              <div class="item-title">{{ app.applicant_name }} 申请 {{ app.resource_name }}</div>
              <div class="item-desc">{{ app.api_name }} · QPS: {{ app.expected_qps }}</div>
              <div class="item-time">{{ formatTime(app.created_at) }}</div>
            </div>
            <el-tag type="warning" size="small" effect="dark">待审核</el-tag>
          </div>
        </div>
      </div>

      <div v-else class="panel-card glass-card animate-fade-in-up" style="animation-delay: 0.3s">
        <div class="panel-header">
          <h3><el-icon><Document /></el-icon> 最近申请</h3>
          <button class="view-all-btn" @click="$router.push('/my-applications')">查看全部 <el-icon><ArrowRight /></el-icon></button>
        </div>
        <div v-if="myRecentApps.length === 0" class="panel-empty">
          <el-icon :size="36" color="var(--text-muted)"><Document /></el-icon>
          <p>暂无申请记录</p>
          <button class="btn-gradient-sm" @click="$router.push('/apply')">立即申请</button>
        </div>
        <div v-else class="list-items">
          <div v-for="app in myRecentApps" :key="app.id" class="list-item" @click="$router.push('/my-applications')">
            <div class="item-avatar" :class="getStatusClass(app.status)">
              <el-icon><component :is="getStatusIcon(app.status)" /></el-icon>
            </div>
            <div class="item-content">
              <div class="item-title">{{ app.resource_name }}</div>
              <div class="item-desc">{{ getTypeLabel(app.resource_type) }} · {{ app.api_name }}</div>
              <div class="item-time">{{ formatTime(app.created_at) }}</div>
            </div>
            <el-tag :type="getStatusType(app.status)" size="small" effect="dark">{{ app.status_text }}</el-tag>
          </div>
        </div>
      </div>

      <div class="panel-card glass-card animate-fade-in-up" style="animation-delay: 0.4s">
        <div class="panel-header">
          <h3><el-icon><Bell /></el-icon> 最新消息</h3>
          <button class="view-all-btn" @click="$router.push('/messages')">查看全部 <el-icon><ArrowRight /></el-icon></button>
        </div>
        <div v-if="recentMessages.length === 0" class="panel-empty">
          <el-icon :size="36" color="var(--text-muted)"><Bell /></el-icon>
          <p>暂无消息</p>
        </div>
        <div v-else class="list-items">
          <div v-for="msg in recentMessages" :key="msg.id" class="list-item" @click="handleMsgClick(msg)">
            <div class="item-avatar msg">
              <el-icon><BellFilled /></el-icon>
            </div>
            <div class="item-content">
              <div class="item-title">{{ msg.title }}</div>
              <div class="item-desc">{{ msg.content }}</div>
              <div class="item-time">{{ formatTime(msg.created_at) }}</div>
            </div>
            <div v-if="!msg.is_read" class="unread-dot"></div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="!isAdmin" class="tips-card glass-card animate-fade-in-up" style="animation-delay: 0.5s">
      <h3><el-icon><InfoFilled /></el-icon> 使用指南</h3>
      <div class="tips-grid">
        <div class="tip-item">
          <div class="tip-num">1</div>
          <div class="tip-text">点击「申请资源」填写大模型API使用申请表单</div>
        </div>
        <div class="tip-item">
          <div class="tip-num">2</div>
          <div class="tip-text">提交后管理员将在工作日尽快处理您的申请</div>
        </div>
        <div class="tip-item">
          <div class="tip-num">3</div>
          <div class="tip-text">申请提交后2分钟内可撤回，审核结果将实时通知</div>
        </div>
        <div class="tip-item">
          <div class="tip-num">4</div>
          <div class="tip-text">通过后即可使用申请的API资源，留意消息通知</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { DataBoard, User, UserFilled, Key, Clock, Document, List, Bell, BellFilled, ArrowRight, CircleCheck, Checked, InfoFilled, Warning } from '@element-plus/icons-vue'
import { getDashboard } from '@/api/auth'
import { useUserStore } from '@/store/user'
import { getAllApplications, getMyApplications, getMessages, getPendingCount, markMessageRead, type AuditApplication, type SysMessage } from '@/api/audit'
import { wsService } from '@/utils/websocket'

const router = useRouter()
const userStore = useUserStore()
const user = computed(() => userStore.userInfo)
const isAdmin = computed(() => userStore.hasRole('admin'))

const adminStats = ref<any>(null)
const myStats = ref({ total: 0, pending: 0, approved: 0, rejected: 0 })
const pendingCount = ref(0)
const pendingApps = ref<AuditApplication[]>([])
const myRecentApps = ref<AuditApplication[]>([])
const recentMessages = ref<SysMessage[]>([])

const now = new Date()
const hour = now.getHours()
const greetingText = computed(() => {
  if (hour < 6) return '夜深了'
  if (hour < 9) return '早上好'
  if (hour < 12) return '上午好'
  if (hour < 14) return '中午好'
  if (hour < 18) return '下午好'
  if (hour < 22) return '晚上好'
  return '夜深了'
})

const adminStatCards = computed(() => [
  { icon: User, label: '用户总数', value: adminStats.value?.stats?.total_users || 0, gradient: 'linear-gradient(135deg, #6366f1, #818cf8)' },
  { icon: UserFilled, label: '角色数量', value: adminStats.value?.stats?.total_roles || 0, gradient: 'linear-gradient(135deg, #06b6d4, #22d3ee)' },
  { icon: Key, label: '权限节点', value: adminStats.value?.stats?.total_permissions || 0, gradient: 'linear-gradient(135deg, #a855f7, #c084fc)' },
  { icon: Warning, label: '待审核', value: pendingCount.value, gradient: 'linear-gradient(135deg, #f59e0b, #f97316)' }
])

const userStatCards = computed(() => [
  { icon: Document, label: '我的申请', value: myStats.value.total, gradient: 'linear-gradient(135deg, #6366f1, #818cf8)' },
  { icon: Clock, label: '审核中', value: myStats.value.pending, gradient: 'linear-gradient(135deg, #f59e0b, #f97316)' },
  { icon: CircleCheck, label: '已通过', value: myStats.value.approved, gradient: 'linear-gradient(135deg, #10b981, #34d399)' },
  { icon: Bell, label: '未读消息', value: unreadCount.value, gradient: 'linear-gradient(135deg, #06b6d4, #22d3ee)' }
])

const unreadCount = computed(() => recentMessages.value.filter(m => !m.is_read).length)

const loadAdminData = async () => {
  try {
    const [dashRes, pendingRes, appsRes] = await Promise.all([
      getDashboard(),
      getPendingCount(),
      getAllApplications({ page: 1, page_size: 5, status: 0 })
    ])
    adminStats.value = dashRes.data
    pendingCount.value = pendingRes.data.count
    pendingApps.value = appsRes.data.list
  } catch (e) {
    console.error(e)
  }
}

const loadUserData = async () => {
  try {
    const [appsAll, appsPending, appsApproved, appsRejected, msgsRes] = await Promise.all([
      getMyApplications({ page: 1, page_size: 5 }),
      getMyApplications({ page: 1, page_size: 1, status: 0 }),
      getMyApplications({ page: 1, page_size: 1, status: 1 }),
      getMyApplications({ page: 1, page_size: 1, status: 2 }),
      getMessages({ page: 1, page_size: 5 })
    ])
    myRecentApps.value = appsAll.data.list
    myStats.value = {
      total: appsAll.data.total,
      pending: appsPending.data.total,
      approved: appsApproved.data.total,
      rejected: appsRejected.data.total
    }
    recentMessages.value = msgsRes.data.list
  } catch (e) {
    console.error(e)
  }
}

const loadMessages = async () => {
  try {
    const res = await getMessages({ page: 1, page_size: 5 })
    recentMessages.value = res.data.list
  } catch (e) {
    console.error(e)
  }
}

const handleMsgClick = async (msg: SysMessage) => {
  if (!msg.is_read) {
    try {
      await markMessageRead(msg.id)
      msg.is_read = true
    } catch (e) {
      console.error(e)
    }
  }
  if (msg.business_type === 'audit_application') {
    if (isAdmin.value && (msg.type === 'new_application' || msg.type === 'application_withdrawn')) {
      router.push('/audit')
    } else {
      router.push('/my-applications')
    }
  } else {
    router.push('/messages')
  }
}

const getStatusType = (status: number) => {
  switch (status) {
    case 0: return 'warning'
    case 1: return 'success'
    case 2: return 'danger'
    default: return 'info'
  }
}

const getStatusClass = (status: number) => {
  switch (status) {
    case 0: return 'pending'
    case 1: return 'approved'
    case 2: return 'rejected'
    default: return ''
  }
}

const getStatusIcon = (status: number) => {
  switch (status) {
    case 0: return Clock
    case 1: return CircleCheck
    case 2: return Warning
    default: return Document
  }
}

const getTypeLabel = (t: string) => {
  const map: Record<string, string> = {
    llm_chat: '对话大模型', llm_code: '代码大模型', image_gen: '图像生成',
    asr: '语音识别', tts: '语音合成', embedding: '向量嵌入', other: '其他'
  }
  return map[t] || t
}

const formatTime = (t: string) => {
  const d = new Date(t)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  return `${d.getMonth() + 1}月${d.getDate()}日 ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
}

const handleWSUpdate = () => {
  if (isAdmin.value) {
    loadAdminData()
  } else {
    loadUserData()
  }
}

onMounted(() => {
  if (isAdmin.value) {
    loadAdminData()
  } else {
    loadUserData()
  }
  wsService.on('new_application', handleWSUpdate)
  wsService.on('review_result', handleWSUpdate)
  wsService.on('application_withdrawn', handleWSUpdate)
})

onUnmounted(() => {
  wsService.off('new_application', handleWSUpdate)
  wsService.off('review_result', handleWSUpdate)
  wsService.off('application_withdrawn', handleWSUpdate)
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
  padding: 32px 36px;
  margin-bottom: 24px;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.12), rgba(6, 182, 212, 0.06));
  border: 1px solid rgba(99, 102, 241, 0.15);
}

.welcome-content {
  flex: 1;
}

.welcome-greeting h1 {
  font-size: 26px;
  font-weight: 700;
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.wave {
  display: inline-block;
  animation: wave 2s ease-in-out infinite;
  transform-origin: 70% 70%;
}

@keyframes wave {
  0%, 100% { transform: rotate(0deg); }
  25% { transform: rotate(20deg); }
  75% { transform: rotate(-10deg); }
}

.welcome-greeting p {
  color: var(--text-secondary);
  margin-bottom: 20px;
  font-size: 14px;
}

.quick-actions {
  display: flex;
  gap: 12px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
  position: relative;
}

.action-btn.primary {
  background: linear-gradient(135deg, var(--primary-600), var(--primary-500));
  color: white;
}

.action-btn.primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(99, 102, 241, 0.4);
}

.action-btn.secondary {
  background: var(--bg-glass);
  border: 1px solid var(--border-glass);
  color: var(--text-primary);
}

.action-btn.secondary:hover {
  background: var(--bg-glass-hover);
  border-color: var(--primary-500);
}

.action-badge {
  position: absolute;
  top: -6px;
  right: -6px;
  background: var(--danger);
  color: white;
  font-size: 11px;
  min-width: 20px;
  height: 20px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 5px;
  font-weight: 600;
}

.welcome-visual {
  position: relative;
  width: 120px;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.pulse-ring {
  position: absolute;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  border: 2px solid var(--primary-500);
  animation: pulse-ring 2.5s ease-out infinite;
}

.pulse-ring.delay {
  animation-delay: 1.25s;
}

@keyframes pulse-ring {
  0% { transform: scale(0.7); opacity: 0.8; }
  100% { transform: scale(1.4); opacity: 0; }
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  padding: 22px;
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 52px;
  height: 52px;
  border-radius: 14px;
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

.content-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 20px;
}

.panel-card {
  padding: 24px;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.panel-header h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  margin: 0;
}

.view-all-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  background: none;
  border: none;
  color: var(--primary-400);
  font-size: 13px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: all 0.2s;
}

.view-all-btn:hover {
  background: rgba(99, 102, 241, 0.1);
}

.panel-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 40px 20px;
  color: var(--text-muted);
  font-size: 14px;
}

.list-items {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.list-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
}

.list-item:hover {
  background: var(--bg-glass);
}

.item-avatar {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: rgba(99, 102, 241, 0.15);
  color: var(--primary-400);
}

.item-avatar.pending {
  background: rgba(245, 158, 11, 0.15);
  color: var(--warning);
}

.item-avatar.approved {
  background: rgba(16, 185, 129, 0.15);
  color: var(--success);
}

.item-avatar.rejected {
  background: rgba(239, 68, 68, 0.15);
  color: var(--danger);
}

.item-avatar.msg {
  background: rgba(6, 182, 212, 0.15);
  color: var(--accent-cyan);
}

.item-content {
  flex: 1;
  min-width: 0;
}

.item-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-desc {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-time {
  font-size: 11px;
  color: var(--text-muted);
}

.unread-dot {
  width: 8px;
  height: 8px;
  background: var(--primary-500);
  border-radius: 50%;
  flex-shrink: 0;
  box-shadow: 0 0 8px rgba(99, 102, 241, 0.5);
}

.btn-gradient-sm {
  padding: 8px 20px;
  background: linear-gradient(135deg, var(--primary-600), var(--primary-500));
  border: none;
  border-radius: 8px;
  color: white;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-gradient-sm:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.4);
}

.tips-card {
  padding: 24px;
  margin-top: 4px;
}

.tips-card h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 16px 0;
  color: var(--primary-300);
}

.tips-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 12px;
}

.tip-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px;
  background: var(--bg-glass);
  border-radius: 10px;
}

.tip-num {
  width: 28px;
  height: 28px;
  background: linear-gradient(135deg, var(--primary-600), var(--primary-500));
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 700;
  font-size: 13px;
  flex-shrink: 0;
}

.tip-text {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.6;
  padding-top: 4px;
}

.dashboard.admin-dashboard .welcome-banner {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.12), rgba(249, 115, 22, 0.06));
  border-color: rgba(245, 158, 11, 0.2);
}

.dashboard.admin-dashboard .action-btn.primary {
  background: linear-gradient(135deg, #f59e0b, #f97316);
}

.dashboard.admin-dashboard .action-btn.primary:hover {
  box-shadow: 0 6px 20px rgba(245, 158, 11, 0.4);
}

.dashboard.admin-dashboard .action-btn.secondary:hover {
  border-color: #f59e0b;
}

.dashboard.admin-dashboard .pulse-ring {
  border-color: #f59e0b;
}

.dashboard.admin-dashboard .btn-gradient-sm {
  background: linear-gradient(135deg, #f59e0b, #f97316);
}

.dashboard.admin-dashboard .btn-gradient-sm:hover {
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.4);
}

.dashboard.admin-dashboard .view-all-btn {
  color: #f59e0b;
}

.dashboard.admin-dashboard .view-all-btn:hover {
  background: rgba(245, 158, 11, 0.1);
}

.dashboard.admin-dashboard .tips-card h3 {
  color: #f59e0b;
}

.dashboard.admin-dashboard .tip-num {
  background: linear-gradient(135deg, #f59e0b, #f97316);
}

.dashboard.admin-dashboard .unread-dot {
  background: #f59e0b;
  box-shadow: 0 0 8px rgba(245, 158, 11, 0.5);
}

.dashboard.admin-dashboard .stat-card:nth-child(4) .stat-icon {
  background: linear-gradient(135deg, #f59e0b, #f97316) !important;
}
</style>
