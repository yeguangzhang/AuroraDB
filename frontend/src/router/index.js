import { createRouter, createWebHashHistory } from 'vue-router'
import DatabaseConnect from '../views/DatabaseConnect.vue'
import DatabaseInfo from '../views/DatabaseInfo.vue'

const routes = [
  {
    path: '/',
    redirect: '/connect'
  },
  {
    path: '/connect',
    name: 'Connect',
    component: DatabaseConnect
  },
  {
    path: '/database/:dbName',
    name: 'DatabaseInfo',
    component: DatabaseInfo
  },
  {
    path: '/connection/:name',
    name: 'ActiveConnection',
    component: () => import('../views/ActiveConnection.vue')
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router 