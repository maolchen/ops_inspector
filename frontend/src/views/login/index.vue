<template>
  <div class="login-container">
    <div class="login-box">
      <h1 class="title">运维巡检平台</h1>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        class="login-form"
        @keyup.enter="handleLogin"
      >
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="用户名"
            size="large"
            :prefix-icon="User"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            size="large"
            :prefix-icon="Lock"
            show-password
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            class="login-btn"
            :loading="loading"
            @click="handleLogin"
          >
            登 录
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { useAuthStore } from '../../store/auth'

const router = useRouter()
const authStore = useAuthStore()
const formRef = ref()
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

async function handleLogin() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await authStore.login(form.username, form.password)
    ElMessage.success('登录成功')
    router.push('/')
  } catch (error: any) {
    // 错误已在 api 拦截器中处理
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
/* 登录页面 - Shadcn 风格 */
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, hsl(var(--primary) / 0.1) 0%, hsl(var(--muted)) 100%);
}

.login-box {
  width: 400px;
  padding: 40px;
  background: hsl(var(--background));
  border-radius: var(--radius);
  box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
  border: 1px solid hsl(var(--border));
}

.title {
  text-align: center;
  font-size: 1.5rem;
  font-weight: 600;
  color: hsl(var(--foreground));
  margin-bottom: 30px;
  letter-spacing: -0.025em;
}

.login-form {
  margin-top: 20px;
}

.login-btn {
  width: 100%;
}
</style>
