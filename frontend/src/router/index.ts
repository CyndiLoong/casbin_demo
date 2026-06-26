import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'
import { useUserStore } from '@/store/user'

NProgress.configure({ showSpinner: false, trickleSpeed: 100 })

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录', requiresAuth: false }
  },
  {
    path: '/401',
    name: 'Error401',
    component: () => import('@/views/error/401.vue'),
    meta: { title: '未授权', requiresAuth: false }
  },
  {
    path: '/403',
    name: 'Error403',
    component: () => import('@/views/error/403.vue'),
    meta: { title: '访问被拒绝', requiresAuth: false }
  },
  {
    path: '/500',
    name: 'Error500',
    component: () => import('@/views/error/500.vue'),
    meta: { title: '服务器错误', requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/views/Layout.vue'),
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表盘' }
      },
      {
        path: 'apply',
        name: 'ApplyResource',
        component: () => import('@/views/ApplyResource.vue'),
        meta: { title: '申请资源' }
      },
      {
        path: 'my-applications',
        name: 'MyApplications',
        component: () => import('@/views/MyApplications.vue'),
        meta: { title: '我的申请' }
      },
      {
        path: 'resources',
        name: 'ResourceList',
        component: () => import('@/views/ResourceList.vue'),
        meta: { title: '资源清单' }
      },
      {
        path: 'messages',
        name: 'Messages',
        component: () => import('@/views/Messages.vue'),
        meta: { title: '消息通知' }
      },
      {
        path: 'audit',
        name: 'AuditList',
        component: () => import('@/views/AuditList.vue'),
        meta: { title: '资源审核', adminOnly: true }
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/Users.vue'),
        meta: { title: '用户管理', adminOnly: true }
      },
      {
        path: 'roles',
        name: 'Roles',
        component: () => import('@/views/Roles.vue'),
        meta: { title: '角色管理', adminOnly: true }
      },
      {
        path: 'permissions',
        name: 'Permissions',
        component: () => import('@/views/Permissions.vue'),
        meta: { title: '权限管理', adminOnly: true }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'Error404',
    component: () => import('@/views/error/404.vue'),
    meta: { title: '页面未找到', requiresAuth: false }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to, from, next) => {
  NProgress.start()
  const token = localStorage.getItem('token')
  const userStore = useUserStore()

  if (to.path === '/401' || to.path === '/403' || to.path === '/404' || to.path === '/500') {
    next()
    return
  }

  if (to.meta.requiresAuth !== false && !token) {
    if (to.path === '/login') {
      next()
    } else {
      next('/login')
    }
    return
  }

  if (to.path === '/login' && token) {
    next('/dashboard')
    return
  }

  if (to.meta.requiresAuth !== false && token) {
    if (!userStore.userInfo) {
      try {
        const info = await userStore.fetchUserInfo()
        if (!info) {
          userStore.logout()
          next('/login')
          return
        }
      } catch {
        userStore.logout()
        next('/login')
        return
      }
    }

    if (to.meta.adminOnly && !userStore.hasRole('admin')) {
      next('/403')
      return
    }
  }

  next()
})

router.afterEach((to) => {
  NProgress.done()
  const title = to.meta.title as string
  if (title) {
    document.title = `${title} - Casbin RBAC Demo`
  }
})

export default router
