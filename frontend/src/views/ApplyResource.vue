<template>
  <div class="apply-page animate-fade-in-up">
    <div class="page-header">
      <div>
        <h2 class="page-title-text">大模型API资源申请</h2>
        <p class="page-desc">填写申请表单，提交后将由管理员审核，审核结果将实时通知</p>
      </div>
    </div>

    <div class="form-container glass-card">
      <el-form :model="form" :rules="rules" ref="formRef" label-position="top" class="audit-form">
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12">
            <el-form-item label="资源名称" prop="resource_name">
              <el-input v-model="form.resource_name" placeholder="例如：GPT-4 API" class="glass-input" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item label="资源类型" prop="resource_type">
              <el-select v-model="form.resource_type" placeholder="请选择资源类型" class="glass-input" style="width: 100%">
                <el-option label="对话大模型 (LLM)" value="llm_chat" />
                <el-option label="代码大模型 (Code)" value="llm_code" />
                <el-option label="图像生成 (Image Gen)" value="image_gen" />
                <el-option label="语音识别 (ASR)" value="asr" />
                <el-option label="语音合成 (TTS)" value="tts" />
                <el-option label="向量嵌入 (Embedding)" value="embedding" />
                <el-option label="其他" value="other" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :xs="24" :sm="12">
            <el-form-item label="API名称" prop="api_name">
              <el-input v-model="form.api_name" placeholder="例如：gpt-4-turbo-preview" class="glass-input" />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12">
            <el-form-item label="预期QPS（每秒请求数）" prop="expected_qps">
              <el-input-number v-model="form.expected_qps" :min="0" :max="10000" placeholder="0" class="glass-input" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="API描述（选填）" prop="api_description">
          <el-input
            v-model="form.api_description"
            type="textarea"
            :rows="3"
            placeholder="请简要描述该API的功能、版本等信息"
            class="glass-input"
          />
        </el-form-item>

        <el-form-item label="使用目的" prop="purpose">
          <el-input
            v-model="form.purpose"
            type="textarea"
            :rows="4"
            placeholder="请详细描述使用场景、业务需求、预计调用量等信息，这将帮助管理员更快审批"
            class="glass-input"
          />
        </el-form-item>

        <el-form-item label="联系方式（选填）" prop="contact_info">
          <el-input v-model="form.contact_info" placeholder="手机号/邮箱/企业微信，便于审核人员沟通" class="glass-input" />
        </el-form-item>

        <el-form-item>
          <div class="form-actions">
            <button class="btn-gradient" @click="handleSubmit" :disabled="submitting">
              <el-icon v-if="submitting" class="is-loading"><Loading /></el-icon>
              {{ submitting ? '提交中...' : '提交审核' }}
            </button>
            <button class="btn-reset" @click="handleReset" :disabled="submitting">重置表单</button>
          </div>
        </el-form-item>
      </el-form>
    </div>

    <div v-if="recentApps.length > 0" class="recent-section">
      <h3 class="section-title">最近提交</h3>
      <div class="recent-list">
        <div v-for="app in recentApps" :key="app.id" class="recent-item glass-card">
          <div class="recent-info">
            <span class="recent-name">{{ app.resource_name }}</span>
            <span class="recent-api">{{ app.api_name }}</span>
          </div>
          <el-tag :type="getStatusType(app.status)" size="small" effect="dark">{{ app.status_text }}</el-tag>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Loading } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { submitAudit, getMyApplications, type AuditApplication } from '@/api/audit'

const router = useRouter()
const formRef = ref<FormInstance>()
const submitting = ref(false)
const recentApps = ref<AuditApplication[]>([])

const form = reactive({
  resource_name: '',
  resource_type: '',
  api_name: '',
  api_description: '',
  purpose: '',
  expected_qps: 10,
  contact_info: ''
})

const rules: FormRules = {
  resource_name: [{ required: true, message: '请输入资源名称', trigger: 'blur' }],
  resource_type: [{ required: true, message: '请选择资源类型', trigger: 'change' }],
  api_name: [{ required: true, message: '请输入API名称', trigger: 'blur' }],
  purpose: [{ required: true, message: '请描述使用目的', trigger: 'blur', min: 5 }],
  expected_qps: [{ required: true, message: '请输入预期QPS', trigger: 'blur' }]
}

const getStatusType = (status: number) => {
  switch (status) {
    case 0: return 'warning'
    case 1: return 'success'
    case 2: return 'danger'
    default: return 'info'
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      await submitAudit({ ...form })
      ElMessage.success('申请已提交，请等待管理员审核')
      handleReset()
      loadRecent()
    } catch (e) {
      console.error(e)
    } finally {
      submitting.value = false
    }
  })
}

const handleReset = () => {
  formRef.value?.resetFields()
  form.resource_name = ''
  form.resource_type = ''
  form.api_name = ''
  form.api_description = ''
  form.purpose = ''
  form.expected_qps = 10
  form.contact_info = ''
}

const loadRecent = async () => {
  try {
    const res = await getMyApplications({ page: 1, page_size: 5 })
    recentApps.value = res.data.list
  } catch (e) {
    console.error(e)
  }
}

onMounted(() => {
  loadRecent()
})
</script>

<style scoped>
.apply-page {
  max-width: 800px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 24px;
}

.page-title-text {
  font-size: 24px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--text-primary), var(--primary-300));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin-bottom: 8px;
}

.page-desc {
  color: var(--text-secondary);
  font-size: 14px;
}

.form-container {
  padding: 32px;
}

.audit-form {
  margin-top: -8px;
}

.audit-form :deep(.el-form-item) {
  margin-bottom: 20px;
}

.form-actions {
  display: flex;
  gap: 12px;
}

.btn-reset {
  padding: 10px 24px;
  background: var(--bg-glass);
  border: 1px solid var(--border-glass);
  border-radius: 12px;
  color: var(--text-secondary);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn-reset:hover {
  background: var(--bg-glass-hover);
  color: var(--text-primary);
}

.recent-section {
  margin-top: 32px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
  color: var(--text-primary);
}

.recent-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.recent-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
}

.recent-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.recent-name {
  font-size: 14px;
  font-weight: 500;
}

.recent-api {
  font-size: 12px;
  color: var(--text-muted);
}

:deep(.el-input__wrapper) {
  background: rgba(255,255,255,0.03) !important;
  border-radius: 10px !important;
  box-shadow: 0 0 0 1px var(--border-glass) inset !important;
}

:deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px rgba(99,102,241,0.3) inset !important;
}

:deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--primary-500) inset, 0 0 0 3px rgba(99,102,241,0.1) !important;
}

:deep(.el-textarea__inner) {
  background: rgba(255,255,255,0.03) !important;
  border: 1px solid var(--border-glass) !important;
  border-radius: 10px !important;
  color: var(--text-primary) !important;
}

:deep(.el-textarea__inner:focus) {
  border-color: var(--primary-500) !important;
  box-shadow: 0 0 0 3px rgba(99,102,241,0.1) !important;
}

:deep(.el-select .el-input__wrapper) {
  background: rgba(255,255,255,0.03) !important;
}

:deep(.el-input-number .el-input__wrapper) {
  background: rgba(255,255,255,0.03) !important;
}
</style>
