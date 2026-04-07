import { ref, onMounted } from 'vue'
import { 
  getUserInfo, 
  getUserActivitiesCalendar, 
  getPublicPrivateNotes, 
  getSharedNotes, 
  getLearnedSubjects 
} from '../api/user'
import type { 
  UserInfoRes, 
  ActivityCalendarItem, 
  PublicPrivateNoteItem, 
  SharedNoteItem, 
  LearnedSubjectItem 
} from '../types/user'

export const useUserProfile = () => {
  // State
  const userInfo = ref<UserInfoRes | null>(null)
  const activities = ref<ActivityCalendarItem[]>([])
  const publicPrivateNotes = ref<PublicPrivateNoteItem[]>([])
  const sharedNotes = ref<SharedNoteItem[]>([])
  const learnedSubjects = ref<LearnedSubjectItem[]>([])

  // Loading states
  const loadingUserInfo = ref(true)
  const loadingActivities = ref(true)
  const loadingPublicPrivateNotes = ref(true)
  const loadingSharedNotes = ref(true)
  const loadingLearnedSubjects = ref(true)

  // Errors (optional, could be handled globally, but here we keep track to show error states)
  const errorUserInfo = ref('')
  
  // Fetchers
  const fetchUserInfo = async () => {
    loadingUserInfo.value = true
    errorUserInfo.value = ''
    try {
      const res = await getUserInfo()
      if (res.data?.code === 200 && res.data.data) {
        userInfo.value = res.data.data
      } else {
        errorUserInfo.value = res.data?.msg || 'Failed to load user info'
      }
    } catch (err: any) {
      errorUserInfo.value = err?.response?.data?.msg || 'Server error'
    } finally {
      loadingUserInfo.value = false
    }
  }

  const fetchActivities = async () => {
    loadingActivities.value = true
    try {
      const res = await getUserActivitiesCalendar()
      if (res.data?.code === 200 && res.data.data) {
        const apiItems = res.data.data.activities || []
        const buildRange = (days: number) => {
          const out: ActivityCalendarItem[] = []
          const today = new Date()
          for (let i = days - 1; i >= 0; i--) {
            const d = new Date(today)
            d.setDate(today.getDate() - i)
            const yyyy = d.getFullYear()
            const mm = String(d.getMonth() + 1).padStart(2, '0')
            const dd = String(d.getDate()).padStart(2, '0')
            out.push({ date: `${yyyy}-${mm}-${dd}`, count: 0 })
          }
          return out
        }
        const rangeDays = 90
        const range = buildRange(rangeDays)
        if (apiItems.length === 0) {
          activities.value = range
        } else {
          const byDate = new Map<string, number>()
          for (const item of apiItems) {
            byDate.set(item.date, item.count || 0)
          }
          activities.value = range.map(d => ({
            date: d.date,
            count: byDate.get(d.date) ?? 0
          }))
        }
      }
    } catch (err) {
      console.error(err)
    } finally {
      loadingActivities.value = false
    }
  }

  const fetchPublicPrivateNotes = async () => {
    loadingPublicPrivateNotes.value = true
    try {
      const res = await getPublicPrivateNotes(1, 10)
      if (res.data?.code === 200 && res.data.data) {
        publicPrivateNotes.value = res.data.data.list || []
      }
    } catch (err) {
      console.error(err)
    } finally {
      loadingPublicPrivateNotes.value = false
    }
  }

  const fetchSharedNotes = async () => {
    loadingSharedNotes.value = true
    try {
      const res = await getSharedNotes(1, 10)
      if (res.data?.code === 200 && res.data.data) {
        sharedNotes.value = res.data.data.list || []
      }
    } catch (err) {
      console.error(err)
    } finally {
      loadingSharedNotes.value = false
    }
  }

  const fetchLearnedSubjects = async () => {
    loadingLearnedSubjects.value = true
    try {
      const res = await getLearnedSubjects()
      if (res.data?.code === 200 && res.data.data) {
        learnedSubjects.value = res.data.data.list || []
      }
    } catch (err) {
      console.error(err)
    } finally {
      loadingLearnedSubjects.value = false
    }
  }

  const refreshAll = () => {
    fetchUserInfo()
    fetchActivities()
    fetchPublicPrivateNotes()
    fetchSharedNotes()
    fetchLearnedSubjects()
  }

  onMounted(() => {
    refreshAll()
  })

  return {
    userInfo,
    activities,
    publicPrivateNotes,
    sharedNotes,
    learnedSubjects,
    loadingUserInfo,
    loadingActivities,
    loadingPublicPrivateNotes,
    loadingSharedNotes,
    loadingLearnedSubjects,
    errorUserInfo,
    refreshAll
  }
}
