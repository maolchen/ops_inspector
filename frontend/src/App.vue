<template>
  <el-config-provider :locale="zhCn">
    <!-- 登录页面不显示侧边栏 -->
    <template v-if="$route.path === '/login'">
      <router-view />
    </template>
    
    <!-- 其他页面显示侧边栏 -->
    <template v-else>
      <el-container class="app-container">
        <el-aside width="220px" class="sidebar">
          <div class="logo">
            <h1>运维巡检平台</h1>
          </div>
          <el-menu
            :default-active="activeMenu"
            router
            background-color="#304156"
            text-color="#bfcbd9"
            active-text-color="#409EFF"
          >
            <el-menu-item index="/projects">
              <el-icon><Folder /></el-icon>
              <span>项目管理</span>
            </el-menu-item>
            <el-menu-item index="/rules">
              <el-icon><Setting /></el-icon>
              <span>规则配置</span>
            </el-menu-item>
            <el-menu-item index="/inspection">
              <el-icon><Search /></el-icon>
              <span>执行巡检</span>
            </el-menu-item>
            <el-menu-item index="/reports">
              <el-icon><Document /></el-icon>
              <span>历史报告</span>
            </el-menu-item>
            <el-menu-item index="/settings">
              <el-icon><Tools /></el-icon>
              <span>系统设置</span>
            </el-menu-item>
          </el-menu>
          
          <!-- 底部用户信息 -->
          <div class="user-panel">
            <el-dropdown trigger="click" @command="handleCommand">
              <div class="user-info">
                <el-avatar :size="32" icon="User" />
                <span class="username">{{ authStore.user?.username || '用户' }}</span>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="settings">
                    <el-icon><Setting /></el-icon>
                    个人设置
                  </el-dropdown-item>
                  <el-dropdown-item divided command="logout">
                    <el-icon><SwitchButton /></el-icon>
                    退出登录
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </el-aside>
        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </template>
  </el-config-provider>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { Folder, Setting, Search, Document, SwitchButton, Tools } from '@element-plus/icons-vue'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import { useAuthStore } from './store/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const activeMenu = computed(() => route.path)

async function handleCommand(command: string) {
  if (command === 'logout') {
    try {
      await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
      await authStore.logout()
      router.push('/login')
    } catch {
      // 用户取消
    }
  } else if (command === 'settings') {
    router.push('/profile')
  }
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

.app-container {
  height: 100vh;
}

/* 侧边栏 - Shadcn 深色风格 */
.sidebar {
  background: linear-gradient(180deg, hsl(240 5.9% 10%) 0%, hsl(240 6% 7%) 100%);
  overflow-x: hidden;
  display: flex;
  flex-direction: column;
  border-right: 1px solid hsl(240 5.9% 15%);
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: hsl(0 0% 98%);
  font-size: 18px;
  font-weight: 600;
  border-bottom: 1px solid hsl(240 5.9% 15%);
  letter-spacing: -0.025em;
}

.logo h1 {
  font-size: 0.9rem;
  font-weight: 600;
}

/* 主内容区域 */
.main-content {
  background: hsl(var(--muted));
  padding: 24px;
  overflow-y: auto;
}

/* 侧边栏菜单样式 */
.el-menu {
  border-right: none;
  flex: 1;
  background: transparent !important;
}

.el-menu-item {
  margin: 4px 8px;
  border-radius: calc(var(--radius) - 4px) !important;
  transition: all 0.2s ease !important;
}

.el-menu-item:hover {
  background-color: hsl(240 5.9% 20%) !important;
}

.el-menu-item.is-active {
  background-color: hsl(var(--primary) / 0.2) !important;
  color: hsl(0 0% 98%) !important;
}

/* 用户面板 */
.user-panel {
  padding: 16px;
  border-top: 1px solid hsl(240 5.9% 15%);
  background: transparent;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  color: hsl(240 5% 64.9%);
  cursor: pointer;
  padding: 10px;
  border-radius: calc(var(--radius) - 4px);
  transition: all 0.2s ease;
}

.user-info:hover {
  background-color: hsl(240 5.9% 15%);
  color: hsl(0 0% 98%);
}

.username {
  font-size: 0.875rem;
  font-weight: 500;
}

.el-dropdown-menu__item {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* 下拉菜单样式 */
.el-dropdown-menu {
  border-radius: var(--radius) !important;
  border: 1px solid hsl(var(--border)) !important;
  box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1) !important;
  padding: 4px !important;
}

.el-dropdown-menu__item {
  border-radius: calc(var(--radius) - 4px) !important;
  margin: 2px !important;
  padding: 8px 12px !important;
}

.el-dropdown-menu__item:hover {
  background-color: hsl(var(--muted)) !important;
  color: hsl(var(--foreground)) !important;
}
</style>
