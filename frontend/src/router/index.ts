import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../store/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/login/index.vue'),
    meta: { title: '登录', public: true }
  },
  {
    path: '/',
    redirect: '/projects'
  },
  {
    path: '/projects',
    name: 'Projects',
    component: () => import('../views/projects/index.vue'),
    meta: { title: '项目管理' }
  },
  {
    path: '/rules',
    name: 'Rules',
    component: () => import('../views/rules/index.vue'),
    meta: { title: '规则配置' }
  },
  {
    path: '/inspection',
    name: 'Inspection',
    component: () => import('../views/inspection/index.vue'),
    meta: { title: '执行巡检' }
  },
  {
    path: '/reports',
    name: 'Reports',
    component: () => import('../views/reports/index.vue'),
    meta: { title: '历史报告' }
  },
  {
    path: '/reports/:id',
    name: 'ReportDetail',
    component: () => import('../views/reports/detail.vue'),
    meta: { title: '报告详情' }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('../views/settings/index.vue'),
    meta: { title: '个人设置' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, _from, next) => {
  document.title = (to.meta?.title as string) || '运维巡检平台'

  const authStore = useAuthStore()
  const isPublic = to.meta?.public

  // 如果不是公开页面且未登录，跳转到登录页
  if (!isPublic && !authStore.isLoggedIn) {
    next('/login')
    return
  }

  // 如果已登录且访问登录页，跳转到首页
  if (to.path === '/login' && authStore.isLoggedIn) {
    next('/')
    return
  }

  next()
})

export default router
