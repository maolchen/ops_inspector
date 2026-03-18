<template>
  <div class="login-container">
    <!-- 背景图片 -->
    <div class="login-bg"></div>
    <!-- 背景遮罩 -->
    <div class="login-overlay"></div>
    
    <div class="login-box">
      <div class="login-header">
        <img src="/images/logo-guardian.png" alt="Logo" class="login-logo" />
        <h1 class="title">运维巡检平台</h1>
        <p class="subtitle">DevOps Inspection Platform</p>
      </div>
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
      <div class="login-footer">
        <span>安全稳定的运维监控解决方案</span>
      </div>
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
/* 登录页面 - 深色科技风格 */
.login-container {
  position: relative;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

/* 背景图片 */
.login-bg {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-image: url('/images/login-bg.png');
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  z-index: 0;
}

/* 深色遮罩 */
.login-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    135deg,
    rgba(10, 10, 20, 0.85) 0%,
    rgba(15, 23, 42, 0.9) 50%,
    rgba(10, 10, 20, 0.85) 100%
  );
  z-index: 1;
}

/* 登录框 */
.login-box {
  position: relative;
  z-index: 2;
  width: 420px;
  padding: 48px 40px;
  background: rgba(15, 23, 42, 0.95);
  backdrop-filter: blur(20px);
  border-radius: var(--radius);
  box-shadow: 
    0 25px 50px -12px rgba(0, 0, 0, 0.5),
    0 0 0 1px rgba(255, 255, 255, 0.05),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.08);
}

/* 登录头部 */
.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-logo {
  width: 64px;
  height: 64px;
  margin-bottom: 16px;
  border-radius: 50%;
  box-shadow: 0 0 30px rgba(34, 211, 238, 0.3);
  object-fit: cover;
}

.title {
  font-size: 1.75rem;
  font-weight: 600;
  color: hsl(0 0% 98%);
  margin-bottom: 8px;
  letter-spacing: -0.025em;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}

.subtitle {
  font-size: 0.875rem;
  color: hsl(215 20% 65%);
  font-weight: 400;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.login-form {
  margin-top: 24px;
}

/* 输入框样式覆盖 */
.login-form :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: none;
  border-radius: calc(var(--radius) - 4px);
  transition: all 0.2s ease;
}

.login-form :deep(.el-input__wrapper:hover) {
  border-color: rgba(255, 255, 255, 0.2);
}

.login-form :deep(.el-input__wrapper.is-focus) {
  border-color: hsl(187 92% 69%);
  box-shadow: 0 0 0 3px rgba(34, 211, 238, 0.1);
}

.login-form :deep(.el-input__inner) {
  color: hsl(0 0% 98%);
}

.login-form :deep(.el-input__inner::placeholder) {
  color: hsl(215 20% 50%);
}

.login-form :deep(.el-input__prefix) {
  color: hsl(215 20% 50%);
}

/* 登录按钮 */
.login-btn {
  width: 100%;
  height: 44px;
  font-weight: 500;
  border-radius: calc(var(--radius) - 4px);
  background: linear-gradient(135deg, hsl(187 92% 50%) 0%, hsl(199 89% 48%) 100%);
  border: none;
  box-shadow: 0 4px 14px rgba(34, 211, 238, 0.3);
  transition: all 0.2s ease;
}

.login-btn:hover {
  background: linear-gradient(135deg, hsl(187 92% 55%) 0%, hsl(199 89% 53%) 100%);
  box-shadow: 0 6px 20px rgba(34, 211, 238, 0.4);
  transform: translateY(-1px);
}

.login-btn:active {
  transform: translateY(0);
}

/* 底部提示 */
.login-footer {
  text-align: center;
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.login-footer span {
  font-size: 0.75rem;
  color: hsl(215 20% 45%);
  letter-spacing: 0.025em;
}
</style>
