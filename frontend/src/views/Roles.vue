<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>角色管理</h2>
        <p>管理系统角色及其权限分配</p>
      </div>
      <el-button type="primary" @click="openCreate">
        <el-icon><Plus /></el-icon> 新建角色
      </el-button>
    </div>

    <div class="glass-card table-card">
      <el-table :data="roles" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="角色标识" min-width="120">
          <template #default="{ row }">
            <code class="code-badge">{{ row.name }}</code>
          </template>
        </el-table-column>
        <el-table-column prop="label" label="角色名称" min-width="120" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="权限数" width="100">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.permissions?.length || 0 }} 个</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openAssign(row)">分配权限</el-button>
            <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-popconfirm title="确认删除该角色?" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button type="danger" link size="small">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑角色' : '新建角色'" width="500px" destroy-on-close>
      <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="角色标识" prop="name">
          <el-input v-model="form.name" :disabled="isEdit" placeholder="如: admin / editor" class="glass-input" />
        </el-form-item>
        <el-form-item label="角色名称" prop="label">
          <el-input v-model="form.label" placeholder="如: 管理员 / 编辑" class="glass-input" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="角色描述" class="glass-input" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ isEdit ? '保存' : '创建' }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="assignVisible" title="分配权限" width="600px" destroy-on-close>
      <p class="assign-hint">为角色 <code>{{ currentRole?.name }}</code> 分配权限：</p>
      <el-checkbox-group v-model="checkedPerms" class="perm-grid">
        <el-checkbox v-for="p in permissions" :key="p.id" :label="p.id">
          <span class="perm-label">{{ p.label }}</span>
          <span class="perm-method" :class="p.method.toLowerCase()">{{ p.method }}</span>
          <span class="perm-path">{{ p.path }}</span>
        </el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="assignVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAssign" :loading="assigning">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getRoleList, createRole, updateRole, deleteRole, assignPermission, getPermissionList, type Role, type Permission } from '@/api/rbac'

const loading = ref(false)
const roles = ref<Role[]>([])
const permissions = ref<Permission[]>([])
const dialogVisible = ref(false)
const assignVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const assigning = ref(false)
const formRef = ref()
const editId = ref<number>(0)
const currentRole = ref<Role | null>(null)
const checkedPerms = ref<number[]>([])

const form = reactive({ name: '', label: '', description: '', status: 1 })
const rules = {
  name: [{ required: true, message: '请输入角色标识', trigger: 'blur' }],
  label: [{ required: true, message: '请输入角色名称', trigger: 'blur' }]
}

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getRoleList()
    roles.value = res.data
  } catch (e) {} finally {
    loading.value = false
  }
}

const fetchPermissions = async () => {
  try {
    const res = await getPermissionList()
    permissions.value = res.data
  } catch (e) {}
}

const openCreate = () => {
  isEdit.value = false
  Object.assign(form, { name: '', label: '', description: '', status: 1 })
  dialogVisible.value = true
}

const openEdit = (row: Role) => {
  isEdit.value = true
  editId.value = row.id
  Object.assign(form, { name: row.name, label: row.label, description: row.description, status: row.status })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  await formRef.value?.validate()
  submitting.value = true
  try {
    if (isEdit.value) {
      await updateRole(editId.value, { ...form })
      ElMessage.success('更新成功')
    } else {
      await createRole({ ...form })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch (e) {} finally {
    submitting.value = false
  }
}

const handleDelete = async (id: number) => {
  await deleteRole(id)
  ElMessage.success('删除成功')
  fetchData()
}

const openAssign = async (row: Role) => {
  currentRole.value = row
  checkedPerms.value = (row.permissions || []).map(p => p.id)
  await fetchPermissions()
  assignVisible.value = true
}

const handleAssign = async () => {
  if (!currentRole.value) return
  assigning.value = true
  try {
    const existingIds = (currentRole.value.permissions || []).map(p => p.id)
    const toAdd = checkedPerms.value.filter(id => !existingIds.includes(id))
    for (const pid of toAdd) {
      await assignPermission(currentRole.value.id, pid)
    }
    ElMessage.success('权限分配成功')
    assignVisible.value = false
    fetchData()
  } catch (e) {} finally {
    assigning.value = false
  }
}

onMounted(() => {
  fetchData()
  fetchPermissions()
})
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
  background: rgba(99, 102, 241, 0.15);
  color: var(--primary-300);
  padding: 4px 10px;
  border-radius: 6px;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 13px;
}

.assign-hint {
  color: var(--text-secondary);
  margin-bottom: 16px;
}

.assign-hint code {
  background: rgba(99, 102, 241, 0.15);
  color: var(--primary-300);
  padding: 2px 8px;
  border-radius: 4px;
}

.perm-grid {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: 400px;
  overflow-y: auto;
  padding: 4px;
}

.perm-grid .el-checkbox {
  margin-right: 0 !important;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  background: var(--bg-glass);
  border-radius: 10px;
  border: 1px solid var(--border-glass);
  transition: all 0.2s;
}

.perm-grid .el-checkbox:hover {
  background: var(--bg-glass-hover);
  border-color: var(--primary-500);
}

.perm-grid .el-checkbox.is-checked {
  background: rgba(99, 102, 241, 0.1);
  border-color: var(--primary-500);
}

.perm-label { font-weight: 500; min-width: 80px; }
.perm-path { color: var(--text-muted); font-size: 13px; flex: 1; }

.perm-method {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
}

.perm-method.get { background: rgba(16, 185, 129, 0.15); color: #10b981; }
.perm-method.post { background: rgba(99, 102, 241, 0.15); color: var(--primary-400); }
.perm-method.put { background: rgba(245, 158, 11, 0.15); color: #f59e0b; }
.perm-method.delete { background: rgba(239, 68, 68, 0.15); color: #ef4444; }
</style>
