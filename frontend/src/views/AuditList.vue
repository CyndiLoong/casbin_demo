<template>
  <div class="audit-page animate-fade-in-up">
    <div class="page-header">
      <div class="page-header-main">
        <h2 class="page-title">审核管理</h2>
        <div class="header-stats">
          <div class="stat-chip pending">
            <el-icon><Clock /></el-icon>
            <span>待审核: <strong>{{ pendingCount }}</strong></span>
          </div>
          <div class="stat-chip approved">
            <el-icon><CircleCheck /></el-icon>
            <span>已通过: <strong>{{ approvedCount }}</strong></span>
          </div>
          <div class="stat-chip rejected">
            <el-icon><Close /></el-icon>
            <span>已驳回: <strong>{{ rejectedCount }}</strong></span>
          </div>
        </div>
      </div>
    </div>

    <div class="tabs-bar glass-card">
      <button
        v-for="tab in tabs"
        :key="tab.value"
        :class="['tab-btn', { active: activeTab === tab.value }]"
        @click="switchTab(tab.value)"
      >
        <el-icon :size="16"><component :is="tab.icon" /></el-icon>
        <span>{{ tab.label }}</span>
        <el-badge v-if="tab.badge && tab.badge() > 0" :value="tab.badge()" :max="99" class="tab-badge" />
      </button>
    </div>

    <div class="filter-bar glass-card" v-if="activeTab === 'history'">
      <el-select v-model="filterStatus" placeholder="全部状态" clearable style="width: 140px" @change="handleFilter">
        <el-option label="已通过" :value="1" />
        <el-option label="已驳回" :value="2" />
        <el-option label="已撤回" :value="3" />
      </el-select>
      <el-input v-model="filterApplicant" placeholder="搜索申请人" clearable style="width: 200px" @keyup.enter="handleFilter">
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <button class="btn-filter" @click="handleFilter">
        <el-icon><Search /></el-icon>搜索
      </button>
    </div>

    <div class="filter-bar glass-card" v-else>
      <el-input v-model="filterApplicant" placeholder="搜索申请人" clearable style="width: 240px" @keyup.enter="handleFilter">
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <button class="btn-filter" @click="handleFilter">
        <el-icon><Search /></el-icon>搜索
      </button>
      <div class="auto-refresh">
        <el-icon><Refresh /></el-icon>
        <span>实时更新</span>
      </div>
    </div>

    <div class="table-container glass-card">
      <div class="table-scroll-wrapper">
        <el-table
          :data="applications"
          v-loading="loading"
          style="width: 100%"
          @row-click="(row: AuditApplication) => openDetail(row)"
          row-class-name="clickable-row"
          :table-layout="'auto'"
        >
          <el-table-column prop="id" label="ID" width="55" />
          <el-table-column prop="applicant_name" label="申请人" min-width="90" />
          <el-table-column prop="resource_name" label="资源名称" min-width="120" show-overflow-tooltip />
          <el-table-column prop="resource_type" label="类型" width="105">
            <template #default="{ row }">
              <el-tag size="small" effect="dark" :type="getTypeColor(row.resource_type)">
                {{ getTypeLabel(row.resource_type) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="api_name" label="API" min-width="140" show-overflow-tooltip>
            <template #default="{ row }">
              <code class="api-code-inline">{{ row.api_name }}</code>
            </template>
          </el-table-column>
          <el-table-column prop="expected_qps" label="QPS" width="70" align="center" />
          <el-table-column v-if="activeTab === 'history'" prop="reviewer_name" label="审核人" min-width="80" />
          <el-table-column prop="status" label="状态" width="85">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)" size="small" effect="dark">{{ row.status_text }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column :prop="activeTab === 'history' ? 'reviewed_at' : 'created_at'" :label="activeTab === 'history' ? '审核时间' : '申请时间'" min-width="140">
            <template #default="{ row }">
              <span class="time-cell">{{ formatTime(activeTab === 'history' ? (row.reviewed_at || row.created_at) : row.created_at) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="80" fixed="right">
            <template #default="{ row }">
              <button class="action-btn view" @click.stop="openDetail(row)">
                {{ row.status === 0 ? '审核' : '查看' }}
              </button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div v-if="applications.length === 0 && !loading" class="table-empty">
        <el-icon :size="48" color="var(--text-muted)"><Document /></el-icon>
        <p>{{ activeTab === 'pending' ? '暂无待审核申请' : '暂无审核记录' }}</p>
      </div>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadData"
          @current-change="loadData"
          background
        />
      </div>
    </div>

    <el-dialog
      v-model="detailVisible"
      :title="currentApp?.status === 0 ? '审核申请' : '申请详情'"
      width="860px"
      :close-on-click-modal="false"
      destroy-on-close
      class="audit-dialog"
      align-center
    >
      <div v-if="currentApp" class="detail-content">
        <div class="detail-status-bar" :class="getStatusBarClass(currentApp.status)">
          <el-icon :size="20"><component :is="getStatusIcon(currentApp.status)" /></el-icon>
          <div class="status-info">
            <span class="status-label">当前状态</span>
            <span class="status-value">{{ currentApp.status_text }}</span>
          </div>
          <el-tag :type="getStatusType(currentApp.status)" size="large" effect="dark">{{ currentApp.status_text }}</el-tag>
        </div>

        <div class="detail-section">
          <h4 class="section-heading">
            <el-icon><User /></el-icon>
            申请人信息
          </h4>
          <div class="detail-grid">
            <div class="detail-item">
              <span class="detail-label">申请人</span>
              <span class="detail-value">{{ currentApp.applicant_name }}</span>
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
        </div>

        <div class="detail-section">
          <h4 class="section-heading">
            <el-icon><Box /></el-icon>
            资源信息
          </h4>
          <div class="detail-grid">
            <div class="detail-item">
              <span class="detail-label">资源名称</span>
              <span class="detail-value highlight">{{ currentApp.resource_name }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">资源类型</span>
              <el-tag size="small" effect="dark" :type="getTypeColor(currentApp.resource_type)">
                {{ getTypeLabel(currentApp.resource_type) }}
              </el-tag>
            </div>
            <div class="detail-item">
              <span class="detail-label">API名称</span>
              <code class="api-code">{{ currentApp.api_name }}</code>
            </div>
            <div class="detail-item">
              <span class="detail-label">预期QPS</span>
              <span class="detail-value">{{ currentApp.expected_qps }}</span>
            </div>
          </div>
        </div>

        <div v-if="currentApp.api_description" class="detail-section">
          <h4 class="section-heading">
            <el-icon><Document /></el-icon>
            API描述
          </h4>
          <p class="detail-text">{{ currentApp.api_description }}</p>
        </div>

        <div class="detail-section">
          <h4 class="section-heading">
            <el-icon><EditPen /></el-icon>
            使用目的
          </h4>
          <p class="detail-text purpose">{{ currentApp.purpose }}</p>
        </div>

        <div v-if="currentApp.status !== 0" class="detail-section">
          <h4 class="section-heading">
            <el-icon><Checked /></el-icon>
            审核信息
          </h4>
          <div class="detail-grid">
            <div class="detail-item">
              <span class="detail-label">审核人</span>
              <span class="detail-value">{{ currentApp.reviewer_name || '-' }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">审核时间</span>
              <span class="detail-value">{{ currentApp.reviewed_at ? formatTime(currentApp.reviewed_at) : '-' }}</span>
            </div>
          </div>
          <div v-if="currentApp.review_comment" class="review-result" :class="currentApp.status === 1 ? 'approved' : 'rejected'">
            <div class="review-result-header">
              <el-icon><component :is="currentApp.status === 1 ? CircleCheck : Close" /></el-icon>
              <span>{{ currentApp.status === 1 ? '审核通过' : '审核驳回' }}</span>
            </div>
            <p>{{ currentApp.review_comment }}</p>
          </div>
        </div>

        <div v-if="currentApp.status === 0" class="review-form">
          <el-divider />
          <h4 class="section-heading">
            <el-icon><Edit /></el-icon>
            审核操作
          </h4>
          <el-form :model="reviewForm" label-position="top">
            <el-form-item label="审核备注">
              <el-input
                v-model="reviewForm.comment"
                type="textarea"
                :rows="3"
                :placeholder="reviewForm.comment ? '' : '请输入审核意见，驳回时建议填写原因'"
                class="glass-input"
              />
            </el-form-item>
            <div class="review-actions">
              <button class="btn-reject" @click="handleReview(false)" :disabled="reviewing">
                <el-icon v-if="reviewing" class="is-loading"><Loading /></el-icon>
                <el-icon v-else><Close /></el-icon>
                {{ reviewing ? '提交中...' : '驳回申请' }}
              </button>
              <button class="btn-approve" @click="handleReview(true)" :disabled="reviewing">
                <el-icon v-if="reviewing" class="is-loading"><Loading /></el-icon>
                <el-icon v-else><Check /></el-icon>
                {{ reviewing ? '提交中...' : '通过申请' }}
              </button>
            </div>
          </el-form>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { Search, Loading, Clock, CircleCheck, Close, Check, Document, User, Box, EditPen, Edit, Checked, Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAllApplications, reviewAudit, getPendingCount, type AuditApplication } from '@/api/audit'
import { wsService } from '@/utils/websocket'

const loading = ref(false)
const applications = ref<AuditApplication[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const filterStatus = ref<number | undefined>(undefined)
const filterApplicant = ref('')
const pendingCount = ref(0)
const approvedCount = ref(0)
const rejectedCount = ref(0)
const activeTab = ref<'pending' | 'history'>('pending')

const detailVisible = ref(false)
const currentApp = ref<AuditApplication | null>(null)
const reviewing = ref(false)
const reviewForm = reactive({ comment: '' })

const tabs = computed(() => [
  {
    value: 'pending' as const,
    label: '待审核',
    icon: Clock,
    badge: (): number => pendingCount.value
  },
  {
    value: 'history' as const,
    label: '审核历史',
    icon: Document,
    badge: (): number => 0
  }
])

const effectiveStatus = computed(() => {
  if (activeTab.value === 'pending') return 0
  return filterStatus.value
})

const isExcludePending = computed(() => {
  return activeTab.value === 'history' && filterStatus.value === undefined
})

const loadStats = async () => {
  try {
    const [pendingRes] = await Promise.all([
      getPendingCount(),
    ])
    pendingCount.value = pendingRes.data.count

    const approvedRes = await getAllApplications({ page: 1, page_size: 1, status: 1 })
    approvedCount.value = approvedRes.data.total
    const rejectedRes = await getAllApplications({ page: 1, page_size: 1, status: 2 })
    rejectedCount.value = rejectedRes.data.total
  } catch (e) {
    console.error(e)
  }
}

const loadData = async () => {
  loading.value = true
  try {
    const params: any = {
      page: page.value,
      page_size: pageSize.value,
    }
    if (effectiveStatus.value !== undefined) {
      params.status = effectiveStatus.value
    } else if (isExcludePending.value) {
      params.exclude_pending = true
    }
    if (filterApplicant.value) {
      params.applicant = filterApplicant.value
    }
    const res = await getAllApplications(params)
    applications.value = res.data.list
    total.value = res.data.total
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const switchTab = (tab: 'pending' | 'history') => {
  activeTab.value = tab
  page.value = 1
  filterStatus.value = undefined
  filterApplicant.value = ''
  loadData()
}

const handleFilter = () => {
  page.value = 1
  loadData()
}

const openDetail = (row: AuditApplication) => {
  currentApp.value = { ...row }
  reviewForm.comment = ''
  detailVisible.value = true
}

const handleReview = async (approved: boolean) => {
  if (!currentApp.value) return
  if (!approved && !reviewForm.comment) {
    try {
      await ElMessageBox.confirm('驳回申请时建议填写审核备注说明原因，是否继续？', '提示', {
        confirmButtonText: '继续驳回',
        cancelButtonText: '返回填写',
        type: 'warning'
      })
    } catch {
      return
    }
  }
  reviewing.value = true
  try {
    await reviewAudit(currentApp.value.id, {
      approved,
      comment: reviewForm.comment || undefined
    })
    ElMessage.success(approved ? '已通过申请' : '已驳回申请')
    detailVisible.value = false
    loadData()
    loadStats()
  } catch (e) {
    console.error(e)
  } finally {
    reviewing.value = false
  }
}

const getStatusType = (status: number) => {
  switch (status) {
    case 0: return 'warning'
    case 1: return 'success'
    case 2: return 'danger'
    case 3: return 'info'
    default: return 'info'
  }
}

const getStatusBarClass = (status: number) => {
  switch (status) {
    case 0: return 'pending'
    case 1: return 'approved'
    case 2: return 'rejected'
    case 3: return 'withdrawn'
    default: return ''
  }
}

const getStatusIcon = (status: number) => {
  switch (status) {
    case 0: return Clock
    case 1: return CircleCheck
    case 2: return Close
    case 3: return Document
    default: return Clock
  }
}

const getTypeColor = (t: string) => {
  const map: Record<string, string> = {
    llm_chat: '', llm_code: 'success', image_gen: 'warning',
    asr: 'info', tts: 'info', embedding: '', other: 'info'
  }
  return map[t] || 'info'
}

const getTypeLabel = (t: string) => {
  const map: Record<string, string> = {
    llm_chat: '对话大模型', llm_code: '代码大模型', image_gen: '图像生成',
    asr: '语音识别', tts: '语音合成', embedding: '向量嵌入', other: '其他'
  }
  return map[t] || t
}

const formatTime = (t: string) => {
  if (!t) return '-'
  const d = new Date(t)
  return `${d.getFullYear()}-${(d.getMonth()+1).toString().padStart(2,'0')}-${d.getDate().toString().padStart(2,'0')} ${d.getHours().toString().padStart(2,'0')}:${d.getMinutes().toString().padStart(2,'0')}`
}

const handleWS = () => {
  loadData()
  loadStats()
}

onMounted(() => {
  loadData()
  loadStats()
  wsService.on('new_application', handleWS)
  wsService.on('review_result', handleWS)
  wsService.on('application_withdrawn', handleWS)
  wsService.setPendingCountCallback((c) => { pendingCount.value = c })
})

onUnmounted(() => {
  wsService.off('new_application', handleWS)
  wsService.off('review_result', handleWS)
  wsService.off('application_withdrawn', handleWS)
})
</script>

<style scoped>
.audit-page {
  padding: 0;
}

.page-header {
  margin-bottom: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-header-main {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
}

.page-title {
  font-size: 24px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--text-primary), var(--primary-300));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.header-stats {
  display: flex;
  gap: 10px;
}

.stat-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
}

.stat-chip.pending {
  background: rgba(245, 158, 11, 0.15);
  color: var(--warning);
  border: 1px solid rgba(245, 158, 11, 0.2);
}

.stat-chip.approved {
  background: rgba(16, 185, 129, 0.15);
  color: var(--success);
  border: 1px solid rgba(16, 185, 129, 0.2);
}

.stat-chip.rejected {
  background: rgba(239, 68, 68, 0.15);
  color: var(--danger);
  border: 1px solid rgba(239, 68, 68, 0.2);
}

.stat-chip strong {
  font-weight: 700;
  font-size: 14px;
}

.tabs-bar {
  display: flex;
  gap: 4px;
  padding: 8px;
  margin-bottom: 16px;
}

.tab-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 20px;
  background: transparent;
  border: none;
  border-radius: 12px;
  color: var(--text-secondary);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
}

.tab-btn:hover {
  background: var(--bg-glass);
  color: var(--text-primary);
}

.tab-btn.active {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.2), rgba(6, 182, 212, 0.1));
  color: var(--primary-300);
}

.tab-badge :deep(.el-badge__content) {
  background: var(--danger);
  border: none;
  font-size: 10px;
  height: 18px;
  line-height: 18px;
  padding: 0 5px;
}

.filter-bar {
  display: flex;
  gap: 12px;
  align-items: center;
  padding: 14px 20px;
  margin-bottom: 16px;
}

.auto-refresh {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--success);
  padding: 4px 12px;
  background: rgba(16, 185, 129, 0.1);
  border-radius: 20px;
}

.btn-filter {
  padding: 8px 18px;
  background: linear-gradient(135deg, var(--primary-600), var(--primary-500));
  border: none;
  border-radius: 10px;
  color: white;
  font-size: 14px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s;
}

.btn-filter:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 15px rgba(99,102,241,0.4);
}

.table-container {
  padding: 16px;
}

.table-scroll-wrapper {
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}

.api-code-inline {
  background: rgba(6, 182, 212, 0.1);
  padding: 2px 8px;
  border-radius: 6px;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 12px;
  color: var(--accent-cyan);
  white-space: nowrap;
}

.time-cell {
  font-size: 13px;
  white-space: nowrap;
}

.clickable-row {
  cursor: pointer;
}

.table-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 60px 20px;
  color: var(--text-muted);
  font-size: 14px;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

.action-btn {
  padding: 6px 14px;
  background: rgba(99,102,241,0.15);
  border: none;
  border-radius: 8px;
  color: var(--primary-300);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.action-btn:hover {
  background: rgba(99,102,241,0.3);
}

.detail-content {
  color: var(--text-primary);
}

.detail-status-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 18px 22px;
  border-radius: 14px;
  margin-bottom: 24px;
}

.detail-status-bar.pending {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.15), rgba(249, 115, 22, 0.08));
  border: 1px solid rgba(245, 158, 11, 0.25);
  color: var(--warning);
}

.detail-status-bar.approved {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.15), rgba(52, 211, 153, 0.08));
  border: 1px solid rgba(16, 185, 129, 0.25);
  color: var(--success);
}

.detail-status-bar.rejected {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.15), rgba(248, 113, 113, 0.08));
  border: 1px solid rgba(239, 68, 68, 0.25);
  color: var(--danger);
}

.detail-status-bar.withdrawn {
  background: var(--bg-glass);
  border: 1px solid var(--border-glass);
  color: var(--text-secondary);
}

.status-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.status-label {
  font-size: 12px;
  opacity: 0.7;
}

.status-value {
  font-size: 16px;
  font-weight: 600;
}

.detail-section {
  margin-bottom: 24px;
}

.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px 24px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.detail-item.full-width {
  grid-column: 1 / -1;
}

.detail-label {
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 500;
}

.detail-value {
  font-size: 14px;
  color: var(--text-primary);
}

.detail-value.highlight {
  font-weight: 600;
  font-size: 15px;
}

.api-code {
  background: var(--bg-glass);
  padding: 4px 10px;
  border-radius: 6px;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 13px;
  color: var(--accent-cyan);
}

.section-heading {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 14px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-glass);
}

.detail-text {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.7;
  background: var(--bg-glass);
  padding: 14px 18px;
  border-radius: 12px;
  border-left: 3px solid var(--primary-500);
}

.detail-text.purpose {
  border-left-color: var(--accent-cyan);
}

.review-result {
  margin-top: 12px;
  padding: 14px 18px;
  border-radius: 12px;
}

.review-result.approved {
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.2);
}

.review-result.rejected {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.2);
}

.review-result-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 8px;
}

.review-result.approved .review-result-header {
  color: var(--success);
}

.review-result.rejected .review-result-header {
  color: var(--danger);
}

.review-result p {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.6;
  margin: 0;
}

.review-form {
  margin-top: 8px;
}

.review-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.btn-reject {
  padding: 10px 28px;
  background: linear-gradient(135deg, var(--danger), #dc2626);
  border: none;
  border-radius: 10px;
  color: white;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn-reject:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 15px rgba(239,68,68,0.4);
}

.btn-approve {
  padding: 10px 28px;
  background: linear-gradient(135deg, var(--success), #059669);
  border: none;
  border-radius: 10px;
  color: white;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn-approve:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 15px rgba(16,185,129,0.4);
}

.btn-reject:disabled,
.btn-approve:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

:deep(.el-divider) {
  border-color: var(--border-glass);
  margin: 20px 0;
}

:deep(.el-textarea__inner) {
  background: rgba(255,255,255,0.03) !important;
  border: 1px solid var(--border-glass) !important;
  border-radius: 10px !important;
  color: var(--text-primary) !important;
}

:deep(.el-textarea__inner:focus) {
  border-color: var(--primary-500) !important;
}

:deep(.el-select .el-input__wrapper),
:deep(.el-input__wrapper) {
  background: rgba(255,255,255,0.03) !important;
  border-radius: 10px !important;
  box-shadow: 0 0 0 1px var(--border-glass) inset !important;
}

:deep(.el-select .el-input__wrapper:hover),
:deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px rgba(99,102,241,0.3) inset !important;
}

:deep(.el-select .el-input.is-focus .el-input__wrapper),
:deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--primary-500) inset, 0 0 0 3px rgba(99,102,241,0.1) !important;
}
</style>
