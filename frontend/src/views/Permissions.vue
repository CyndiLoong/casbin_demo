<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>权限管理</h2>
        <p>管理API访问权限节点</p>
      </div>
      <el-button type="primary" @click="openCreate">
        <el-icon><Plus /></el-icon> 新建权限
      </el-button>
    </div>

    <div class="glass-card table-card">
      <el-table :data="permissions" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="权限标识" min-width="140">
          <template #default="{ row }">
            <code class="code-badge">{{ row.name }}</code>
          </template>
        </el-table-column>
        <el-table-column prop="label" label="权限名称" min-width="120" />
        <el-table-column label="方法" width="100">
          <template #default="{ row }">
            <span class="method-badge" :class="row.method.toLowerCase()">{{ row.method }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="路径" min-width="200">
          <template #default="{ row }">
            <code class="path-badge">{{ row.path }}</code>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-popconfirm title="确认删除该权限?" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button type="danger" link size="small">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑权限' : '新建权限'" width="500px" destroy-on-close>
      <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="权限标识" prop="name">
          <el-input v-model="form.name" :disabled="isEdit" placeholder="如: user:list" class="glass-input" />
        </el-form-item>
        <el-form-item label="权限名称" prop="label">
          <el-input v-model="form.label" placeholder="如: 查看用户列表" class="glass-input" />
        </el-form-item>
        <el-form-item label="请求方法" prop="method">
          <el-select v-model="form.method" class="glass-input w-full" placeholder="选择方法">
            <el-option label="GET" value="GET" />
            <el-option label="POST" value="POST" />
            <el-option label="PUT" value="PUT" />
            <el-option label="DELETE" value="DELETE" />
            <el-option label="PATCH" value="PATCH" />
          </el-select>
        </el-form-item>
        <el-form-item label="API路径" prop="path">
          <el-input v-model="form.path" placeholder="如: /api/users" class="glass-input" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="权限描述" class="glass-input" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ isEdit ? '保存' : '创建' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getPermissionList, createPermission, updatePermission, deletePermission, type Permission } from '@/api/rbac'

const loading = ref(false)
const permissions = ref<Permission[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref()
const editId = ref<number>(0)

const form = reactive({ name: '', label: '', method: 'GET', path: '', description: '' })
const rules = {
  name: [{ required: true, message: '请输入权限标识', trigger: 'blur' }],
  label: [{ required: true, message: '请输入权限名称', trigger: 'blur' }],
  method: [{ required: true, message: '请选择请求方法', trigger: 'change' }],
  path: [{ required: true, message: '请输入API路径', trigger: 'blur' }]
}

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getPermissionList()
    permissions.value = res.data
  } catch (e) {} finally {
    loading.value = false
  }
}

const openCreate = () => {
  isEdit.value = false
  Object.assign(form, { name: '', label: '', method: 'GET', path: '', description: '' })
  dialogVisible.value = true
}

const openEdit = (row: Permission) => {
  isEdit.value = true
  editId.value = row.id
  Object.assign(form, { name: row.name, label: row.label, method: row.method, path: row.path, description: row.description })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  await formRef.value?.validate()
  submitting.value = true
  try {
    if (isEdit.value) {
      await updatePermission(editId.value, { ...form })
      ElMessage.success('更新成功')
    } else {
      await createPermission({ ...form })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch (e) {} finally {
    submitting.value = false
  }
}

const handleDelete = async (id: number) => {
  await deletePermission(id)
  ElMessage.success('删除成功')
  fetchData()
}

onMounted(fetchData)
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.page-header h2 { font-size: 24px; font-weight: 700; }
.page-header p { color: var(--text-muted); margin-top: 4px; font-size: 14px; }

.table-card { padding: 24px; }

.code-badge {
  background: rgba(168, 85, 247, 0.15);
  color: #c084fc;
  padding: 4px 10px;
  border-radius: 6px;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 13px;
}

.path-badge {
  background: rgba(6, 182, 212, 0.1);
  color: var(--accent-cyan);
  padding: 4px 10px;
  border-radius: 6px;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 13px;
}

.method-badge {
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
}

.method-badge.get { background: rgba(16, 185, 129, 0.15); color: #10b981; }
.method-badge.post { background: rgba(99, 102, 241, 0.15); color: var(--primary-400); }
.method-badge.put { background: rgba(245, 158, 11, 0.15); color: #f59e0b; }
.method-badge.delete { background: rgba(239, 68, 68, 0.15); color: #ef4444; }
.method-badge.patch { background: rgba(168, 85, 247, 0.15); color: #c084fc; }

.w-full { width: 100%; }
</style>
