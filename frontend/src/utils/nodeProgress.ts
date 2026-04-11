export type NodeProgressStatus = 'unstarted' | 'learning' | 'completed'

const STATUS_ALIAS_MAP: Record<string, NodeProgressStatus> = {
  unstarted: 'unstarted',
  learning: 'learning',
  completed: 'completed',
  '0': 'unstarted',
  '1': 'learning',
  '2': 'completed',
  '未开始': 'unstarted',
  '学习中': 'learning',
  '已掌握': 'completed',
  '已完成': 'completed'
}

export const normalizeNodeProgressStatus = (
  status: unknown,
  fallback: NodeProgressStatus = 'unstarted'
): NodeProgressStatus => {
  const normalizedFallback = STATUS_ALIAS_MAP[String(fallback).trim().toLowerCase()] || 'unstarted'

  if (status === null || status === undefined) {
    return normalizedFallback
  }

  const rawValue = String(status).trim()
  if (!rawValue) {
    return normalizedFallback
  }

  return STATUS_ALIAS_MAP[rawValue.toLowerCase()] || STATUS_ALIAS_MAP[rawValue] || normalizedFallback
}

export const getNodeProgressStatusLabel = (status: unknown) => {
  const normalizedStatus = normalizeNodeProgressStatus(status)

  if (normalizedStatus === 'completed') return '已掌握'
  if (normalizedStatus === 'learning') return '学习中'
  return '未开始'
}
