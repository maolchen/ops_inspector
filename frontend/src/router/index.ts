import { createRouter, createWebHistory } from 'vue-router'

const routes = [
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
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, _from, next) => {
  document.title = (to.meta?.title as string) || '运维巡检平台'
  next()
})

export default router
