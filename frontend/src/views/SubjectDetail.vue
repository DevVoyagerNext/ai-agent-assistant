<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getTopNodes, getChildNodes, getNodeDetail, getNodeNote } from '../api/node'
import type { SubjectNode, SubjectNodeDetail, NodeNote } from '../types/node'
import { ChevronRight, ChevronDown, FileText, ArrowLeft, Edit3, CheckCircle2, Circle } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const subjectId = Number(route.params.id)
const isLoggedIn = computed(() => !!localStorage.getItem('token'))

// ----------------- 目录树相关 -----------------
// 定义一个扩展类型用于在树中维护展开状态和子节点
interface TreeNode extends SubjectNode {
  expanded?: boolean;
  children?: TreeNode[];
  loadingChildren?: boolean;
}

const topNodes = ref<TreeNode[]>([])
const loadingTree = ref(false)

const fetchTopNodes = async () => {
  loadingTree.value = true
  try {
    const res = await getTopNodes(subjectId)
    if (res.data?.code === 200 && res.data.data) {
      topNodes.value = res.data.data.map(node => ({ ...node, expanded: false, children: [] }))
    }
  } catch (error) {
    console.error('获取顶级节点失败', error)
  } finally {
    loadingTree.value = false
  }
}

const toggleNode = async (node: TreeNode) => {
  // 如果是叶子节点，直接加载详情
  if (node.isLeaf === 1) {
    await selectNode(node.id)
    return
  }

  // 展开/收起
  node.expanded = !node.expanded
  
  // 如果展开且没有子节点，则去拉取
  if (node.expanded && (!node.children || node.children.length === 0)) {
    node.loadingChildren = true
    try {
      const res = await getChildNodes(node.id)
      if (res.data?.code === 200 && res.data.data) {
        node.children = res.data.data.map(child => ({ ...child, expanded: false, children: [] }))
      }
    } catch (error) {
      console.error('获取子节点失败', error)
    } finally {
      node.loadingChildren = false
    }
  }

  // 顺便也加载一下当前点击的节点详情（可选）
  await selectNode(node.id)
}

// ----------------- 正文和笔记相关 -----------------
const currentNodeId = ref<number | null>(null)
const nodeDetail = ref<SubjectNodeDetail | null>(null)
const nodeNote = ref<NodeNote | null>(null)
const loadingDetail = ref(false)

const selectNode = async (id: number) => {
  if (currentNodeId.value === id) return
  currentNodeId.value = id
  loadingDetail.value = true
  
  try {
    const resDetail = await getNodeDetail(id)
    if (resDetail.data?.code === 200 && resDetail.data.data) {
      nodeDetail.value = resDetail.data.data
    } else {
      nodeDetail.value = null
    }

    if (isLoggedIn.value) {
      const resNote = await getNodeNote(id)
      if (resNote.data?.code === 200 && resNote.data.data) {
        nodeNote.value = resNote.data.data
      } else {
        nodeNote.value = null
      }
    }
  } catch (error) {
    console.error('获取详情或笔记失败', error)
  } finally {
    loadingDetail.value = false
  }
}

onMounted(() => {
  if (!subjectId) {
    router.replace('/')
    return
  }
  fetchTopNodes()
})

const goBack = () => {
  router.push('/')
}

</script>

<template>
  <div class="study-layout">
    <!-- 左侧目录树 -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <button class="back-btn" @click="goBack">
          <ArrowLeft :size="18" />
          返回大厅
        </button>
        <h3>课程目录</h3>
      </div>
      
      <div class="tree-container">
        <div v-if="loadingTree" class="loading-state">加载目录中...</div>
        <div v-else-if="topNodes.length === 0" class="empty-state">暂无目录数据</div>
        <ul v-else class="tree-list">
          <template v-for="node in topNodes" :key="node.id">
            <!-- 递归树节点，由于层级不深，先写死两层或抽取递归组件，这里手写简易递归 -->
            <li class="tree-item">
              <div 
                class="node-content" 
                :class="{ active: currentNodeId === node.id }"
                @click="toggleNode(node)"
              >
                <span class="expand-icon" v-if="node.isLeaf === 0">
                  <ChevronDown v-if="node.expanded" :size="16" />
                  <ChevronRight v-else :size="16" />
                </span>
                <span class="expand-icon placeholder" v-else></span>
                
                <CheckCircle2 v-if="node.userProgressStatus === 'completed'" class="status-icon completed" :size="16" />
                <Circle v-else class="status-icon unstarted" :size="16" />
                
                <span class="node-name">{{ node.name }}</span>
              </div>
              
              <!-- 子节点 -->
              <ul v-if="node.expanded && node.children && node.children.length > 0" class="tree-list sub-list">
                <li v-for="child in node.children" :key="child.id" class="tree-item">
                  <div 
                    class="node-content sub-content" 
                    :class="{ active: currentNodeId === child.id }"
                    @click="toggleNode(child)"
                  >
                    <span class="expand-icon" v-if="child.isLeaf === 0">
                      <ChevronDown v-if="child.expanded" :size="16" />
                      <ChevronRight v-else :size="16" />
                    </span>
                    <span class="expand-icon placeholder" v-else></span>

                    <CheckCircle2 v-if="child.userProgressStatus === 'completed'" class="status-icon completed" :size="16" />
                    <Circle v-else class="status-icon unstarted" :size="16" />

                    <span class="node-name">{{ child.name }}</span>
                  </div>
                  
                  <!-- 支持第三层级（如果有） -->
                  <ul v-if="child.expanded && child.children && child.children.length > 0" class="tree-list sub-list">
                    <li v-for="grandchild in child.children" :key="grandchild.id" class="tree-item">
                      <div 
                        class="node-content sub-content grand-content" 
                        :class="{ active: currentNodeId === grandchild.id }"
                        @click="selectNode(grandchild.id)"
                      >
                        <span class="expand-icon placeholder"></span>
                        <CheckCircle2 v-if="grandchild.userProgressStatus === 'completed'" class="status-icon completed" :size="16" />
                        <Circle v-else class="status-icon unstarted" :size="16" />
                        <span class="node-name">{{ grandchild.name }}</span>
                      </div>
                    </li>
                  </ul>
                  <div v-if="child.loadingChildren" class="loading-sub">加载中...</div>
                </li>
              </ul>
              <div v-if="node.loadingChildren" class="loading-sub">加载中...</div>
            </li>
          </template>
        </ul>
      </div>
    </aside>

    <!-- 中间正文区 -->
    <main class="main-content">
      <div v-if="loadingDetail" class="loading-state">
        内容加载中...
      </div>
      <div v-else-if="nodeDetail" class="content-area">
        <div class="content-header">
          <h1>{{ nodeDetail.name }}</h1>
          <div class="tags">
            <span class="tag">难度: {{ nodeDetail.easyCount }}易 / {{ nodeDetail.mediumCount }}中 / {{ nodeDetail.hardCount }}难</span>
            <span class="tag status" :class="nodeDetail.userProgressStatus">
              {{ nodeDetail.userProgressStatus === 'completed' ? '已掌握' : nodeDetail.userProgressStatus === 'learning' ? '学习中' : '未开始' }}
            </span>
          </div>
        </div>
        <div class="markdown-body">
          <!-- 实际项目中建议使用 marked.js + highlight.js 渲染 -->
          <pre>{{ nodeDetail.content }}</pre>
        </div>
      </div>
      <div v-else class="empty-state">
        <FileText :size="48" color="#cbd5e1" />
        <p>请在左侧选择要学习的章节</p>
      </div>
    </main>

    <!-- 右侧随堂笔记区 -->
    <aside class="note-sidebar">
      <div class="note-header">
        <Edit3 :size="20" class="note-icon" />
        <h3>随堂笔记</h3>
      </div>
      
      <div class="note-container">
        <div v-if="!isLoggedIn" class="not-logged-in">
          <p>登录后即可记录随堂笔记</p>
          <button class="login-btn" @click="router.push('/login')">去登录</button>
        </div>
        <div v-else-if="!currentNodeId" class="empty-note">
          <p>请先选择一个知识点</p>
        </div>
        <div v-else class="note-editor">
          <textarea 
            class="note-textarea" 
            :value="nodeNote?.noteContent || ''"
            placeholder="在这里记录你的思考和感悟..."
            readonly
          ></textarea>
          <div class="note-footer">
            <span class="save-time" v-if="nodeNote?.updatedAt">上次保存: {{ nodeNote.updatedAt }}</span>
            <span class="save-time" v-else>暂无笔记</span>
            <button class="save-btn" disabled>保存笔记(仅展示)</button>
          </div>
        </div>
      </div>
    </aside>
  </div>
</template>

<style scoped>
.study-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  background: #f8fafc;
  overflow: hidden;
}

/* 左侧侧边栏 */
.sidebar {
  width: 300px;
  background: white;
  border-right: 1px solid #e2e8f0;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-header {
  padding: 16px 20px;
  border-bottom: 1px solid #f1f5f9;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: none;
  color: #64748b;
  font-size: 14px;
  cursor: pointer;
  margin-bottom: 12px;
  padding: 4px 0;
  transition: color 0.2s;
}

.back-btn:hover {
  color: #3b82f6;
}

.sidebar-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: #0f172a;
}

.tree-container {
  flex: 1;
  overflow-y: auto;
  padding: 12px 0;
}

.tree-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.sub-list {
  background: #fafafa;
}

.node-content {
  display: flex;
  align-items: center;
  padding: 10px 16px;
  cursor: pointer;
  transition: background 0.2s;
  color: #334155;
  font-size: 14px;
}

.node-content:hover {
  background: #f1f5f9;
}

.node-content.active {
  background: #eff6ff;
  color: #2563eb;
  font-weight: 500;
}

.sub-content {
  padding-left: 32px;
}

.grand-content {
  padding-left: 48px;
}

.expand-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  color: #94a3b8;
  margin-right: 4px;
}

.expand-icon.placeholder {
  visibility: hidden;
}

.status-icon {
  margin-right: 8px;
}

.status-icon.completed {
  color: #10b981;
}

.status-icon.unstarted {
  color: #cbd5e1;
}

.node-name {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.loading-sub {
  padding: 8px 16px 8px 48px;
  font-size: 12px;
  color: #94a3b8;
}

/* 中间正文区 */
.main-content {
  flex: 1;
  background: white;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.content-area {
  padding: 40px;
  max-width: 800px;
  margin: 0 auto;
  width: 100%;
}

.content-header {
  margin-bottom: 30px;
  border-bottom: 1px solid #f1f5f9;
  padding-bottom: 20px;
}

.content-header h1 {
  font-size: 28px;
  font-weight: 700;
  color: #0f172a;
  margin-bottom: 16px;
}

.tags {
  display: flex;
  gap: 12px;
}

.tag {
  font-size: 12px;
  padding: 4px 10px;
  background: #f1f5f9;
  color: #64748b;
  border-radius: 4px;
  font-weight: 500;
}

.tag.status.completed { background: #d1fae5; color: #059669; }
.tag.status.learning { background: #dbeafe; color: #2563eb; }
.tag.status.unstarted { background: #f1f5f9; color: #64748b; }

.markdown-body {
  font-size: 16px;
  line-height: 1.8;
  color: #334155;
}

.markdown-body pre {
  background: #f8fafc;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
  white-space: pre-wrap;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

/* 右侧笔记区 */
.note-sidebar {
  width: 320px;
  background: #f8fafc;
  border-left: 1px solid #e2e8f0;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.note-header {
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 8px;
  border-bottom: 1px solid #e2e8f0;
}

.note-icon {
  color: #8b5cf6;
}

.note-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: #0f172a;
}

.note-container {
  flex: 1;
  padding: 20px;
  display: flex;
  flex-direction: column;
}

.not-logged-in, .empty-note {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #64748b;
  font-size: 14px;
}

.login-btn {
  margin-top: 12px;
  padding: 8px 24px;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
}

.note-editor {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  overflow: hidden;
}

.note-textarea {
  flex: 1;
  border: none;
  resize: none;
  padding: 16px;
  font-size: 14px;
  line-height: 1.6;
  color: #334155;
  outline: none;
  background: transparent;
}

.note-footer {
  padding: 12px 16px;
  border-top: 1px solid #f1f5f9;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fafafa;
}

.save-time {
  font-size: 12px;
  color: #94a3b8;
}

.save-btn {
  background: #8b5cf6;
  color: white;
  border: none;
  padding: 6px 16px;
  border-radius: 4px;
  font-size: 13px;
  cursor: not-allowed;
  opacity: 0.7;
}

/* 公共空状态和加载态 */
.loading-state, .empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
  font-size: 14px;
  gap: 12px;
}
</style>