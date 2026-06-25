<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h2>用户管理</h2>
        <p>管理系统用户及其角色分配</p>
      </div>
      <el-button type="primary" @click="openCreate">
        <el-icon><Plus /></el-icon> 新建用户
      </el-button>
    </div>

    <div class="glass-card table-card">
      <el-table :data="users" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="username" label="用户名" min-width="120">
          <template #default="{ row }">
            <div class="user-cell">
              <div class="user-avatar-sm">
                <el-icon><UserFilled /></el-icon>
              </div>
              <span>{{ row.username }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="nickname" label="昵称" min-width="120" />
        <el-table-column prop="email" label="邮箱" min-width="180" />
        <el-table-column label="角色" min-width="160">
          <template #default="{ row }">
            <el-tag v-for="r in row.roles" :key="r" size="small" type="primary" effect="plain" style="margin-right: 4px">
              {{ r }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="170">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEdit(row)">编辑</el-button>
            <el-popconfirm title="确认删除该用户?" @confirm="handleDelete(row.id)">
              <template #reference>
                <el-button type="danger" link size="small">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="fetchData"
          @current-change="fetchData"
        />
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑用户' : '新建用户'" width="500px" destroy-on-close>
      <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" :disabled="isEdit" placeholder="请输入用户名" class="glass-input" />
        </el-form-item>
        <el-form-item v-if="!isEdit" label="密码" prop="password">
          <el-input v-model="form.password" type="password" show-password placeholder="请输入密码" class="glass-input" />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="form.nickname" placeholder="请输入昵称" class="glass-input" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱" class="glass-input" />
        </el-form-item>
        <el-form-item v-if="isEdit" label="状态" prop="status">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="正常" inactive-text="禁用" />
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
import { Plus, UserFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getUserList, createUser, updateUser, deleteUser, type User } from '@/api/user'

const loading = ref(false)
const users = ref<User[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref()
const editId = ref<number>(0)

const form = reactive({
  username: '',
  password: '',
  nickname: '',
  email: '',
  status: 1
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const formatDate = (d: string) => new Date(d).toLocaleString('zh-CN')

const fetchData = async () => {
  loading.value = true
  try {
    const res = await getUserList(page.value, pageSize.value)
    users.value = res.data.list
    total.value = res.data.total
  } catch (e) {} finally {
    loading.value = false
  }
}

const openCreate = () => {
  isEdit.value = false
  Object.assign(form, { username: '', password: '', nickname: '', email: '', status: 1 })
  dialogVisible.value = true
}

const openEdit = (row: User) => {
  isEdit.value = true
  editId.value = row.id
  Object.assign(form, {
    username: row.username,
    password: '',
    nickname: row.nickname,
    email: row.email,
    status: row.status
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  await formRef.value?.validate()
  submitting.value = true
  try {
    if (isEdit.value) {
      await updateUser(editId.value, { nickname: form.nickname, email: form.email, status: form.status })
      ElMessage.success('更新成功')
    } else {
      await createUser({ username: form.username, password: form.password, nickname: form.nickname, email: form.email })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch (e) {} finally {
    submitting.value = false
  }
}

const handleDelete = async (id: number) => {
  await deleteUser(id)
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

.page-header h2 {
  font-size: 24px;
  font-weight: 700;
}

.page-header p {
  color: var(--text-muted);
  margin-top: 4px;
  font-size: 14px;
}

.table-card {
  padding: 24px;
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-avatar-sm {
  width: 32px;
  height: 32px;
  background: linear-gradient(135deg, var(--primary-500), var(--accent-purple));
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 14px;
}

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}
</style>
