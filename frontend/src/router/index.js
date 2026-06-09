import { createRouter, createWebHashHistory } from 'vue-router'
import HomeLayout from '../views/Home/layout/index.vue'
import SSHLayout from '../views/SSH/layout/index.vue'

const routes = [
  {
    path: '/',
    redirect: '/home'
  },
  // Home 模块路由（使用 Home 自己的布局）
  {
    path: '/home',
    component: HomeLayout,
    redirect: '/home/new',
    meta: { title: '首页' },
    children: [
      {
        path: 'new',
        name: 'NewConnection',
        component: () => import('../views/Home/NewConnection.vue'),
        meta: { title: '新建连接' }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('../views/Home/Settings.vue'),
        meta: { title: '设置' }
      }
      // 后续可以添加更多 Home 子路由
    ]
  },
  // SSH 模块路由（使用 SSH 自己的布局）
  {
    path: '/ssh',
    component: SSHLayout,
    redirect: '/ssh/terminal',
    meta: { title: 'SSH终端' },
    children: [
      {
        path: 'terminal',
        name: 'SSHTerminal',
        component: () => import('../views/SSH/Terminal.vue'),
        meta: { title: '终端' }
      }
      // 后续可以添加文件管理、性能监控等子路由
    ]
  }
  // {
  //   path: '/files',
  //   component: FileManager,  // 文件管理页面使用自己的 FileLayout
  //   redirect: '/files/browser',
  //   meta: { title: '文件管理' },
  //   children: [
  //     {
  //       path: 'browser',
  //       name: 'FileBrowser',
  //       component: () => import('../views/FileManager/Browser.vue'),
  //       meta: { title: '文件浏览' }
  //     }
  //   ]
  // }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

// 路由守卫 - 设置页面标题
router.beforeEach((to, from) => {
  const title = to.meta.title
  if (title) {
    document.title = `${title} - 启SSH`
  } else {
    document.title = '启SSH - SSH工具'
  }
  return true
})

export default router
