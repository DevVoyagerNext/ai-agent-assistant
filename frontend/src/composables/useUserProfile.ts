import { ref, onMounted } from 'vue'
import { 
  getUserInfo, 
  getUserActivitiesCalendar, 
  getPublicPrivateNotes, 
  getSharedNotes, 
  getLearnedSubjects,
  getUserCollectFolders,
  getUserLikedSubjects,
  getUserLearningSubjects,
  getUserCompletedSubjects,
  getUserRecentSubjects
} from '../api/user'
import type { 
  UserInfoRes, 
  ActivityCalendarItem, 
  PublicPrivateNoteItem, 
  SharedNoteItem, 
  LearnedSubjectItem,
  CollectFolderRes,
  UserSubjectProgressRes
} from '../types/user'
import type { Subject } from '../types/subject'

export const useUserProfile = () => {
  // State
  const userInfo = ref<UserInfoRes | null>(null)
  const activities = ref<ActivityCalendarItem[]>([])
  const publicPrivateNotes = ref<PublicPrivateNoteItem[]>([])
  const sharedNotes = ref<SharedNoteItem[]>([])
  const learnedSubjects = ref<LearnedSubjectItem[]>([])
  
  // 新增 State
  const collectFolders = ref<CollectFolderRes[]>([])
  const likedSubjects = ref<Subject[]>([])
  const learningSubjects = ref<UserSubjectProgressRes[]>([])
  const completedSubjects = ref<UserSubjectProgressRes[]>([])
  const recentSubjects = ref<UserSubjectProgressRes[]>([])

  // Loading states
  const loadingUserInfo = ref(true)
  const loadingActivities = ref(true)
  const loadingPublicPrivateNotes = ref(true)
  const loadingSharedNotes = ref(true)
  const loadingLearnedSubjects = ref(true)
  
  // 新增 Loading states
  const loadingCollectFolders = ref(true)
  const loadingLikedSubjects = ref(true)
  const loadingLearningSubjects = ref(true)
  const loadingCompletedSubjects = ref(true)
  const loadingRecentSubjects = ref(true)

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

  // ----------- 新增 Fetchers -----------
  const fetchCollectFolders = async () => {
    loadingCollectFolders.value = true
    try {
      const res = await getUserCollectFolders()
      if (res.data?.code === 200 && res.data.data) {
        collectFolders.value = res.data.data || []
      }
    } catch (err) {
      console.error(err)
    } finally {
      loadingCollectFolders.value = false
    }
  }

  const fetchLikedSubjects = async () => {
    loadingLikedSubjects.value = true
    try {
      const res = await getUserLikedSubjects()
      if (res.data?.code === 200 && res.data.data) {
        likedSubjects.value = res.data.data || []
      }
    } catch (err) {
      console.error(err)
    } finally {
      loadingLikedSubjects.value = false
    }
  }

  const fetchLearningSubjects = async () => {
    loadingLearningSubjects.value = true
    try {
      const res = await getUserLearningSubjects()
      if (res.data?.code === 200 && res.data.data) {
        learningSubjects.value = res.data.data || []
      }
    } catch (err) {
      console.error(err)
    } finally {
      loadingLearningSubjects.value = false
    }
  }

  const fetchCompletedSubjects = async () => {
    loadingCompletedSubjects.value = true
    try {
      const res = await getUserCompletedSubjects()
      if (res.data?.code === 200 && res.data.data) {
        completedSubjects.value = res.data.data || []
      }
    } catch (err) {
      console.error(err)
    } finally {
      loadingCompletedSubjects.value = false
    }
  }

  const fetchRecentSubjects = async () => {
    loadingRecentSubjects.value = true
    try {
      const res = await getUserRecentSubjects(1, 10)
      if (res.data?.code === 200 && res.data.data) {
        recentSubjects.value = res.data.data.list || []
      }
    } catch (err) {
      console.error(err)
    } finally {
      loadingRecentSubjects.value = false
    }
  }

  const refreshAll = () => {
    fetchUserInfo()
    fetchActivities()
    fetchPublicPrivateNotes()
    fetchSharedNotes()
    fetchLearnedSubjects()
    fetchCollectFolders()
    fetchLikedSubjects()
    fetchLearningSubjects()
    fetchCompletedSubjects()
    fetchRecentSubjects()
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
    collectFolders,
    likedSubjects,
    learningSubjects,
    completedSubjects,
    recentSubjects,

    loadingUserInfo,
    loadingActivities,
    loadingPublicPrivateNotes,
    loadingSharedNotes,
    loadingLearnedSubjects,
    loadingCollectFolders,
    loadingLikedSubjects,
    loadingLearningSubjects,
    loadingCompletedSubjects,
    loadingRecentSubjects,
    
    errorUserInfo,
    refreshAll
  }
}
