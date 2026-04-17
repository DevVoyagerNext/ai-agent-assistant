<script setup lang="ts">
import { ref, computed, onMounted, watch, onBeforeUnmount } from 'vue'
import { Search, BookOpen, Star, ArrowRight, User, Compass, Bookmark, Bot } from 'lucide-vue-next'
import { useRouter } from 'vue-router'
import { getSubjectCategories, getAllSubjects, getSubjectsByCategory, searchSubjects } from '../api/subject'
import type { Subject, SubjectCategory } from '../types/subject'

const router = useRouter()
const isLoggedIn = ref(false)
const syncLoginState = () => {
  isLoggedIn.value = !!localStorage.getItem('token')
}

const subjects = ref<Subject[]>([])
const categories = ref<SubjectCategory[]>([])
const currentCategory = ref<number | 'all'>('all')
const searchQuery = ref('')
const committedKeyword = ref('')
const suggestions = ref<Subject[]>([])
const showSuggestions = ref(false)
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
let suggestTimer: number | undefined
let suggestSeq = 0
let searchSeq = 0

const fetchCategories = async () => {
  try {
    const res = await getSubjectCategories()
    if (res.data?.code === 200 && res.data.data) {
      categories.value = res.data.data
    }
  } catch (error) {
    console.error('获取分类失败', error)
  }
}

const fetchSubjects = async () => {
  loading.value = true
  try {
    let res;
    if (currentCategory.value === 'all') {
      res = await getAllSubjects()
    } else {
      res = await getSubjectsByCategory(currentCategory.value as number)
    }
    
    if (res.data?.code === 200 && res.data.data) {
      subjects.value = res.data.data
      total.value = res.data.data.length
    } else {
      subjects.value = []
      total.value = 0
    }
  } catch (error) {
    console.error('获取教材失败', error)
    subjects.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  syncLoginState()
  await fetchCategories()
  await fetchSubjects()
})

const handleCategoryChange = async (catId: number | 'all') => {
  searchQuery.value = ''
  committedKeyword.value = ''
  suggestions.value = []
  showSuggestions.value = false
  page.value = 1
  total.value = 0
  searchSeq += 1
  suggestSeq += 1
  currentCategory.value = catId
  await fetchSubjects()
}

const resetSearch = async () => {
  searchQuery.value = ''
  committedKeyword.value = ''
  suggestions.value = []
  showSuggestions.value = false
  page.value = 1
  total.value = 0
  searchSeq += 1
  suggestSeq += 1
  await fetchSubjects()
}

// 获取分类名称辅助函数
const getCategoryName = (catId: number | 'all') => {
  if (catId === 'all') return '全部'
  const cat = categories.value.find(c => c.id === catId)
  return cat ? cat.name : '未知分类'
}

const getCoverStyle = (subject: Subject) => {
  const palettes: Array<[string, string]> = [
    ['#3b82f6', '#8b5cf6'],
    ['#06b6d4', '#3b82f6'],
    ['#22c55e', '#14b8a6'],
    ['#f97316', '#ef4444'],
    ['#a855f7', '#ec4899']
  ]
  const [from, to] = palettes[subject.id % palettes.length]
  return {
    background: `linear-gradient(135deg, ${from}, ${to})`
  }
}

const isSearchMode = computed(() => committedKeyword.value.trim().length > 0)
const hasMore = computed(() => isSearchMode.value && subjects.value.length < total.value)
const tagText = computed(() => (isSearchMode.value ? '搜索结果' : getCategoryName(currentCategory.value)))

const runSearch = async (reset: boolean) => {
  const keyword = committedKeyword.value.trim()
  if (!keyword) {
    page.value = 1
    total.value = 0
    await fetchSubjects()
    return
  }

  if (reset) {
    page.value = 1
    subjects.value = []
    total.value = 0
  }

  const currentPage = page.value
  const seq = (searchSeq += 1)
  loading.value = true
  try {
    const res = await searchSubjects(keyword, currentPage, pageSize.value)
    if (seq !== searchSeq) return
    if (res.data?.code === 200 && res.data.data) {
      total.value = res.data.data.total || 0
      const list = res.data.data.list || []
      subjects.value = currentPage === 1 ? list : [...subjects.value, ...list]
    } else {
      subjects.value = []
      total.value = 0
    }
  } catch (error) {
    if (seq !== searchSeq) return
    console.error('搜索教材失败', error)
    subjects.value = []
    total.value = 0
  } finally {
    if (seq === searchSeq) loading.value = false
  }
}

const handleSearchClick = async () => {
  const keyword = searchQuery.value.trim()
  suggestions.value = []
  showSuggestions.value = false
  committedKeyword.value = keyword
  page.value = 1
  total.value = 0
  searchSeq += 1
  await runSearch(true)
}

const loadMore = async () => {
  if (!hasMore.value || loading.value) return
  page.value += 1
  await runSearch(false)
}

const fetchSuggestions = async () => {
  const keyword = searchQuery.value.trim()
  if (!keyword || keyword === committedKeyword.value.trim()) {
    suggestions.value = []
    showSuggestions.value = false
    return
  }

  const seq = (suggestSeq += 1)
  try {
    const res = await searchSubjects(keyword, 1, 8)
    if (seq !== suggestSeq) return
    if (res.data?.code === 200 && res.data.data) {
      suggestions.value = res.data.data.list || []
      showSuggestions.value = suggestions.value.length > 0
    } else {
      suggestions.value = []
      showSuggestions.value = false
    }
  } catch (error) {
    if (seq !== suggestSeq) return
    console.error('搜索建议失败', error)
    suggestions.value = []
    showSuggestions.value = false
  }
}

const selectSuggestion = async (subject: Subject) => {
  searchQuery.value = subject.name
  committedKeyword.value = subject.name
  suggestions.value = []
  showSuggestions.value = false
  page.value = 1
  total.value = 0
  searchSeq += 1
  await runSearch(true)
}

watch(
  () => searchQuery.value,
  (val) => {
    if (suggestTimer) window.clearTimeout(suggestTimer)
    const keyword = val.trim()
    if (!keyword) {
      page.value = 1
      total.value = 0
      committedKeyword.value = ''
      suggestions.value = []
      showSuggestions.value = false
      searchSeq += 1
      fetchSubjects()
      return
    }
    suggestTimer = window.setTimeout(() => {
      fetchSuggestions()
    }, 400)
  }
)

onBeforeUnmount(() => {
  if (suggestTimer) window.clearTimeout(suggestTimer)
})

const goToStudy = (id: number) => {
  // 跳转到沉浸式学习页
  router.push(`/subject/${id}`)
}

const handleUserAction = () => {
  syncLoginState()
  if (isLoggedIn.value) {
    router.push('/me')
  } else {
    router.push('/login')
  }
}
</script>

<template>
  <div class="market-container">
    <!-- 顶部导航 -->
    <nav class="market-nav">
      <div class="nav-brand">
        <Compass class="brand-icon" :size="28" />
        <span class="brand-name">AI 学海大厅</span>
      </div>
      <div class="nav-actions">
        <button class="login-btn ai-btn" @click="router.push('/ai-chat')">
          <Bot :size="18" />
          <span>AI 助手</span>
        </button>
        <button class="login-btn" @click="handleUserAction">
          <User :size="18" />
          <span>{{ isLoggedIn ? '个人中心' : '登录 / 注册' }}</span>
        </button>
      </div>
    </nav>

    <!-- 华丽的 Hero 区域 -->
    <header class="hero-section">
      <div class="hero-content">
        <h1 class="hero-title">探索知识的无限可能</h1>
        <p class="hero-subtitle">发现最前沿的 AI 技术、编程语言与计算机科学课程。无需等待，即刻开启你的专属学习图谱。</p>
        
        <div class="search-wrap">
          <div class="search-bar">
            <Search class="search-icon" :size="20" />
            <input 
              v-model="searchQuery" 
              type="text" 
              placeholder="搜索你感兴趣的教材，例如 'AI Agent'..." 
              class="search-input"
              @focus="showSuggestions = suggestions.length > 0"
              @keydown.enter.prevent="handleSearchClick"
            />
            <button class="search-btn" @click="handleSearchClick">探索</button>
          </div>
          <div v-if="showSuggestions" class="search-dropdown">
            <button
              v-for="s in suggestions"
              :key="s.id"
              type="button"
              class="search-item"
              @click="selectSuggestion(s)"
            >
              <span class="search-item-title">{{ s.name }}</span>
              <span class="search-item-progress">{{ s.progressPercent || 0 }}%</span>
            </button>
          </div>
        </div>
      </div>
      <div class="hero-bg-shapes">
        <div class="shape shape-1"></div>
        <div class="shape shape-2"></div>
        <div class="shape shape-3"></div>
      </div>
    </header>

    <main class="main-content">
      <!-- 全量教材大厅 -->
      <section class="all-subjects-section">
        <div class="section-header" v-if="!isSearchMode">
          <h2 class="section-title">
            <BookOpen class="title-icon" :size="24" />
            发现教材
          </h2>
        </div>
        
        <!-- 分类筛选器 -->
        <div class="category-filter" v-if="!isSearchMode">
          <button 
            class="filter-btn"
            :class="{ active: currentCategory === 'all' }"
            @click="handleCategoryChange('all')"
          >
            全部
          </button>
          <button 
            v-for="cat in categories" 
            :key="cat.id"
            class="filter-btn"
            :class="{ active: currentCategory === cat.id }"
            @click="handleCategoryChange(cat.id)"
          >
            {{ cat.name }}
          </button>
        </div>

        <!-- 加载状态 -->
        <div v-if="loading" class="empty-state">
          <h3>加载中...</h3>
        </div>

        <template v-else>
          <div v-if="subjects.length > 0" class="subject-grid">
            <div v-for="subject in subjects" :key="subject.id" class="subject-card" @click="goToStudy(subject.id)">
              <div class="card-cover" :style="getCoverStyle(subject)">
                <div class="card-overlay">
                  <button class="start-btn">开始学习 <ArrowRight :size="16" /></button>
                </div>
              </div>
              <div class="card-body">
                <span class="category-tag">{{ tagText }}</span>
                <h3 class="subject-title">{{ subject.name }}</h3>
                <p class="subject-desc">{{ subject.description }}</p>
                <div class="card-footer">
                  <div class="stats-left">
                    <div class="stat" v-if="subject.isLiked"><Star class="star-icon" :size="14" /> 点赞</div>
                    <div class="stat" v-if="subject.isCollected"><Bookmark class="collect-icon" :size="14" /> 收藏</div>
                  </div>
                  <div class="stat"><BookOpen :size="14" /> {{ subject.progressPercent || 0 }}% 进度</div>
                </div>
              </div>
            </div>
          </div>

          <div v-if="subjects.length > 0 && hasMore" class="load-more-wrap">
            <button class="load-more-btn" @click="loadMore">加载更多</button>
          </div>

          <div v-else-if="subjects.length === 0 && isSearchMode" class="empty-state">
            <div class="empty-icon">🔍</div>
            <h3>未找到相关教材</h3>
            <p>尝试更换关键词或分类筛选</p>
            <button class="reset-btn" @click="resetSearch">重置筛选</button>
          </div>

          <div v-else-if="subjects.length === 0" class="empty-state">
            <div class="empty-icon">📚</div>
            <h3>暂无教材</h3>
            <p>请稍后重试</p>
            <button class="reset-btn" @click="fetchSubjects">刷新列表</button>
          </div>
        </template>
      </section>
    </main>
  </div>
</template>

<style scoped>
.market-container {
  width: 100vw;
  height: 100vh;
  overflow-y: auto;
  background-color: #f8fafc;
  font-family: system-ui, -apple-system, sans-serif;
}

/* 导航栏 */
.market-nav {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 64px;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(12px);
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 40px;
  z-index: 100;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.nav-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #3b82f6;
}

.brand-name {
  font-size: 20px;
  font-weight: 700;
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.nav-actions {
  display: flex;
  gap: 12px;
}

.login-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 20px;
  color: #475569;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.login-btn:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
  color: #0f172a;
}

.ai-btn {
  background: linear-gradient(135deg, #eff6ff, #e0e7ff);
  border-color: #bfdbfe;
  color: #3b82f6;
}

.ai-btn:hover {
  background: linear-gradient(135deg, #dbeafe, #c7d2fe);
  border-color: #93c5fd;
  color: #2563eb;
}

/* Hero 区域 */
.hero-section {
  position: relative;
  padding: 120px 20px 80px;
  background: linear-gradient(135deg, #eff6ff 0%, #e0e7ff 100%);
  overflow: hidden;
  display: flex;
  justify-content: center;
  align-items: center;
  text-align: center;
}

.hero-content {
  position: relative;
  z-index: 10;
  max-width: 800px;
}

.hero-title {
  font-size: 48px;
  font-weight: 800;
  color: #0f172a;
  margin-bottom: 20px;
  line-height: 1.2;
  letter-spacing: -0.02em;
}

.hero-subtitle {
  font-size: 18px;
  color: #475569;
  margin-bottom: 40px;
  line-height: 1.6;
}

.search-wrap {
  position: relative;
  max-width: 600px;
  margin: 0 auto;
}

.search-bar {
  display: flex;
  align-items: center;
  background: white;
  padding: 8px;
  border-radius: 99px;
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1);
  width: 100%;
}

.search-icon {
  color: #94a3b8;
  margin-left: 16px;
}

.search-input {
  flex: 1;
  border: none;
  outline: none;
  padding: 12px 16px;
  font-size: 16px;
  background: transparent;
}

.search-btn {
  background: #3b82f6;
  color: white;
  border: none;
  padding: 12px 28px;
  border-radius: 99px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.search-btn:hover {
  background: #2563eb;
}

.search-dropdown {
  position: absolute;
  left: 0;
  right: 0;
  top: calc(100% + 10px);
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(226, 232, 240, 0.9);
  border-radius: 14px;
  overflow: hidden;
  box-shadow: 0 18px 30px rgba(15, 23, 42, 0.12);
  z-index: 20;
}

.search-item {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: transparent;
  border: none;
  cursor: pointer;
  text-align: left;
}

.search-item + .search-item {
  border-top: 1px solid rgba(241, 245, 249, 0.9);
}

.search-item:hover {
  background: rgba(241, 245, 249, 0.9);
}

.search-item-title {
  font-size: 14px;
  font-weight: 650;
  color: #0f172a;
}

.search-item-progress {
  font-size: 13px;
  font-weight: 800;
  color: #3b82f6;
}

/* 装饰背景 */
.hero-bg-shapes {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
  z-index: 1;
}

.shape {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.6;
}

.shape-1 {
  width: 400px;
  height: 400px;
  background: #bfdbfe;
  top: -100px;
  left: -100px;
}

.shape-2 {
  width: 300px;
  height: 300px;
  background: #ddd6fe;
  bottom: -50px;
  right: 10%;
}

.shape-3 {
  width: 250px;
  height: 250px;
  background: #fbcfe8;
  top: 20%;
  right: -50px;
}

/* 主体内容 */
.main-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 60px 20px;
}

.section-header {
  margin-bottom: 30px;
  display: flex;
  align-items: flex-end;
  gap: 16px;
}

.section-title {
  font-size: 28px;
  font-weight: 700;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 10px;
}

.title-icon {
  color: #3b82f6;
}

.hot-icon {
  color: #ef4444;
}

.section-desc {
  color: #64748b;
  font-size: 15px;
  margin-bottom: 6px;
}

/* 卡片网格 */
.subject-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 30px;
}

.subject-card {
  background: white;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  cursor: pointer;
  display: flex;
  flex-direction: column;
}

.subject-card:hover {
  transform: translateY(-8px);
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1);
}

.subject-card:hover .card-overlay {
  opacity: 1;
}

.subject-card:hover .card-cover {
  transform: scale(1.02);
}

.card-cover {
  height: 180px;
  background-size: cover;
  background-position: center;
  position: relative;
  transition: transform 0.5s ease;
}

.hot-badge {
  position: absolute;
  top: 12px;
  left: 12px;
  background: linear-gradient(135deg, #ef4444, #f97316);
  color: white;
  padding: 4px 12px;
  border-radius: 99px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.5px;
  box-shadow: 0 2px 4px rgba(239, 68, 68, 0.3);
}

.card-overlay {
  position: absolute;
  inset: 0;
  background: rgba(15, 23, 42, 0.4);
  display: flex;
  justify-content: center;
  align-items: center;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.start-btn {
  background: white;
  color: #0f172a;
  border: none;
  padding: 10px 20px;
  border-radius: 99px;
  font-weight: 600;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 6px;
  transform: translateY(10px);
  transition: all 0.3s ease;
}

.subject-card:hover .start-btn {
  transform: translateY(0);
}

.card-body {
  padding: 24px;
  display: flex;
  flex-direction: column;
  flex: 1;
  background: white;
  position: relative;
  z-index: 2;
}

.category-tag {
  align-self: flex-start;
  font-size: 12px;
  font-weight: 600;
  color: #3b82f6;
  background: #eff6ff;
  padding: 4px 10px;
  border-radius: 6px;
  margin-bottom: 12px;
}

.subject-title {
  font-size: 18px;
  font-weight: 700;
  color: #0f172a;
  margin-bottom: 8px;
  line-height: 1.4;
}

.subject-desc {
  font-size: 14px;
  color: #64748b;
  line-height: 1.6;
  margin-bottom: 20px;
  flex: 1;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 16px;
  border-top: 1px solid #f1f5f9;
  color: #64748b;
  font-size: 13px;
  font-weight: 500;
}

.stats-left {
  display: flex;
  gap: 12px;
}

.stat {
  display: flex;
  align-items: center;
  gap: 4px;
}

.star-icon {
  color: #f59e0b;
  fill: #f59e0b;
}

.collect-icon {
  color: #f59e0b;
  fill: #f59e0b;
}

/* 分类筛选器 */
.all-subjects-section {
  margin-top: 80px;
}

.category-filter {
  display: flex;
  gap: 12px;
  margin-bottom: 30px;
  flex-wrap: wrap;
}

.filter-btn {
  padding: 8px 16px;
  border: 1px solid #e2e8f0;
  background: white;
  border-radius: 99px;
  color: #475569;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-btn:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.filter-btn.active {
  background: #3b82f6;
  color: white;
  border-color: #3b82f6;
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.2);
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 60px 20px;
  background: white;
  border-radius: 16px;
  border: 1px dashed #cbd5e1;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-state h3 {
  font-size: 18px;
  color: #0f172a;
  margin-bottom: 8px;
}

.empty-state p {
  color: #64748b;
  margin-bottom: 24px;
}

.reset-btn {
  background: #eff6ff;
  color: #3b82f6;
  border: none;
  padding: 8px 20px;
  border-radius: 99px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.reset-btn:hover {
  background: #dbeafe;
}

.load-more-wrap {
  margin-top: 24px;
  display: flex;
  justify-content: center;
}

.load-more-btn {
  background: #eff6ff;
  color: #3b82f6;
  border: none;
  padding: 10px 24px;
  border-radius: 99px;
  font-weight: 700;
  cursor: pointer;
  transition: background 0.2s;
}

.load-more-btn:hover {
  background: #dbeafe;
}

/* 响应式 */
@media (max-width: 768px) {
  .hero-title {
    font-size: 36px;
  }
  .market-nav {
    padding: 0 20px;
  }
}
</style>
