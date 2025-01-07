<template>
  <div class="container">
    <!-- 顶部卡片 -->
    <n-card size="small">
      <n-space align="center" justify="space-between">
        <n-select v-model:value="currentDatabase" :options="databaseOptions" @update:value="handleDatabaseChange"
          style="width: 200px" />
        <n-space>
          <n-tag :type="getStatusType(connectionStatus)">
            {{ connectionStatus }}
          </n-tag>
          <n-button size="small" type="error" @click="handleDisconnect">
            断开连接
          </n-button>
        </n-space>
      </n-space>
    </n-card>

    <!-- 主要内容区域 -->
    <div class="main-content">
      <!-- 左侧菜单 -->
      <div class="sidebar">
        <n-card size="small" :bordered="false">
<n-menu
  v-model:value="selectedTable"
  :options="tableMenuOptions"
  @update:value="handleTableSelect"
>
  <template #default="{ option }">
    <n-tooltip placement="right">
      <template #trigger>
        <span>{{ option.label }}</span>
      </template>
      {{ option.label }}
    </n-tooltip>
  </template>
</n-menu>
        </n-card>
      </div>

      <!-- 右侧内容区域 -->
      <div class="content">
        <n-card :bordered="false">
          <!-- 数据库统计图表 -->
          <div v-if="!selectedTable" class="database-stats">

            <n-space justify="space-around">
                <div class="stat-item">
                  <span class="label">数据库名称：</span>
                  <span class="value">{{ currentDatabase }}</span>
                </div>
                <div class="stat-item">
                  <span class="label">表数量：</span>
                  <span class="value">{{ tables.length }}</span>
                </div>
                <div class="stat-item">
                  <span class="label">总记录数：</span>
                  <span class="value">{{ totalRecords }}</span>
                </div>
              </n-space>
              <n-divider />
            <div ref="recordsChartRef" class="chart-container"></div>
          </div>
          
          <n-tabs v-model:value="activeTab" v-else>
            <n-tab-pane name="data" tab="数据预览">
              <div class="table-wrapper">
                <n-table
                  v-if="tableColumns.length > 0"
                  :bordered="true"
                  :single-line="false"
                >
                  <thead>
                    <tr>
                      <th v-for="col in tableColumns" :key="col.key">
                        {{ col.title }}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(row, index) in tableData" :key="index">
                      <td v-for="(col, colIndex) in tableColumns" :key="colIndex">
                        {{ Array.isArray(row) ? row[colIndex] : row[col.key] }}
                      </td>
                    </tr>
                  </tbody>
                </n-table>
                <n-pagination
                  v-if="total > 0"
                  v-model:page="currentPage"
                  v-model:page-size="pageSize"
                  :item-count="total"
                  :page-sizes="[10, 20, 50, 100]"
                  show-size-picker
                  show-quick-jumper
                  @update:page="handlePageChange"
                  @update:page-size="handlePageSizeChange"
                />
                <n-empty v-else description="无数据" />
              </div>
            </n-tab-pane>
            
            <n-tab-pane name="structure" tab="表结构">
              <div class="table-wrapper">
                <n-table
                  :bordered="true"
                  :single-line="false"
                >
                  <thead>
                    <tr>
                      <th>字段名</th>
                      <th>类型</th>
                      <th>允许空</th>
                      <th>默认值</th>
                      <th>备注</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(item, index) in tableStructure" :key="index">
                      <td>{{ item.name }}</td>
                      <td>{{ item.type }}</td>
                      <td>{{ item.nullable }}</td>
                      <td>{{ item.default || '-' }}</td>
                      <td>{{ item.comment || '-' }}</td>
                    </tr>
                  </tbody>
                </n-table>
              </div>
            </n-tab-pane>
          </n-tabs>
        </n-card>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage ,NTooltip} from 'naive-ui'
import { h } from 'vue'
import * as echarts from 'echarts'
import {
  DisconnectDatabase,
  GetDatabases,
  GetTables,
  UseDatabase,
  GetTableStats,
  GetTableData,
  GetTableStructure
} from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { useMenuStore } from '../stores/menuStore'
import { TableOutlined } from '@vicons/antd'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const menuStore = useMenuStore()

const connectionName = ref(route.params.name)
const defaultDB = ref(route.query.defaultDB || '')
const connectionStatus = ref('connected')
const currentDatabase = ref('')
const databases = ref([])
const tables = ref([])
const selectedTable = ref(null)
const totalRecords = ref(0)
const recordsChartRef = ref(null)
let recordsChart = null
const tableStats = ref([])

// 添加这些变量的声明
const expandedKeys = ref([])
const selectedKeys = ref([])

// 添加新的状态变量
const loading = ref(false)
const tableColumns = ref([])
const tableData = ref([])
const tableStructure = ref([])

// 添加分页相关的状态
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 计算当前页显示的数据
const displayData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return tableData.value.slice(start, end)
})

// 处理页码变化
const handlePageChange = (page) => {
  currentPage.value = page
  if (selectedTable.value) {
    loadTableData(selectedTable.value)
  }
}

// 处理每页条数变化
const handlePageSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1 // 重置到第一页
  if (selectedTable.value) {
    loadTableData(selectedTable.value)
  }
}

// 表结构列定义
const structureColumns = ref([
  { title: '字段名', key: 'name' },
  { title: '类型', key: 'type' },
  { title: '允许空', key: 'nullable' },
  { title: '默认值', key: 'default' },
  { title: '备注', key: 'comment' }
])

// 数据库选项
const databaseOptions = computed(() => {
  return databases.value.map(db => ({
    label: db,
    value: db
  }))
})

// 表菜单选项
const tableMenuOptions = computed(() => {
  return tables.value.map(table => ({
    label: table.name,
    key: table.name,
    icon: renderIcon(TableOutlined)
  }))
})

// 渲染图标
function renderIcon(icon) {
  return () => h(icon)
}

// 处理数据库切换
async function handleDatabaseChange(dbName) {
  if (!dbName) return

  try {
    loading.value = true
    selectedTable.value = null

    await UseDatabase(connectionName.value, dbName)
    await Promise.all([
      loadTables(dbName),
      loadDatabaseStats()
    ])

    currentDatabase.value = dbName
  } catch (error) {
    message.error('切换数据库失败：' + error)
  } finally {
    loading.value = false
  }
}

// 处理表选择
async function handleTableSelect(tableName) {
  if (!tableName) return
  
  selectedTable.value = tableName
  activeTab.value = 'data' // 重置为数据预览标签
  
  try {
    // 确保已经连接到数据库
    if (!connectionName.value) {
      throw new Error('未连接到数据库')
    }
    
    loading.value = true
    currentPage.value = 1 // 重置页码
    
    // 加载数据
    await Promise.all([
      loadTableData(tableName),
      loadTableStructure(tableName)
    ])
  } catch (error) {
    console.error('加载表详情失败:', error)
    message.error('加载表详情失败：' + error)
  } finally {
    loading.value = false
  }
}

// 加载表数据
async function loadTableData(tableName) {
  try {
    loading.value = true
    console.log('Loading table data for:', tableName)
    
    const result = await GetTableData(
      connectionName.value,
      tableName,
      {
        page: currentPage.value,
        pageSize: pageSize.value
      }
    )

    console.log('Received data:', result)

    if (!result || !result.columns || !result.data) {
      console.warn('数据格式不正确:', result)
      tableColumns.value = []
      tableData.value = []
      total.value = 0
      return
    }

    // 设置列
    tableColumns.value = result.columns.map(col => ({
      title: col.title || col.key,
      key: col.key
    }))

    // 设置数据 - 直接使用数组格式
    tableData.value = result.data
    total.value = result.total || 0

  } catch (error) {
    console.error('加载表数据失败:', error)
    message.error('加载表数据失败：' + error)
    tableColumns.value = []
    tableData.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// 加载表结构
async function loadTableStructure(tableName) {
  try {
    console.log('Loading structure for:', tableName)
    const structure = await GetTableStructure(
      connectionName.value,
      currentDatabase.value,
      tableName
    )
    console.log('Received structure:', structure)
    tableStructure.value = structure || []
  } catch (error) {
    console.error('加载表结构失败:', error)
    message.error('加载表结构失败：' + error)
    tableStructure.value = []
  }
}

// 加载表统计信息
async function loadTableStats(tableName) {
  try {
    tableStats.value = await GetTableStats(
      connectionName.value,
      currentDatabase.value,
      tableName
    )
  } catch (error) {
    message.error('加载表统计失败：' + error)
  }
}

// 加载数据库统计信息
async function loadDatabaseStats() {
  try {
    const stats = await GetTableStats(connectionName.value, currentDatabase.value)
    totalRecords.value = stats.totalRecords
    tableStats.value = stats.tableStats

    await nextTick()
    initChart()
  } catch (error) {
    message.error('加载统计信息失败：' + error)
  }
}

// 初始化图表
function initChart() {
  console.log('初始化图表...')
  console.log('图表容器:', recordsChartRef.value)
  console.log('统计数据:', tableStats.value)

  if (!recordsChartRef.value) {
    console.error('图表容器未找到')
    return
  }

  if (recordsChart) {
    console.log('销毁旧图表实例')
    recordsChart.dispose()
  }

  recordsChart = echarts.init(recordsChartRef.value)
  
  // 添加resize监听
  window.addEventListener('resize', () => {
    if (recordsChart) {
      console.log('窗口大小改变，重新调整图表')
      recordsChart.resize()
    }
  })

  if (!tableStats.value || tableStats.value.length === 0) {
    console.error('没有可用的统计数据')
    return
  }

  const option = {
      title: {
        text: '数据库表记录统计',
        left: 'center',
        top: 10
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'shadow'
        }
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        top: '60px',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        data: tableStats.value.map(stat => stat.tableName),
        axisLabel: {
          rotate: 45
        }
      },
      yAxis: {
        type: 'value'
      },
      series: [{
        name: '记录数',
        type: 'bar',
        data: tableStats.value.map(stat => stat.recordCount),
        itemStyle: {
          color: '#409EFF'
        },
        emphasis: {
          focus: 'series'
        }
      }],
      dataZoom: [{
        type: 'inside',
        start: 0,
        end: 100
      }, {
        start: 0,
        end: 100
      }]
    }

  recordsChart.setOption(option)
  
  // 确保图表正确渷
  setTimeout(() => {
    recordsChart.resize()
  }, 0)
}

// 获取数据库列表
async function loadDatabases() {
  try {
    const dbList = await GetDatabases(connectionName.value)
    databases.value = dbList
    if (dbList.length > 0 && !currentDatabase.value) {
      currentDatabase.value = dbList[0]
      defaultDB.value = dbList[0]
      await loadTables(dbList[0])
    }
  } catch (error) {
    console.error('Failed to load databases:', error)
    message.error('获取数据库列表失败：' + error)
  }
}

// 获取数据表列表
async function loadTables(dbName) {
  try {
    const tableList = await GetTables(connectionName.value, dbName)
    tables.value = tableList.map(table => ({
      name: table
    }))
    // 强制更新 treeData
    currentDatabase.value = dbName
  } catch (error) {
    console.error('Failed to load tables:', error)
    message.error('获取数据表列表失败：' + error)
  }
}

// 处理菜单选择
async function handleMenuSelect(key) {
  if (key.startsWith('db-')) {
    const dbName = key.replace('db-', '')
    currentDatabase.value = dbName
    showTables.value = false
    try {
      await UseDatabase(connectionName.value, dbName)
      await loadTables(dbName)
    } catch (error) {
      message.error('切换数据库失败：' + error)
    }
  } else if (key.startsWith('tables-')) {
    const dbName = key.replace('tables-', '')
    currentDatabase.value = dbName
    showTables.value = true
  }
}

// 查看表详情
function handleViewTable(tableName) {
  router.push({
    name: 'TableDetail',
    params: {
      connection: connectionName.value,
      database: currentDatabase.value,
      table: tableName
    }
  })
}

const getStatusType = (status) => {
  switch (status) {
    case 'connected': return 'success'
    case 'disconnected': return 'error'
    default: return 'default'
  }
}

async function handleDisconnect() {
  try {
    await DisconnectDatabase(connectionName.value)
    menuStore.removeActiveConnection(`active-connection-${connectionName.value}`)
    router.push('/connect')
    message.success('已断开连接')
  } catch (error) {
    message.error('断开连接失败：' + error)
  }
}

// 设置事件监听
function setupEventListeners() {
  EventsOn(`db:${connectionName.value}:status`, (status) => {
    connectionStatus.value = status
  })

  EventsOn(`db:${connectionName.value}:error`, (error) => {
    message.error(error)
  })
}

// 清理事件监听
function cleanupEventListeners() {
  EventsOff(`db:${connectionName.value}:status`)
  EventsOff(`db:${connectionName.value}:error`)
}

// 处理窗口大小变化
const handleResize = () => {
  if (recordsChart) {
    recordsChart.resize()
  }
}

// 组件挂载时初始化
onMounted(async () => {
  setupEventListeners()
  await loadDatabases()
  // 设置菜单选中状态
  menuStore.addActiveConnection({
    key: `active-connection-${connectionName.value}`,
    label: connectionName.value,
    defaultDB: defaultDB.value
  })

  // 如果有默认数据库，自动选择并加载统计信息
  if (defaultDB.value) {
    currentDatabase.value = defaultDB.value
    await UseDatabase(connectionName.value, defaultDB.value)
    await Promise.all([
      loadTables(defaultDB.value),
      loadDatabaseStats()
    ])
  }

  // 确保DOM渲染完成后再初始化图表
  nextTick(() => {
    window.addEventListener('resize', handleResize)
    initChart()
  })
})

// 组件卸载时清理
onUnmounted(() => {
  cleanupEventListeners()
  window.removeEventListener('resize', handleResize)
  if (recordsChart) {
    recordsChart.dispose()
    recordsChart = null
  }
})

// 添加激活标签页的状态
const activeTab = ref('data')

// 修改数据结构
const treeData = computed(() => {
  return databases.value.map(db => ({
    key: db,
    label: db,
    icon: renderIcon(DatabaseOutlined),
    children: currentDatabase.value === db ? tables.value.map(table => ({
      key: `${db}-${table.name}`,
      label: table.name,
      icon: renderIcon(TableOutlined)
    })) : [] // 只有当前选中的数据库才显示表列表
  }))
})

// 处理树节点选择
async function handleSelect(keys) {
  if (keys.length === 0) return

  const key = keys[0]
  selectedKeys.value = keys

  // 如果是数据库节点
  if (!key.includes('-')) {
    currentDatabase.value = key
    try {
      await UseDatabase(connectionName.value, key)
      await loadTables(key)
      // 自动展开当前数据库节点
      if (!expandedKeys.value.includes(key)) {
        expandedKeys.value = [...expandedKeys.value, key]
      }
    } catch (error) {
      message.error('切换数据库失败：' + error)
    }
  }
}

// 处理树节点展开
function handleExpand(keys) {
  expandedKeys.value = keys
}

// 初始化加载
async function initializeConnection() {
  try {
    loading.value = true
    await loadDatabases()

    if (defaultDB.value) {
      currentDatabase.value = defaultDB.value
      await UseDatabase(connectionName.value, defaultDB.value)
      await Promise.all([
        loadTables(defaultDB.value),
        loadDatabaseStats()
      ])
    }
  } catch (error) {
    message.error('初始化连接失败：' + error)
  } finally {
    loading.value = false
  }
}

</script>

<style scoped src="./ActiveConnection.css"></style>
