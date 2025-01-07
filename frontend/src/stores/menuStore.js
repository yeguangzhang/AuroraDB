import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { h } from 'vue'
import { NIcon } from 'naive-ui'
import { DatabaseOutlined, ApiOutlined } from '@vicons/antd'

// 渲染图标的辅助函数
function renderIcon(icon) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

export const useMenuStore = defineStore('menu', () => {
  const currentKey = ref('connect')
  const expandedKeys = ref(['database'])
  const activeConnections = ref([])

  // 基础菜单项
  const baseMenuItems = [
    {
      label: '数据库管理',
      key: 'database',
      icon: renderIcon(DatabaseOutlined),
      children: [
        {
          label: '配置管理',
          key: 'connect',
          icon: renderIcon(ApiOutlined),
          path: '/connect'
        }
      ]
    }
  ]

  // 计算菜单项
  const menuItems = computed(() => {
    const items = [...baseMenuItems]
    
    if (activeConnections.value.length > 0) {
      items.push({
        label: '活动连接',
        key: 'active-connections',
        icon: renderIcon(ApiOutlined),
        children: activeConnections.value.map(conn => ({
          label: conn.label,
          key: conn.key,
          icon: renderIcon(DatabaseOutlined),
          path: `/connection/${conn.label}`,
          defaultDB: conn.defaultDB
        }))
      })
    }
    
    return items
  })

  // 添加活动连接
  function addActiveConnection(connection) {
    const existingIndex = activeConnections.value.findIndex(
      conn => conn.key === connection.key
    )
    
    if (existingIndex === -1) {
      activeConnections.value.push(connection)
      // 确保 'active-connections' 菜单展开
      if (!expandedKeys.value.includes('active-connections')) {
        expandedKeys.value.push('active-connections')
      }
    }
    
    currentKey.value = connection.key
  }

  // 移除活动连接
  function removeActiveConnection(key) {
    const index = activeConnections.value.findIndex(conn => conn.key === key)
    if (index !== -1) {
      activeConnections.value.splice(index, 1)
      currentKey.value = 'connect'
      
      // 如果没有活动连接了，从展开列表中移除
      if (activeConnections.value.length === 0) {
        expandedKeys.value = expandedKeys.value.filter(k => k !== 'active-connections')
      }
    }
  }

  // 更新展开的菜单项
  function updateExpandedKeys(keys) {
    expandedKeys.value = keys
  }

  // 获取菜单项的路径
  function getPathByKey(key) {
    const findPath = (items) => {
      for (const item of items) {
        if (item.key === key) return item.path
        if (item.children) {
          const path = findPath(item.children)
          if (path) return path
        }
      }
      return null
    }
    
    return findPath(menuItems.value)
  }

  // 设置当前选中的菜单项
  function setCurrentKey(key) {
    currentKey.value = key
  }

  const selectedKey = computed(() => currentKey.value)

  return {
    menuItems,
    selectedKey,
    expandedKeys,
    activeConnections,
    addActiveConnection,
    removeActiveConnection,
    setCurrentKey,
    getPathByKey,
    updateExpandedKeys
  }
}) 