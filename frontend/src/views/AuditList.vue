<template>
  <div class="audit-page animate-fade-in-up">
    <div class="page-header">
      <div class="page-header-main">
        <h2 class="page-title">审核管理</h2>
        <span v-if="pendingCount > 0" class="pending-badge">
          {{ pendingCount }} 条待审核
        </span>
      </div>
    </div>

    <div class="filter-bar glass-card">
      <el-select v-model="filterStatus" placeholder="全部状态" clearable style="width: 140px" @change="handleFilter">
        <el-option label="待审核" :value="0" />
        <el-option label="已通过" :value="1" />
        <el-option label="已拒绝" :value="2" />
      </el-select>
      <el-input v-model="filterApplicant" placeholder="搜索申请人" clearable style="width: 200px" @keyup.enter="handleFilter">
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
      <button class="btn-filter" @click="handleFilter">
        <el-icon><Search /></el-icon>搜索
      </button>
    </div>

    <div class="table-container glass-card">
      <el-table
        :data="applications"
        v-loading="loading"
        style="width: 100%"
        @row-click="(row: AuditApplication) => openDetail(row)"
        row-class-name="clickable-row"
      >
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="applicant_name" label="申请人" width="120" />
        <el-table-column prop="resource_name" label="资源名称" min-width="140" />
        <el-table-column prop="resource_type" label="类型" width="130">
          <template #default="{ row }">
            <el-tag size="small" effect="dark" :type="getTypeColor(row.resource_type)">
              {{ getTypeLabel(row.resource_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="api_name" label="API" min-width="160" />
        <el-table-column prop="expected_qps" label="预期QPS" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small" effect="dark">{{ row.status_text }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="申请时间" width="160">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <button class="action-btn view" @click.stop="openDetail(row)">查看</button>
          </template>
        </el-table-column>
      </el-table>

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
      title="审核详情"
      width="640px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <div v-if="currentApp" class="detail-content">
        <div class="detail-section">
          <div class="detail-grid">
            <div class="detail-item">
              <span class="detail-label">申请人</span>
              <span class="detail-value">{{ currentApp.applicant_name }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">状态</span>
              <el-tag :type="getStatusType(currentApp.status)" size="small">{{ currentApp.status_text }}</el-tag>
            </div>
            <div class="detail-item">
              <span class="detail-label">资源名称</span>
              <span class="detail-value">{{ currentApp.resource_name }}</span>
            </div>
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
            <div v-if="currentApp.contact_info" class="detail-item">
              <span class="detail-label">联系方式</span>
              <span class="detail-value">{{ currentApp.contact_info }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">申请时间</span>
              <span class="detail-value">{{ formatTime(currentApp.created_at) }}</span>
            </div>
          </div>
        </div>

        <div v-if="currentApp.api_description" class="detail-section">
          <h4 class="section-heading">API描述</h4>
          <p class="detail-text">{{ currentApp.api_description }}</p>
        </div>

        <div class="detail-section">
          <h4 class="section-heading">使用目的</h4>
          <p class="detail-text">{{ currentApp.purpose }}</p>
        </div>

        <div v-if="currentApp.status !== 0" class="detail-section">
          <h4 class="section-heading">审核信息</h4>
          <div class="detail-grid">
            <div class="detail-item">
              <span class="detail-label">审核人</span>
              <span class="detail-value">{{ currentApp.reviewer_name || '-' }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">审核时间</span>
              <span class="detail-value">{{ currentApp.reviewed_at ? formatTime(currentApp.reviewed_at) : '-' }}</span>
            </div>
            <div v-if="currentApp.review_comment" class="detail-item full-width">
              <span class="detail-label">审核备注</span>
              <p class="detail-text">{{ currentApp.review_comment }}</p>
            </div>
          </div>
        </div>

        <div v-if="currentApp.status === 0" class="review-form">
          <el-divider />
          <h4 class="section-heading">审核操作</h4>
          <el-form :model="reviewForm" label-position="top">
            <el-form-item label="审核备注（选填）">
              <el-input
                v-model="reviewForm.comment"
                type="textarea"
                :rows="3"
                placeholder="请输入审核意见，例如通过原因或拒绝理由"
                class="glass-input"
              />
            </el-form-item>
            <div class="review-actions">
              <button class="btn-reject" @click="handleReview(false)" :disabled="reviewing">
                <el-icon v-if="reviewing" class="is-loading"><Loading /></el-icon>
                {{ reviewing ? '提交中...' : '拒绝' }}
              </button>
              <button class="btn-approve" @click="handleReview(true)" :disabled="reviewing">
                <el-icon v-if="reviewing" class="is-loading"><Loading /></el-icon>
                {{ reviewing ? '提交中...' : '通过' }}
              </button>
            </div>
          </el-form>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { Search, Loading } from '@element-plus/icons-vue'
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

const detailVisible = ref(false)
const currentApp = ref<AuditApplication | null>(null)
const reviewing = ref(false)
const reviewForm = reactive({ comment: '' })

const loadData = async () => {
  loading.value = true
  try {
    const res = await getAllApplications({
      page: page.value,
      page_size: pageSize.value,
      status: filterStatus.value,
      applicant: filterApplicant.value || undefined
    })
    applications.value = res.data.list
    total.value = res.data.total
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const loadPendingCount = async () => {
  try {
    const res = await getPendingCount()
    pendingCount.value = res.data.count
  } catch (e) {
    console.error(e)
  }
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
      await ElMessageBox.confirm('拒绝申请时建议填写审核备注，是否继续？', '提示', {
        confirmButtonText: '继续拒绝',
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
    ElMessage.success(approved ? '已通过申请' : '已拒绝申请')
    detailVisible.value = false
    loadData()
    loadPendingCount()
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
    default: return 'info'
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
  const d = new Date(t)
  return `${d.getFullYear()}-${(d.getMonth()+1).toString().padStart(2,'0')}-${d.getDate().toString().padStart(2,'0')} ${d.getHours().toString().padStart(2,'0')}:${d.getMinutes().toString().padStart(2,'0')}`
}

const handleWS = () => {
  loadData()
  loadPendingCount()
}

onMounted(() => {
  loadData()
  loadPendingCount()
  wsService.on('new_application', handleWS)
  wsService.setPendingCountCallback((c) => { pendingCount.value = c })
})

onUnmounted(() => {
  wsService.off('new_application', handleWS)
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
  gap: 12px;
}

.page-title {
  font-size: 24px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--text-primary), var(--primary-300));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.pending-badge {
  background: linear-gradient(135deg, var(--warning), #f97316);
  color: white;
  font-size: 12px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 20px;
  animation: pulse-glow 2s ease-in-out infinite;
}

.filter-bar {
  display: flex;
  gap: 12px;
  align-items: center;
  padding: 16px 20px;
  margin-bottom: 20px;
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
  padding: 20px;
}

.clickable-row {
  cursor: pointer;
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

.detail-section {
  margin-bottom: 20px;
}

.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px 24px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
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

.section-heading {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.detail-text {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.6;
  background: var(--bg-glass);
  padding: 12px 16px;
  border-radius: 10px;
  border-left: 3px solid var(--primary-500);
}

.review-form {
  margin-top: 16px;
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

:deep(.el-divider) {
  border-color: var(--border-glass);
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
</style>
