import { createRouter, createWebHashHistory } from 'vue-router'
import HomeLayout from '../views/Home/layout/index.vue'
import SSHLayout from '../views/SSH/layout/index.vue'

const routes = [
  {
    path: '/',
    redirect: '/home'
  },
  {
    path: '/home',
    component: HomeLayout,
    redirect: '/home/new',
    meta: { title: 'Home' },
    children: [
      {
        path: 'new',
        name: 'NewConnection',
        component: () => import('../views/Home/NewConnection.vue'),
        meta: { title: 'New Connection' }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('../views/Home/Settings.vue'),
        meta: { title: 'Settings' }
      }
    ]
  },
  {
    path: '/ssh',
    component: SSHLayout,
    meta: { title: 'SSH' }
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

router.beforeEach((to, from) => {
  const title = to.meta.title
  if (title) {
    document.title = `${title} - Qssh`
  } else {
    document.title = 'Qssh'
  }
  return true
})

export default router
