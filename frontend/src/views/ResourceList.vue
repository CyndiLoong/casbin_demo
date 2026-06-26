<template>
  <div class="resource-list" :class="{ 'admin-view': isAdmin }">
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">{{ isAdmin ? '资源管理' : '资源清单' }}</h2>
        <p class="page-desc">
          {{ isAdmin ? '管理所有大模型API资源，支持新增、编辑、删除操作' : '浏览所有可用的大模型API资源，选择适合您业务的资源进行申请' }}
        </p>
      </div>
      <div class="header-right" v-if="isAdmin">
        <button class="btn-gradient admin-gradient" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          新增资源
        </button>
      </div>
    </div>

    <div class="filter-bar glass-card">
      <div class="filter-item">
        <el-input
          v-model="keyword"
          placeholder="搜索资源名称、API名称..."
          clearable
          :prefix-icon="Search"
          @keyup.enter="handleSearch"
          @clear="handleSearch"
          class="search-input"
        />
      </div>
      <div class="filter-item">
        <el-select v-model="filterType" placeholder="资源类型" clearable @change="handleSearch" class="filter-select">
          <el-option label="对话大模型" value="llm_chat" />
          <el-option label="代码大模型" value="llm_code" />
          <el-option label="图像生成" value="image_gen" />
          <el-option label="语音识别" value="asr" />
          <el-option label="语音合成" value="tts" />
          <el-option label="向量嵌入" value="embedding" />
          <el-option label="其他" value="other" />
        </el-select>
      </div>
      <div class="filter-item" v-if="isAdmin">
        <el-select v-model="filterStatus" placeholder="状态" clearable @change="handleSearch" class="filter-select">
          <el-option label="可用" :value="1" />
          <el-option label="不可用" :value="0" />
        </el-select>
      </div>
    </div>

    <div class="resource-grid">
      <div
        v-for="resource in resources"
        :key="resource.id"
        class="resource-card glass-card"
        :class="{ inactive: resource.status !== 1 }"
      >
        <div class="card-header">
          <div class="resource-icon" :class="getTypeClass(resource.type)">
            <el-icon :size="24">
              <component :is="getTypeIcon(resource.type)" />
            </el-icon>
          </div>
          <el-tag :type="resource.status === 1 ? 'success' : 'info'" size="small" effect="dark">
            {{ resource.status === 1 ? '可用' : '暂不可用' }}
          </el-tag>
        </div>

        <h3 class="resource-name">{{ resource.name }}</h3>
        <p class="resource-api">{{ resource.api_name }}</p>
        <p class="resource-desc">{{ resource.description }}</p>

        <div class="resource-meta">
          <div class="meta-item">
            <el-icon><OfficeBuilding /></el-icon>
            <span>{{ resource.provider }}</span>
          </div>
          <div class="meta-item">
            <el-icon><Tickets /></el-icon>
            <span>{{ resource.version }}</span>
          </div>
        </div>

        <div class="resource-qps">
          <div class="qps-info">
            <span class="qps-label">默认QPS</span>
            <span class="qps-value">{{ resource.default_qps }}</span>
          </div>
          <div class="qps-info">
            <span class="qps-label">最大QPS</span>
            <span class="qps-value">{{ resource.max_qps }}</span>
          </div>
        </div>

        <div class="card-tags" v-if="resource.tags">
          <el-tag
            v-for="(tag, idx) in parseTags(resource.tags)"
            :key="idx"
            size="small"
            type="info"
            effect="plain"
          >
            {{ tag }}
          </el-tag>
        </div>

        <div class="card-actions">
          <template v-if="isAdmin">
            <button class="action-btn secondary" @click="handleEdit(resource)">
              <el-icon><Edit /></el-icon>
              编辑
            </button>
            <button class="action-btn danger" @click="handleDelete(resource)">
              <el-icon><Delete /></el-icon>
              删除
            </button>
          </template>
          <template v-else>
            <a
              v-if="resource.docs_url"
              :href="resource.docs_url"
              target="_blank"
              class="action-btn secondary"
            >
              <el-icon><Document /></el-icon>
              文档
            </a>
            <button
              class="action-btn primary"
              :disabled="resource.status !== 1"
              @click="handleApply(resource)"
            >
              <el-icon><Promotion /></el-icon>
              申请使用
            </button>
          </template>
        </div>
      </div>
    </div>

    <div class="pagination-wrapper" v-if="total > pageSize">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[8, 16, 24, 40]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
        background
      />
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑资源' : '新增资源'"
      width="600px"
      :close-on-click-modal="false"
      class="resource-dialog"
    >
      <el-form :model="formData" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item label="资源名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入资源名称" />
        </el-form-item>
        <el-form-item label="资源类型" prop="type">
          <el-select v-model="formData.type" placeholder="请选择资源类型" style="width: 100%">
            <el-option label="对话大模型" value="llm_chat" />
            <el-option label="代码大模型" value="llm_code" />
            <el-option label="图像生成" value="image_gen" />
            <el-option label="语音识别" value="asr" />
            <el-option label="语音合成" value="tts" />
            <el-option label="向量嵌入" value="embedding" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="API名称" prop="api_name">
          <el-input v-model="formData.api_name" placeholder="请输入API名称" />
        </el-form-item>
        <el-form-item label="提供厂商" prop="provider">
          <el-input v-model="formData.provider" placeholder="请输入提供厂商" />
        </el-form-item>
        <el-form-item label="版本" prop="version">
          <el-input v-model="formData.version" placeholder="请输入版本号" />
        </el-form-item>
        <el-form-item label="资源描述" prop="description">
          <el-input v-model="formData.description" type="textarea" :rows="3" placeholder="请输入资源描述" />
        </el-form-item>
        <el-form-item label="默认QPS" prop="default_qps">
          <el-input-number v-model="formData.default_qps" :min="1" :max="1000" />
        </el-form-item>
        <el-form-item label="最大QPS" prop="max_qps">
          <el-input-number v-model="formData.max_qps" :min="1" :max="10000" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="formData.status">
            <el-radio :value="1">可用</el-radio>
            <el-radio :value="0">不可用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="文档链接" prop="docs_url">
          <el-input v-model="formData.docs_url" placeholder="请输入文档链接" />
        </el-form-item>
        <el-form-item label="标签" prop="tags">
          <el-input v-model="formData.tags" placeholder='请输入标签，JSON数组格式，如: ["标签1","标签2"]' />
        </el-form-item>
      </el-form>
      <template #footer>
        <button class="btn-glass" @click="dialogVisible = false">取消</button>
        <button class="btn-gradient" @click="handleSubmit" :disabled="loading">
          <span v-if="loading">
            <el-icon class="is-loading"><Loading /></el-icon>
            保存中...
          </span>
          <span v-else>确定</span>
        </button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  Search, Plus, Edit, Delete, Document, Promotion,
  OfficeBuilding, Tickets, Loading,
  ChatDotSquare, Monitor, Picture, Microphone, Headset, MagicStick, SetUp
} from '@element-plus/icons-vue'
import { useUserStore } from '@/store/user'
import { getResources, createResource, updateResource, deleteResource, type Resource } from '@/api/resource'

const router = useRouter()
const userStore = useUserStore()
const isAdmin = computed(() => userStore.hasRole('admin'))

const resources = ref<Resource[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(8)
const keyword = ref('')
const filterType = ref('')
const filterStatus = ref<number | ''>('')
const loading = ref(false)

const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref<FormInstance>()
const formData = reactive({
  name: '',
  type: '',
  api_name: '',
  provider: '',
  version: '',
  description: '',
  default_qps: 10,
  max_qps: 100,
  status: 1,
  docs_url: '',
  tags: ''
})

const formRules: FormRules = {
  name: [{ required: true, message: '请输入资源名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择资源类型', trigger: 'change' }],
  api_name: [{ required: true, message: '请输入API名称', trigger: 'blur' }],
}

const loadResources = async () => {
  loading.value = true
  try {
    const params: any = {
      page: page.value,
      page_size: pageSize.value
    }
    if (keyword.value) params.keyword = keyword.value
    if (filterType.value) params.type = filterType.value
    if (filterStatus.value !== '') params.status = filterStatus.value
    
    const res = await getResources(params)
    resources.value = res.data.list
    total.value = res.data.total
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  page.value = 1
  loadResources()
}

const handlePageChange = (p: number) => {
  page.value = p
  loadResources()
}

const handleSizeChange = (size: number) => {
  pageSize.value = size
  page.value = 1
  loadResources()
}

const getTypeClass = (type: string) => {
  const map: Record<string, string> = {
    llm_chat: 'type-chat',
    llm_code: 'type-code',
    image_gen: 'type-image',
    asr: 'type-asr',
    tts: 'type-tts',
    embedding: 'type-embedding',
    other: 'type-other'
  }
  return map[type] || 'type-other'
}

const getTypeIcon = (type: string) => {
  const map: Record<string, any> = {
    llm_chat: ChatDotSquare,
    llm_code: Monitor,
    image_gen: Picture,
    asr: Microphone,
    tts: Headset,
    embedding: MagicStick,
    other: SetUp
  }
  return map[type] || SetUp
}

const parseTags = (tagsStr: string) => {
  try {
    return JSON.parse(tagsStr)
  } catch {
    return []
  }
}

const handleCreate = () => {
  isEdit.value = false
  Object.assign(formData, {
    name: '',
    type: '',
    api_name: '',
    provider: '',
    version: '',
    description: '',
    default_qps: 10,
    max_qps: 100,
    status: 1,
    docs_url: '',
    tags: ''
  })
  dialogVisible.value = true
}

const handleEdit = (resource: Resource) => {
  isEdit.value = true
  Object.assign(formData, {
    id: resource.id,
    name: resource.name,
    type: resource.type,
    api_name: resource.api_name,
    provider: resource.provider,
    version: resource.version,
    description: resource.description,
    default_qps: resource.default_qps,
    max_qps: resource.max_qps,
    status: resource.status,
    docs_url: resource.docs_url,
    tags: resource.tags
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  await formRef.value?.validate()
  loading.value = true
  try {
    if (isEdit.value) {
      await updateResource((formData as any).id, formData)
      ElMessage.success('更新成功')
    } else {
      await createResource(formData)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadResources()
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const handleDelete = async (resource: Resource) => {
  try {
    await ElMessageBox.confirm(`确定要删除资源「${resource.name}」吗？`, '确认删除', {
      type: 'warning'
    })
    await deleteResource(resource.id)
    ElMessage.success('删除成功')
    loadResources()
  } catch (e: any) {
    if (e !== 'cancel') {
      console.error(e)
    }
  }
}

const handleApply = (resource: Resource) => {
  router.push({
    path: '/apply',
    query: {
      resource_name: resource.name,
      resource_type: resource.type,
      api_name: resource.api_name
    }
  })
}

onMounted(() => {
  loadResources()
})
</script>

<style scoped>
.resource-list {
  max-width: 1400px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.header-left h2 {
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 6px 0;
}

.page-desc {
  color: var(--text-secondary);
  font-size: 14px;
  margin: 0;
}

.filter-bar {
  display: flex;
  gap: 16px;
  padding: 16px 20px;
  margin-bottom: 20px;
  align-items: center;
  flex-wrap: wrap;
}

.filter-item {
  flex: 1;
  min-width: 200px;
}

.search-input {
  width: 100%;
}

.filter-select {
  width: 100%;
}

.resource-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.resource-card {
  padding: 24px;
  display: flex;
  flex-direction: column;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}

.resource-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 40px rgba(99, 102, 241, 0.15);
  border-color: rgba(99, 102, 241, 0.3);
}

.resource-card.inactive {
  opacity: 0.6;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.resource-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.type-chat {
  background: linear-gradient(135deg, #6366f1, #818cf8);
}

.type-code {
  background: linear-gradient(135deg, #06b6d4, #22d3ee);
}

.type-image {
  background: linear-gradient(135deg, #a855f7, #c084fc);
}

.type-asr {
  background: linear-gradient(135deg, #f59e0b, #fbbf24);
}

.type-tts {
  background: linear-gradient(135deg, #10b981, #34d399);
}

.type-embedding {
  background: linear-gradient(135deg, #ec4899, #f472b6);
}

.type-other {
  background: linear-gradient(135deg, #6b7280, #9ca3af);
}

.resource-name {
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 4px 0;
}

.resource-api {
  font-size: 13px;
  color: var(--primary-400);
  margin: 0 0 12px 0;
  font-family: 'Courier New', monospace;
}

.resource-desc {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.6;
  margin: 0 0 16px 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 41px;
}

.resource-meta {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-muted);
}

.resource-qps {
  display: flex;
  gap: 16px;
  padding: 12px 0;
  border-top: 1px solid var(--border-glass);
  border-bottom: 1px solid var(--border-glass);
  margin-bottom: 16px;
}

.qps-info {
  flex: 1;
  text-align: center;
}

.qps-label {
  display: block;
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 4px;
}

.qps-value {
  font-size: 20px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--primary-400), var(--accent-cyan));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.card-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 16px;
}

.card-actions {
  display: flex;
  gap: 8px;
  margin-top: auto;
}

.action-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 16px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
  text-decoration: none;
}

.action-btn.primary {
  background: linear-gradient(135deg, var(--primary-600), var(--primary-500));
  color: white;
}

.action-btn.primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.4);
}

.action-btn.primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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

.action-btn.danger {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: var(--danger);
}

.action-btn.danger:hover {
  background: rgba(239, 68, 68, 0.2);
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  padding: 16px 0;
}

:deep(.el-dialog) {
  background: var(--bg-glass) !important;
  backdrop-filter: blur(20px);
  border: 1px solid var(--border-glass) !important;
}

:deep(.el-dialog__title) {
  color: var(--text-primary) !important;
}

:deep(.el-dialog__headerbtn .el-dialog__close) {
  color: var(--text-secondary) !important;
}

:deep(.el-form-item__label) {
  color: var(--text-secondary) !important;
}

.btn-gradient.admin-gradient {
  background: linear-gradient(135deg, #f59e0b, #f97316);
}

.btn-gradient.admin-gradient:hover {
  box-shadow: 0 6px 20px rgba(245, 158, 11, 0.4);
}

.resource-list.admin-view .resource-card:hover {
  box-shadow: 0 12px 40px rgba(245, 158, 11, 0.15);
  border-color: rgba(245, 158, 11, 0.3);
}

.resource-list.admin-view .qps-value {
  background: linear-gradient(135deg, #f59e0b, #f97316);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.resource-list.admin-view .action-btn.primary {
  background: linear-gradient(135deg, #f59e0b, #f97316);
}

.resource-list.admin-view .action-btn.primary:hover:not(:disabled) {
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.4);
}

.resource-list.admin-view .action-btn.secondary:hover {
  border-color: #f59e0b;
}

@media (max-width: 768px) {
  .resource-grid {
    grid-template-columns: 1fr;
  }
  
  .page-header {
    flex-direction: column;
    gap: 12px;
  }
  
  .filter-bar {
    flex-direction: column;
  }
  
  .filter-item {
    min-width: 100%;
  }
}
</style>
