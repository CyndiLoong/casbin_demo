<template>
  <div class="my-apps-page animate-fade-in-up">
    <div class="page-header">
      <div>
        <h2 class="page-title">我的申请</h2>
        <p class="page-desc">查看您提交的API资源申请及审核状态</p>
      </div>
      <button class="btn-new-app" @click="$router.push('/apply')">
        <el-icon><Plus /></el-icon>
        新建申请
      </button>
    </div>

    <div class="filter-bar glass-card">
      <el-radio-group v-model="filterStatus" @change="handleFilter" class="status-radio">
        <el-radio-button :value="undefined">全部</el-radio-button>
        <el-radio-button :value="0">待审核</el-radio-button>
        <el-radio-button :value="1">已通过</el-radio-button>
        <el-radio-button :value="2">已拒绝</el-radio-button>
      </el-radio-group>
    </div>

    <div v-if="applications.length === 0 && !loading" class="empty-state glass-card">
      <el-icon :size="48" color="var(--text-muted)"><Document /></el-icon>
      <p class="empty-text">暂无申请记录</p>
      <button class="btn-gradient" @click="$router.push('/apply')">立即申请</button>
    </div>

    <div v-else class="app-list">
      <div v-for="app in applications" :key="app.id" class="app-card glass-card" @click="openDetail(app)">
        <div class="app-card-main">
          <div class="app-card-header">
            <div class="app-title-group">
              <h3 class="app-name">{{ app.resource_name }}</h3>
              <el-tag :type="getStatusType(app.status)" size="small" effect="dark">{{ app.status_text }}</el-tag>
            </div>
            <span class="app-time">{{ formatTime(app.created_at) }}</span>
          </div>
          <div class="app-info-row">
            <span class="app-tag">
              <el-icon><Box /></el-icon>
              {{ getTypeLabel(app.resource_type) }}
            </span>
            <span class="app-tag">
              <el-icon><Link /></el-icon>
              {{ app.api_name }}
            </span>
            <span class="app-tag">
              <el-icon><Odometer /></el-icon>
              QPS: {{ app.expected_qps }}
            </span>
          </div>
          <p class="app-purpose">{{ app.purpose }}</p>
          <div v-if="app.status !== 0 && app.review_comment" class="review-comment">
            <span class="comment-label">{{ app.status === 1 ? '通过说明：' : '拒绝原因：' }}</span>
            {{ app.review_comment }}
          </div>
          <div v-if="app.can_withdraw && app.withdraw_remain_ms && app.withdraw_remain_ms > 0" class="withdraw-countdown">
            <el-icon><Clock /></el-icon>
            <span>可在 {{ formatCountdown(app.withdraw_remain_ms) }} 内撤回</span>
            <button class="btn-withdraw-inline" @click.stop="handleWithdraw(app)" :disabled="withdrawing">
              撤回申请
            </button>
          </div>
        </div>
        <div class="app-card-arrow">
          <el-icon><ArrowRight /></el-icon>
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
        @size-change="loadData"
        @current-change="loadData"
        background
      />
    </div>

    <el-dialog v-model="detailVisible" title="申请详情" width="560px" destroy-on-close>
      <div v-if="currentApp" class="detail-content">
        <div class="detail-header">
          <h3 class="detail-title">{{ currentApp.resource_name }}</h3>
          <el-tag :type="getStatusType(currentApp.status)" effect="dark">{{ currentApp.status_text }}</el-tag>
        </div>

        <div v-if="currentApp.can_withdraw && currentApp.withdraw_remain_ms && currentApp.withdraw_remain_ms > 0" class="withdraw-banner">
          <el-icon><Clock /></el-icon>
          <span>可在 <strong>{{ formatCountdown(currentApp.withdraw_remain_ms) }}</strong> 内撤回申请</span>
          <button class="btn-withdraw" @click="handleWithdraw(currentApp)" :disabled="withdrawing">
            撤回申请
          </button>
        </div>

        <div class="detail-grid">
          <div class="detail-item">
            <span class="detail-label">资源类型</span>
            <span class="detail-value">{{ getTypeLabel(currentApp.resource_type) }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">API名称</span>
            <span class="detail-value">{{ currentApp.api_name }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">预期QPS</span>
            <span class="detail-value">{{ currentApp.expected_qps }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">申请时间</span>
            <span class="detail-value">{{ formatTime(currentApp.created_at) }}</span>
          </div>
          <div v-if="currentApp.contact_info" class="detail-item">
            <span class="detail-label">联系方式</span>
            <span class="detail-value">{{ currentApp.contact_info }}</span>
          </div>
        </div>

        <div v-if="currentApp.api_description" class="detail-block">
          <h4>API描述</h4>
          <p>{{ currentApp.api_description }}</p>
        </div>
        <div class="detail-block">
          <h4>使用目的</h4>
          <p>{{ currentApp.purpose }}</p>
        </div>

        <div v-if="currentApp.status !== 0" class="detail-block review-block" :class="{ approved: currentApp.status === 1, rejected: currentApp.status === 2 }">
          <h4>审核结果</h4>
          <div class="review-info">
            <span v-if="currentApp.reviewer_name">审核人：{{ currentApp.reviewer_name }}</span>
            <span v-if="currentApp.reviewed_at">{{ formatTime(currentApp.reviewed_at) }}</span>
          </div>
          <p v-if="currentApp.review_comment">{{ currentApp.review_comment }}</p>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, onBeforeUnmount } from 'vue'
import { Plus, Document, Box, Link, Odometer, ArrowRight, Clock } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getMyApplications, withdrawAudit, type AuditApplication } from '@/api/audit'
import { wsService } from '@/utils/websocket'

const loading = ref(false)
const withdrawing = ref(false)
const applications = ref<AuditApplication[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const filterStatus = ref<number | undefined>(undefined)
const detailVisible = ref(false)
const currentApp = ref<AuditApplication | null>(null)
const countdownTimers = ref<Map<number, number>>(new Map())

const loadData = async () => {
  loading.value = true
  try {
    const res = await getMyApplications({
      page: page.value,
      page_size: pageSize.value,
      status: filterStatus.value
    })
    applications.value = res.data.list
    total.value = res.data.total
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const handleFilter = () => {
  page.value = 1
  loadData()
}

const openDetail = (app: AuditApplication) => {
  currentApp.value = app
  detailVisible.value = true
}

const getStatusType = (status: number) => {
  switch (status) {
    case 0: return 'warning'
    case 1: return 'success'
    case 2: return 'danger'
    default: return 'info'
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
  return `${d.getFullYear()}-${(d.getMonth()+1).toString().padStart(2,'0')}-${d.getDate().toString().padStart(2,'0')} ${d.getHours().toString().padStart(2,'0')}:${d.getMinutes().toString().padStart(2,'0')}`
}

const handleWithdraw = async (app: AuditApplication) => {
  try {
    await ElMessageBox.confirm(
      `确定要撤回「${app.resource_name}」的API申请吗？撤回后需重新提交。`,
      '撤回申请',
      {
        confirmButtonText: '确认撤回',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
  } catch {
    return
  }

  withdrawing.value = true
  try {
    await withdrawAudit(app.id)
    ElMessage.success('申请已撤回')
    loadData()
    if (currentApp.value?.id === app.id) {
      detailVisible.value = false
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '撤回失败')
  } finally {
    withdrawing.value = false
  }
}

const formatCountdown = (ms: number) => {
  if (ms <= 0) return '00:00'
  const totalSeconds = Math.floor(ms / 1000)
  const minutes = Math.floor(totalSeconds / 60)
  const seconds = totalSeconds % 60
  return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`
}

const handleReviewWS = () => {
  loadData()
}

const handleWithdrawWS = () => {
  loadData()
}

let countdownInterval: number | null = null

const startCountdown = () => {
  if (countdownInterval) return
  countdownInterval = window.setInterval(() => {
    applications.value.forEach(app => {
      if (app.can_withdraw && app.withdraw_remain_ms && app.withdraw_remain_ms > 0) {
        app.withdraw_remain_ms = Math.max(0, app.withdraw_remain_ms - 1000)
      }
    })
  }, 1000)
}

const stopCountdown = () => {
  if (countdownInterval) {
    clearInterval(countdownInterval)
    countdownInterval = null
  }
}

onMounted(() => {
  loadData()
  startCountdown()
  wsService.on('review_result', handleReviewWS)
  wsService.on('application_withdrawn', handleWithdrawWS)
  wsService.on('withdraw_confirmed', handleWithdrawWS)
})

onBeforeUnmount(() => {
  stopCountdown()
  wsService.off('review_result', handleReviewWS)
  wsService.off('application_withdrawn', handleWithdrawWS)
  wsService.off('withdraw_confirmed', handleWithdrawWS)
})
</script>

<style scoped>
.my-apps-page {
  max-width: 900px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.page-title {
  font-size: 24px;
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

.btn-new-app {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  background: linear-gradient(135deg, var(--primary-600), var(--accent-cyan));
  border: none;
  border-radius: 12px;
  color: white;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-new-app:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(99,102,241,0.4);
}

.filter-bar {
  padding: 12px 20px;
  margin-bottom: 20px;
}

.empty-state {
  padding: 60px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.empty-text {
  color: var(--text-muted);
  font-size: 14px;
}

.app-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.app-card {
  padding: 20px 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  cursor: pointer;
}

.app-card-main {
  flex: 1;
  min-width: 0;
}

.app-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.app-title-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.app-name {
  font-size: 16px;
  font-weight: 600;
}

.app-time {
  font-size: 12px;
  color: var(--text-muted);
}

.app-info-row {
  display: flex;
  gap: 16px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}

.app-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--text-secondary);
}

.app-purpose {
  font-size: 13px;
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  line-height: 1.5;
}

.review-comment {
  margin-top: 10px;
  font-size: 13px;
  color: var(--text-secondary);
  background: var(--bg-glass);
  padding: 8px 12px;
  border-radius: 8px;
  border-left: 3px solid var(--primary-500);
}

.comment-label {
  color: var(--text-muted);
  font-weight: 500;
}

.app-card-arrow {
  color: var(--text-muted);
  flex-shrink: 0;
}

.pagination-wrapper {
  padding: 16px 20px;
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

.detail-content {
  color: var(--text-primary);
}

.detail-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}

.detail-title {
  font-size: 18px;
  font-weight: 600;
}

.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px 24px;
  margin-bottom: 16px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-label {
  font-size: 12px;
  color: var(--text-muted);
}

.detail-value {
  font-size: 14px;
}

.detail-block {
  margin-bottom: 16px;
}

.detail-block h4 {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 6px;
}

.detail-block p {
  font-size: 14px;
  color: var(--text-primary);
  line-height: 1.6;
  background: var(--bg-glass);
  padding: 12px 16px;
  border-radius: 10px;
}

.review-block.approved p {
  border-left: 3px solid var(--success);
}

.review-block.rejected p {
  border-left: 3px solid var(--danger);
}

.review-info {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 6px;
}

.withdraw-countdown {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
  padding: 10px 14px;
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.2);
  border-radius: 10px;
  font-size: 13px;
  color: var(--warning);
}

.btn-withdraw-inline {
  margin-left: auto;
  padding: 4px 12px;
  background: rgba(245, 158, 11, 0.2);
  border: none;
  border-radius: 6px;
  color: var(--warning);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-withdraw-inline:hover {
  background: rgba(245, 158, 11, 0.3);
}

.btn-withdraw-inline:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.withdraw-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 18px;
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.15), rgba(249, 115, 22, 0.1));
  border: 1px solid rgba(245, 158, 11, 0.3);
  border-radius: 12px;
  margin-bottom: 20px;
  font-size: 14px;
  color: var(--warning);
}

.withdraw-banner strong {
  color: var(--warning);
  font-weight: 600;
}

.btn-withdraw {
  margin-left: auto;
  padding: 8px 20px;
  background: linear-gradient(135deg, var(--warning), #f97316);
  border: none;
  border-radius: 8px;
  color: white;
  font-weight: 500;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-withdraw:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.4);
}

.btn-withdraw:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

:deep(.el-radio-group) {
  display: flex;
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
</style>
