<template>
  <div class="page-container">
    <div class="page-header">
      <h2>个人设置</h2>
    </div>

    <el-row :gutter="20">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>修改密码</span>
          </template>

          <el-form ref="formRef" :model="passwordForm" :rules="rules" label-width="100px">
            <el-form-item label="当前密码" prop="oldPassword">
              <el-input
                v-model="passwordForm.oldPassword"
                type="password"
                placeholder="请输入当前密码"
                show-password
              />
            </el-form-item>
            <el-form-item label="新密码" prop="newPassword">
              <el-input
                v-model="passwordForm.newPassword"
                type="password"
                placeholder="请输入新密码（至少4位）"
                show-password
              />
            </el-form-item>
            <el-form-item label="确认密码" prop="confirmPassword">
              <el-input
                v-model="passwordForm.confirmPassword"
                type="password"
                placeholder="请再次输入新密码"
                show-password
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleChangePassword" :loading="loading">
                修改密码
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card>
          <template #header>
            <span>账户信息</span>
          </template>

          <el-descriptions :column="1" border>
            <el-descriptions-item label="用户ID">
              {{ authStore.user?.id }}
            </el-descriptions-item>
            <el-descriptions-item label="用户名">
              {{ authStore.user?.username }}
            </el-descriptions-item>
            <el-descriptions-item label="显示名称">
              {{ authStore.user?.display_name || '-' }}
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useAuthStore } from '../../store/auth'
import api from '../../api/index'

const authStore = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const validateConfirmPassword = (_rule: any, value: string, callback: any) => {
  if (value !== passwordForm.newPassword) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  oldPassword: [
    { required: true, message: '请输入当前密码', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 4, message: '密码长度至少4位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

const handleChangePassword = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    loading.value = true
    try {
      await api.put('/auth/password', {
        old_password: passwordForm.oldPassword,
        new_password: passwordForm.newPassword
      })
      ElMessage.success('密码修改成功，请重新登录')
      
      // 清空表单
      passwordForm.oldPassword = ''
      passwordForm.newPassword = ''
      passwordForm.confirmPassword = ''
      
      // 退出登录
      setTimeout(async () => {
        await authStore.logout()
        window.location.href = '/login'
      }, 1500)
    } catch (error: any) {
      const msg = error.response?.data?.error || '修改失败'
      ElMessage.error(msg)
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
/* 页面样式由全局 shadcn.css 统一管理 */
</style>
