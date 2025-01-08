<template>
  <n-config-provider :theme="darkTheme">
    <n-message-provider>
      <n-layout style="height: 100vh">
        <!-- 顶部 -->
<!--        <n-layout-header bordered style="height: 48px; padding: 0">-->
<!--          <div class="header-content">-->
<!--            <div class="header-left">-->
<!--              <n-button quaternary circle @click="collapsed = !collapsed">-->
<!--                <template #icon>-->
<!--                  <n-icon>-->
<!--                    <menu-fold-outlined v-if="collapsed" />-->
<!--                    <menu-unfold-outlined v-else />-->
<!--                  </n-icon>-->
<!--                </template>-->
<!--              </n-button>-->
<!--              <div class="header-title">AuroraDB</div>-->
<!--            </div>-->
<!--            <n-button quaternary circle size="small">-->
<!--              <template #icon>-->
<!--                <n-icon><setting-outlined /></n-icon>-->
<!--              </template>-->
<!--            </n-button>-->
<!--          </div>-->
<!--        </n-layout-header>-->

        <!-- 中间主体 -->
        <n-layout has-sider position="absolute" style="top: 0; bottom: 32px">
          <n-layout-sider
            bordered
            show-trigger
            :collapsed="collapsed"
            :collapsed-width="64"
            collapse-mode="width"
            :width="240"
            @collapse="collapsed = true"
            @expand="collapsed = false"
          >
            <n-menu
              :options="menuStore.menuItems"
              :value="menuStore.selectedKey"
              :expanded-keys="menuStore.expandedKeys"
              :collapsed="collapsed"

              @update:value="handleMenuUpdate"
              @update:expanded-keys="menuStore.updateExpandedKeys"
            />
          </n-layout-sider>
          <n-layout-content content-style="padding: 16px;">
            <router-view></router-view>
          </n-layout-content>
        </n-layout>

        <!-- 底部 -->
        <n-layout-footer bordered style="height: 32px; padding: 0">
          <div class="footer-content">
            &copy; {{ new Date().getFullYear() }} Database Tool
          </div>
        </n-layout-footer>
      </n-layout>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { darkTheme, NIcon } from 'naive-ui'
import { SettingOutlined } from '@vicons/antd'
import { MenuFoldOutlined, MenuUnfoldOutlined } from '@vicons/antd'
import { useMenuStore } from './stores/menuStore'

const router = useRouter()
const menuStore = useMenuStore()
const collapsed = ref(false)

const handleMenuUpdate = (key) => {
  menuStore.setCurrentKey(key)
  const path = menuStore.getPathByKey(key)
  if (path) {
    router.push(path)
  }
}
</script>

<style>
/* 顶部区域 */
.header-content {
  height: 100%;
  padding: 0 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(90deg, #1867C0 0%, #5CBBF6 100%);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-title {
  font-size: 16px;
  font-weight: bold;
  color: white;
}

/* 底部区域 */
.footer-content {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
}

/* 按钮样式 */
.header-content .n-button {
  color: white !important;
}

.header-content .n-button:hover {
  background-color: rgba(255, 255, 255, 0.2) !important;
}

/* 滚动条样式 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: var(--n-scrollbar-color);
}

::-webkit-scrollbar-thumb {
  background: var(--n-scrollbar-color-hover);
  border-radius: 3px;
}
</style>
