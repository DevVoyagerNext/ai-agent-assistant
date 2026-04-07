<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { ActivityCalendarItem } from '../types/user'

const props = defineProps<{
  items: ActivityCalendarItem[]
  weeks?: number
}>()

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

const byDate = computed<Map<string, number>>(() => {
  const map = new Map<string, number>()
  for (const it of props.items || []) {
    if (it?.date) {
      map.set(it.date, it.count || 0)
    }
  }
  return map
})

const startDate = computed(() => {
  const last = new Date(endDate.value)
  const days = WEEKS.value * 7 - 1
  const alignedLast = new Date(last)
  alignedLast.setHours(0, 0, 0, 0)
  const firstCandidate = shiftDays(alignedLast, -days)
  return startOfWeekSunday(firstCandidate)
})

const weeksData = computed(() => {
  const cols: { date: string; count: number; d: Date }[][] = []
  let cursor = new Date(startDate.value)
  for (let w = 0; w < WEEKS.value; w++) {
    const col: { date: string; count: number; d: Date }[] = []
    for (let i = 0; i < 7; i++) {
      const ymd = toYMD(cursor)
      col.push({
        date: ymd,
        count: byDate.value.get(ymd) ?? 0,
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

const levelFor = (count: number) => {
  if (!count || count <= 0) return 0
  if (count <= 3) return 1
  if (count <= 7) return 2
  if (count <= 15) return 3
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

const selectYear = (year: number) => {
  selectedYear.value = year
  if (year === today.value.getFullYear()) {
    endDate.value = today.value
  } else {
    // 设置为选中年份的最后一天（12月31日）
    endDate.value = new Date(year, 11, 31)
  }
}

watch(() => props.items, () => {
  // Keep current endDate, data will reactively update
})
</script>

<template>
  <div class="gh-calendar-container">
    <div class="gh-calendar">
      <div class="calendar-header">
        <div class="left">
        <div class="y-spacer"></div>
        <div class="months" :style="{ gridTemplateColumns: `repeat(${weeksData.length}, 15px)` }">
          <span
            v-for="m in monthLabels"
            :key="m.weekIndex"
            class="month"
            :style="{ gridColumnStart: m.weekIndex + 1 }"
          >
            {{ m.text }}
          </span>
        </div>
      </div>
    </div>

    <div class="calendar-body">
      <div class="y-labels">
        <span class="y-label" style="grid-row: 2">{{ labels[1] }}</span>
        <span class="y-label" style="grid-row: 4">{{ labels[3] }}</span>
        <span class="y-label" style="grid-row: 6">{{ labels[5] }}</span>
      </div>
      <div class="grid" :style="{ gridTemplateColumns: `repeat(${weeksData.length}, 15px)` }">
        <div v-for="(week, wi) in weeksData" :key="wi" class="col">
          <div
            v-for="(cell, di) in week"
            :key="`${wi}-${di}`"
            class="cell"
            :class="`lv-${levelFor(cell.count)}`"
            :title="`${cell.date}: ${cell.count} 次活跃`"
          />
        </div>
      </div>
    </div>

      <div class="legend">
        <span>Less</span>
        <span class="legend-cell lv-0"></span>
        <span class="legend-cell lv-1"></span>
        <span class="legend-cell lv-2"></span>
        <span class="legend-cell lv-3"></span>
        <span class="legend-cell lv-4"></span>
        <span>More</span>
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
  gap: 20px;
  align-items: flex-start;
  justify-content: flex-start;
}
.gh-calendar {
  display: flex;
  flex-direction: column;
  gap: 10px;
  overflow-x: auto;
}
.year-nav {
  width: 60px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.year-btn {
  padding: 8px 0;
  width: 100%;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: #64748b;
  font-size: 13px;
  cursor: pointer;
  text-align: center;
  transition: all 0.2s;
}
.year-btn:hover {
  background: rgba(148, 163, 184, 0.1);
  color: #1e293b;
}
.year-btn.active {
  background: #3b82f6;
  color: white;
  font-weight: 500;
}
.calendar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.left {
  display: flex;
  align-items: center;
  gap: 8px;
}
.y-spacer {
  width: 25px;
}
.months {
  display: grid;
  grid-auto-flow: column;
  gap: 0px;
}
.month {
  color: #475569;
  font-size: 12px;
  transform: translateX(-4px); /* Slight adjustment to align with the first column of the month */
}
.tools {
  display: flex;
  gap: 8px;
}
.nav {
  padding: 6px 10px;
  font-size: 12px;
  border-radius: 6px;
  border: 1px solid #cbd5e1;
  background: white;
  color: #0f172a;
}
.nav:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.calendar-body {
  display: flex;
  gap: 8px;
}
.y-labels {
  display: grid;
  grid-template-rows: repeat(7, 15px);
  gap: 0;
  padding-top: 0px;
}
.y-label {
  font-size: 11px;
  color: #64748b;
  line-height: 15px;
  text-align: right;
  padding-right: 4px;
}
.grid {
  display: grid;
  gap: 0;
}
.col {
  display: grid;
  grid-template-rows: repeat(7, 15px);
  gap: 0;
}
.cell {
  width: 12px;
  height: 12px;
  margin: 1.5px;
  border-radius: 2px;
  background: #e5e7eb;
}
.lv-0 { background: #e5e7eb; }
.lv-1 { background: #bbf7d0; }
.lv-2 { background: #86efac; }
.lv-3 { background: #4ade80; }
.lv-4 { background: #16a34a; }
.legend {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #64748b;
  font-size: 12px;
}
.legend-cell {
  width: 12px;
  height: 12px;
  border-radius: 2px;
  display: inline-block;
}
.legend-cell.lv-0 { background: #e5e7eb; }
.legend-cell.lv-1 { background: #bbf7d0; }
.legend-cell.lv-2 { background: #86efac; }
.legend-cell.lv-3 { background: #4ade80; }
.legend-cell.lv-4 { background: #16a34a; }
</style>
