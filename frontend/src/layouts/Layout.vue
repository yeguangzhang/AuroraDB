<template>
  <n-layout has-sider>
    <!-- 侧边栏 -->
    <n-layout-sider
      bordered
      collapse-mode="width"
      :collapsed-width="64"
      :width="240"
      :collapsed="collapsed"
      show-trigger
      @collapse="collapsed = true"
      @expand="collapsed = false"
    >
      <n-menu
        :collapsed="collapsed"
        :collapsed-width="64"
        :collapsed-icon-size="22"
        :options="menuStore.menuItems"
        :value="menuStore.selectedKey"
        @update:value="handleMenuSelect"
      />
    </n-layout-sider>

    <!-- 主要内容区域 -->
    <n-layout>
      <n-layout-content>
        <router-view></router-view>
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMenuStore } from '../stores/menuStore'

const router = useRouter()
const menuStore = useMenuStore()
const collapsed = ref(false)

// 处理菜单选择
function handleMenuSelect(key) {
  menuStore.setCurrentKey(key)
  
  if (key === 'connect') {
    router.push('/connect')
  } else if (key.startsWith('active-connection-')) {
    const connectionName = key.replace('active-connection-', '')
    router.push(`/connection/${connectionName}`)
  }
}
</script>

<style scoped>
.n-layout {
  height: 100vh;
}

.n-layout-content {
  height: 100%;
}
</style> 