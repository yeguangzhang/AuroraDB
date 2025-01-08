<template>
  <n-space vertical class="database-connect">
    <!-- 保存的配置列表 -->
    <n-card title="数据库连接">
      <n-space>
        <n-select
            v-model:value="selectedConfig"
            :options="savedConfigs"
            style="min-width: 200px"
            placeholder="选择已保存的配置"
        />
        <n-button type="error" @click="deleteConfig" :disabled="!selectedConfig">
          删除连接
        </n-button>
        <n-button type="success" @click="showCreateModal">
          创建连接
        </n-button>
        <n-button type="success" @click="showCreateModal">
          修改连接
        </n-button>
        <n-button type="primary"  :loading="loading" @click="handleConnect">
          连接
        </n-button>
      </n-space>
    </n-card>

    <!-- 创建连接弹窗 -->
    <n-modal
        v-model:show="showModal"
        preset="dialog"
        title="新建连接"
        :style="{ width: '600px' }"
        positive-text="确认"
        negative-text="取消"
        @positive-click="handleCreateConfig"
        @negative-click="closeModal"
    >
      <n-form
          ref="createFormRef"
          :model="newConfig"
          :rules="rules"
          label-placement="left"
          label-width="100"
          require-mark-placement="right-hanging"
      >
        <n-form-item label="配置名称" path="name">
          <n-input v-model:value="newConfig.name" placeholder="请输入配置名称"/>
        </n-form-item>

        <n-form-item label="数据库类型" path="type">
          <n-select
              v-model:value="newConfig.type"
              :options="dbTypeOptions"
              placeholder="请选择数据库类型"
          />
        </n-form-item>

        <n-divider>数据库连接信息</n-divider>

        <n-form-item label="主机地址" path="host">
          <n-input v-model:value="newConfig.host" placeholder="请输入主机地址"/>
        </n-form-item>

        <n-form-item label="端口" path="port">
          <n-input-number v-model:value="newConfig.port" placeholder="请输入端口号"/>
        </n-form-item>

        <n-form-item label="用户名" path="username">
          <n-input v-model:value="newConfig.username" placeholder="请输入用户名"/>
        </n-form-item>

        <n-form-item label="密码" path="password">
          <n-input
              v-model:value="newConfig.password"
              type="password"
              show-password-on="click"
              placeholder="请输入密码"
          />
        </n-form-item>

        <n-form-item label="数据库名" path="database">
          <n-input v-model:value="newConfig.database" placeholder="请输入数据库名"/>
        </n-form-item>

        <n-divider>SSH 隧道</n-divider>

        <n-form-item label="使用 SSH">
          <n-switch v-model:value="newConfig.useSSH"/>
        </n-form-item>

        <template v-if="newConfig.useSSH">
          <n-form-item label="SSH 主机" path="ssh.host">
            <n-input v-model:value="newConfig.ssh.host" placeholder="请输入 SSH 主机地址"/>
          </n-form-item>

          <n-form-item label="SSH 端口" path="ssh.port">
            <n-input-number v-model:value="newConfig.ssh.port" placeholder="SSH 端口号"/>
          </n-form-item>

          <n-form-item label="SSH 用户名" path="ssh.username">
            <n-input v-model:value="newConfig.ssh.username" placeholder="SSH 用户名"/>
          </n-form-item>

          <n-form-item label="认证方式" path="ssh.authType">
            <n-radio-group v-model:value="newConfig.ssh.authType">
              <n-space>
                <n-radio value="password">密码</n-radio>
                <n-radio value="privateKey">私钥</n-radio>
              </n-space>
            </n-radio-group>
          </n-form-item>

          <n-form-item v-if="newConfig.ssh.authType === 'password'" label="SSH 密码" path="ssh.password">
            <n-input
                v-model:value="newConfig.ssh.password"
                type="password"
                show-password-on="click"
                placeholder="SSH 密码"
            />
          </n-form-item>

          <template v-else>
            <n-form-item label="私钥文件" path="ssh.privateKey">
              <n-input
                  v-model:value="newConfig.ssh.privateKey"
                  type="textarea"
                  placeholder="请输入私钥内容"
              />
            </n-form-item>

            <n-form-item label="私钥密码" path="ssh.passphrase">
              <n-input
                  v-model:value="newConfig.ssh.passphrase"
                  type="password"
                  show-password-on="click"
                  placeholder="私钥密码（如果有）"
              />
            </n-form-item>
          </template>
        </template>
      </n-form>
    </n-modal>

    <!-- 显示连接状态 -->
<!--    <n-tag v-if="connectionStatus[selectedConfig.name]" :type="getStatusType(connectionStatus[selectedConfig.name])">-->
<!--      {{ connectionStatus[selectedConfig.name] }}-->
<!--    </n-tag>-->
  </n-space>
</template>

<script setup>
import {onMounted, onUnmounted, ref} from 'vue'
import {useMessage} from 'naive-ui'
import {
  ConnectDatabase,
  DeleteConfiguration,
  DisconnectDatabase,
  LoadConfigurations,
  SaveConfiguration
} from '../../wailsjs/go/main/App'
import {EventsOff, EventsOn} from '../../wailsjs/runtime/runtime'
import {useRouter} from 'vue-router'
import {useMenuStore} from '../stores/menuStore'

const message = useMessage()
const loading = ref(false)
const createFormRef = ref(null)
const selectedConfig = ref(null)
const savedConfigs = ref([])
const showModal = ref(false)
const connectionStatus = ref({})
const router = useRouter()
const menuStore = useMenuStore()

const dbTypeOptions = [
  {label: 'MySQL', value: 'mysql'},
  {label: 'PostgreSQL', value: 'postgresql'},
  {label: 'SQLite', value: 'sqlite'},
  {label: 'TdEngine', value: 'TdEngine'},
]

// 新的配置模板
const defaultConfig = {
  name: '',
  type: 'mysql',
  host: 'localhost',
  port: 3306,
  username: 'root',
  password: '',
  database: '',
  useSSH: false,
  ssh: {
    host: '',
    port: 22,
    username: '',
    authType: 'password',
    password: '',
    privateKey: '',
    passphrase: ''
  }
}

const newConfig = ref({...defaultConfig})

// 表单验证规则
const rules = {
  name: {required: true, message: '请输入配置名称'},
  type: {required: true, message: '请选择数据库类型'},
  host: {required: true, message: '请输入主机地址'},
  port: {required: true, message: '请输入端口号'},
  username: {required: true, message: '请输入用户名'},
  'ssh.host': {
    required: true,
    message: '请输入 SSH 主机地址',
    trigger: 'blur',
    validator: (rule, value) => {
      // 只在启用 SSH 时验证
      return !newConfig.value.useSSH || !!value
    }
  },
  'ssh.port': {
    required: true,
    message: '请输入 SSH 端口号',
    trigger: 'blur',
    validator: (rule, value) => {
      // 只在启用 SSH 时验证
      return !newConfig.value.useSSH || !!value
    }
  },
  'ssh.username': {
    required: true,
    message: '请输入 SSH 用户名',
    trigger: 'blur',
    validator: (rule, value) => {
      // 只在启用 SSH 时验证
      return !newConfig.value.useSSH || !!value
    }
  }
}

// 监听连接状态
function setupEventListeners(configName) {
  // 监听连接状态
  EventsOn(`db:${configName}:status`, (status) => {
    connectionStatus.value[configName] = status
  })

  // 监听错误消息
  EventsOn(`db:${configName}:error`, (error) => {
    message.error(error)
  })

  // 监听查询结果
  EventsOn(`db:${configName}:queryResult`, (results) => {
    console.log('Query results:', results)
  })

  // 监听表名
  EventsOn(`db:${configName}:tables`, (tables) => {
    console.log('Tables:', tables)
    router.push({
      name: 'DatabaseInfo',
      params: {dbName: configName},
      query: {tables: tables.join(',')}
    })
  })

  // 监听数据库名
  EventsOn(`db:${configName}:database`, (databaseName) => {
    console.log('DatabaseName:', databaseName)
    router.push({
      name: 'DatabaseInfo',
      params: {dbName: configName},
      query: {database: databaseName}
    })
  })
}

// 清理事件监听
function cleanupEventListeners(configName) {
  EventsOff(`db:${configName}:status`)
  EventsOff(`db:${configName}:error`)
  EventsOff(`db:${configName}:queryResult`)
}

// 显示创建弹窗
function showCreateModal() {
  showModal.value = true
  newConfig.value = {...defaultConfig}
}

// 关闭弹窗
function closeModal() {
  showModal.value = false
}

// 加载保存的配置
async function loadSavedConfigs() {
  try {
    const configs = await LoadConfigurations()
    savedConfigs.value = configs.map(config => ({
      label: config.name,
      value: config.name,
      config: config
    }))
  } catch (error) {
    message.error('加载配置失败：' + error)
  }
}

// 修改 onMounted
onMounted(() => {
  loadSavedConfigs()
})

// 保存配置
async function saveConfig() {
  if (!config.value.name) {
    message.error('请输入配置名称')
    return
  }

  try {
    await SaveConfiguration(config.value)
    await loadSavedConfigs()
    message.success('配置已保存')
  } catch (error) {
    message.error('保存配置失败：' + error)
  }
}

// 删除配置
async function deleteConfig() {
  if (!selectedConfig.value) return

  try {
    await DeleteConfiguration(selectedConfig.value)
    await loadSavedConfigs()
    selectedConfig.value = null
    message.success('配置已删除')
  } catch (error) {
    message.error('删除配置失败：' + error)
  }
}

// 创建新配置
async function handleCreateConfig() {
  try {
    await createFormRef.value?.validate()

    if (savedConfigs.value.some(c => c.config.name === newConfig.value.name)) {
      message.error('配置名称已存在')
      return false
    }

    if (!newConfig.value.useSSH) {
      delete newConfig.value.ssh
    }

    await SaveConfiguration(newConfig.value)
    await loadSavedConfigs()
    message.success('配置已创建')
    selectedConfig.value = newConfig.value.name
    config.value = {...newConfig.value}
    showModal.value = false
    return true
  } catch (error) {
    console.error('创建配置失败:', error)
    // 修改错误处理
    if (error.message) {
      message.error('创建配置失败：' + error.message)
    } else if (typeof error === 'string') {
      message.error('创建配置失败：' + error)
    } else {
      message.error('创建配置失败，请检查配置信息')
    }
    return false
  }
}

// 加载选中的配置
function loadConfig() {
  if (!selectedConfig.value) return

  const selected = savedConfigs.value.find(c => c.value === selectedConfig.value)
  if (selected) {
    config.value = {...selected.config}
    message.success('配置已加载')
  }
}

// 添加状态类型判断函数
const getStatusType = (status) => {
  switch (status) {
    case 'connected':
      return 'success'
    case 'connecting':
      return 'warning'
    case 'disconnected':
      return 'error'
    default:
      return 'default'
  }
}

// 修改 handleConnect 函数
async function handleConnect() {
  try {
    loading.value = true
    console.log(selectedConfig.value)
    setupEventListeners(selectedConfig.value)
    connectionStatus.value[selectedConfig.value] = 'connecting'

    // 建立连接
    await ConnectDatabase(selectedConfig.value)

    // 添加到活跃连接菜单
    menuStore.addActiveConnection({
      label: selectedConfig.value,
      key: `active-connection-${selectedConfig.value}`,
      path: `/connection/${selectedConfig.value}`,
      icon: 'database',
      defaultDB: selectedConfig.value || ''
    })

    // 跳转到新连接页面，并传递默认数据库信息
    router.push({
      name: 'ActiveConnection',
      params: {
        name: selectedConfig.value,
      },
      query: {
        defaultDB: selectedConfig.value || ''
      }
    })

  } catch (error) {
    message.error('连接失败：' + error)
    connectionStatus.value[selectedConfig.value] = 'disconnected'
  } finally {
    loading.value = false
  }
}

// 修改 handleDisconnect 函数
async function handleDisconnect(configName) {
  try {
    await DisconnectDatabase(configName)
    cleanupEventListeners(configName)
    // 从活跃连接菜单中移除
    menuStore.removeActiveConnection(`active-connection-${configName}`)
  } catch (error) {
    message.error('断开连接失败：' + error)
  }
}

// 组件卸载时清理所有事件监听
onUnmounted(() => {
  Object.keys(connectionStatus.value).forEach(configName => {
    cleanupEventListeners(configName)
  })
})
</script>

<style scoped>
.database-connect {
  height: 100%;
  overflow: visible;
}

.n-space {
  overflow: visible;
}

.card-title {
  font-size: 16px;
  font-weight: bold;
}
</style> 