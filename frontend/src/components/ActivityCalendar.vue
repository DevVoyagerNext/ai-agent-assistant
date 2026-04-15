<script setup lang="ts">
import { computed, ref, watch, onMounted, nextTick } from 'vue'
import type { ActivityCalendarItem } from '../types/user'

const props = defineProps<{
  items: ActivityCalendarItem[]
  weeks?: number
}>()

const scrollAreaRef = ref<HTMLElement | null>(null)
const WEEKS = computed(() => props.weeks ?? 53)

const endDate = ref(new Date())
const today = computed(() => new Date())

const toYMD = (d: Date) => {
  const yyyy = d.getFullYear()
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  return `${yyyy}-${mm}-${dd}`
}

const startOfWeekSunday = (d: Date) => {
  const copy = new Date(d)
  const day = copy.getDay()
  copy.setDate(copy.getDate() - day)
  copy.setHours(0, 0, 0, 0)
  return copy
}

const shiftDays = (d: Date, days: number) => {
  const copy = new Date(d)
  copy.setDate(copy.getDate() + days)
  return copy
}

const labels = computed(() => ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'])

const byDate = computed<Map<string, { count: number; score: number }>>(() => {
  const map = new Map<string, { count: number; score: number }>()
  for (const it of props.items || []) {
    if (it?.date) {
      map.set(it.date, { count: it.count || 0, score: it.score || 0 })
    }
  }
  return map
})

const endOfWeekSaturday = (d: Date) => {
  const copy = new Date(d)
  const day = copy.getDay()
  copy.setDate(copy.getDate() + (6 - day))
  copy.setHours(23, 59, 59, 999)
  return copy
}

const startDate = computed(() => {
  // 始终以当前 endDate 所在周的周六作为网格终点
  const last = endOfWeekSaturday(endDate.value)
  const days = WEEKS.value * 7 - 1
  const firstCandidate = shiftDays(last, -days)
  // 网格起点始终是 53 周前的周日
  return startOfWeekSunday(firstCandidate)
})

const weeksData = computed(() => {
  const cols: { date: string; count: number; score: number; d: Date }[][] = []
  let cursor = new Date(startDate.value)
  for (let w = 0; w < WEEKS.value; w++) {
    const col: { date: string; count: number; score: number; d: Date }[] = []
    for (let i = 0; i < 7; i++) {
      const ymd = toYMD(cursor)
      const mapped = byDate.value.get(ymd)
      col.push({
        date: ymd,
        count: mapped?.count ?? 0,
        score: mapped?.score ?? 0,
        d: new Date(cursor),
      })
      cursor = shiftDays(cursor, 1)
    }
    cols.push(col)
  }
  return cols
})

const monthLabels = computed(() => {
  const labels: { text: string; weekIndex: number }[] = []
  const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  for (let w = 0; w < weeksData.value.length; w++) {
    const firstDayOfWeek = weeksData.value[w][0].d
    const isFirstOfMonth = firstDayOfWeek.getDate() <= 7
    if (isFirstOfMonth) {
      labels.push({
        text: monthNames[firstDayOfWeek.getMonth()],
        weekIndex: w,
      })
    }
  }
  return labels
})

const levelFor = (value: number) => {
  if (!value || value <= 0) return 0
  if (value <= 5) return 1
  if (value <= 15) return 2
  if (value <= 30) return 3
  return 4
}

const availableYears = computed(() => {
  const years = new Set<number>()
  years.add(today.value.getFullYear())
  for (const it of props.items || []) {
    if (it?.date) {
      const year = new Date(it.date).getFullYear()
      years.add(year)
    }
  }
  return Array.from(years).sort((a, b) => b - a)
})

const selectedYear = ref(today.value.getFullYear())

const scrollToToday = async (smooth = true) => {
  await nextTick()
  if (!scrollAreaRef.value) return

  const todayYMD = toYMD(today.value)
  let todayWeekIndex = -1

  for (let w = 0; w < weeksData.value.length; w++) {
    if (weeksData.value[w].some(cell => cell.date === todayYMD)) {
      todayWeekIndex = w
      break
    }
  }

  if (todayWeekIndex !== -1) {
    const columnWidth = 14 // 12px cell + 2px gap
    const targetScroll = todayWeekIndex * columnWidth
    // 尽量让今天显示在中间或者偏右侧，而不是紧贴左边缘
    const containerWidth = scrollAreaRef.value.clientWidth
    const offset = Math.max(0, targetScroll - containerWidth / 2)
    
    scrollAreaRef.value.scrollTo({
      left: offset,
      behavior: smooth ? 'smooth' : 'auto'
    })
  } else {
    // 如果找不到今天（比如选择了往年），则滚动到最右侧
    scrollAreaRef.value.scrollTo({
      left: scrollAreaRef.value.scrollWidth,
      behavior: smooth ? 'smooth' : 'auto'
    })
  }
}

const selectYear = (year: number) => {
  selectedYear.value = year
  if (year === today.value.getFullYear()) {
    endDate.value = today.value
  } else {
    // 设置为选中年份的最后一天（12月31日）
    endDate.value = new Date(year, 11, 31)
  }
  scrollToToday(true)
}

onMounted(() => {
  scrollToToday(false) // 初始进入时不平滑，直接定位
})

watch(() => props.items, () => {
  scrollToToday(true)
})
</script>

<template>
  <div class="gh-calendar-container">
    <div class="gh-calendar-main">
      <div class="calendar-content">
        <!-- 固定在左侧的星期标签 -->
        <div class="y-labels-column">
          <div class="y-label-spacer"></div>
          <div class="y-labels">
            <span class="y-label" style="grid-row: 2">{{ labels[1] }}</span>
            <span class="y-label" style="grid-row: 4">{{ labels[3] }}</span>
            <span class="y-label" style="grid-row: 6">{{ labels[5] }}</span>
          </div>
        </div>

        <!-- 可横向滚动的日历区域 -->
        <div ref="scrollAreaRef" class="gh-calendar-scroll-area">
          <div class="months" :style="{ gridTemplateColumns: `repeat(${weeksData.length}, 14px)` }">
            <span
              v-for="m in monthLabels"
              :key="m.weekIndex"
              class="month"
              :style="{ gridColumnStart: m.weekIndex + 1 }"
            >
              {{ m.text }}
            </span>
          </div>
          <div class="grid" :style="{ gridTemplateColumns: `repeat(${weeksData.length}, 14px)` }">
            <div v-for="(week, wi) in weeksData" :key="wi" class="col">
              <div
                v-for="(cell, di) in week"
                :key="`${wi}-${di}`"
                class="cell"
                :class="`lv-${levelFor(cell.score)}`"
                :title="`${cell.date}: 操作 ${cell.count} 次，活跃度 ${cell.score}${cell.date === toYMD(today) ? ' (今天)' : ''}`"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- 固定在下方的图例 -->
      <div class="legend">
        <span>低</span>
        <span class="legend-cell lv-0"></span>
        <span class="legend-cell lv-1"></span>
        <span class="legend-cell lv-2"></span>
        <span class="legend-cell lv-3"></span>
        <span class="legend-cell lv-4"></span>
        <span>高</span>
      </div>
    </div>
    
    <div class="year-nav">
      <button 
        v-for="year in availableYears" 
        :key="year"
        class="year-btn"
        :class="{ active: selectedYear === year }"
        @click="selectYear(year)"
      >
        {{ year }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.gh-calendar-container {
  display: flex;
  gap: 24px;
  align-items: flex-start;
  justify-content: flex-start;
  padding: 8px 0;
}

.gh-calendar-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

.calendar-content {
  display: flex;
  gap: 8px;
}

.y-labels-column {
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.y-label-spacer {
  height: 20px; /* 与月份行高度一致 */
}

.gh-calendar-scroll-area {
  flex: 1;
  overflow-x: auto;
  padding-bottom: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.year-nav {
  width: 72px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  border-left: 1px solid rgba(0,0,0,0.05);
  padding-left: 16px;
}

.year-btn {
  padding: 6px 0;
  width: 100%;
  border-radius: 6px;
  border: 1px solid transparent;
  background: transparent;
  color: #615d59;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  text-align: center;
  transition: all 0.2s;
}

.year-btn:hover {
  background: rgba(0,0,0,0.05);
  color: rgba(0,0,0,0.95);
}

.year-btn.active {
  background: #f2f9ff;
  color: #0075de;
  border-color: rgba(0,117,222,0.1);
}

.months {
  display: grid;
  height: 20px;
  align-items: center;
  gap: 0px;
}

.month {
  color: #a39e98;
  font-size: 11px;
  font-weight: 500;
}

.y-labels {
  display: grid;
  grid-template-rows: repeat(7, 14px);
  gap: 2px;
}

.y-label {
  font-size: 11px;
  color: #a39e98;
  line-height: 14px;
  text-align: right;
  padding-right: 4px;
  font-weight: 500;
}

.grid {
  display: grid;
  gap: 2px;
}

.col {
  display: grid;
  grid-template-rows: repeat(7, 14px);
  gap: 2px;
}

.cell {
  width: 12px;
  height: 12px;
  border-radius: 2px;
  background: #f6f5f4;
  transition: transform 0.1s ease;
}

.cell:hover {
  transform: scale(1.2);
  z-index: 10;
}

/* Notion Green activity levels */
.lv-0 { background: #f6f5f4; }
.lv-1 { background: #d3f4e0; }
.lv-2 { background: #92e6b5; }
.lv-3 { background: #4fd38a; }
.lv-4 { background: #1aae39; }

.legend {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #a39e98;
  font-size: 11px;
  font-weight: 500;
  margin-top: 4px;
  padding-left: 36px; /* 对齐 y-labels-column 的宽度 */
}

.legend-cell {
  width: 10px;
  height: 10px;
  border-radius: 2px;
  display: inline-block;
}

.legend-cell.lv-0 { background: #f6f5f4; }
.legend-cell.lv-1 { background: #d3f4e0; }
.legend-cell.lv-2 { background: #92e6b5; }
.legend-cell.lv-3 { background: #4fd38a; }
.legend-cell.lv-4 { background: #1aae39; }
</style>
