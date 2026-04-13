export const NOTE_MAX_LENGTH = 1000

export const validateNoteContent = (noteContent: unknown) => {
  const content = String(noteContent ?? '')

  if (!content.trim()) {
    throw new Error('笔记内容不能为空')
  }

  if (content.length > NOTE_MAX_LENGTH) {
    throw new Error(`笔记内容不能超过 ${NOTE_MAX_LENGTH} 字符`)
  }

  if (content.includes('<') || content.includes('>')) {
    throw new Error('笔记内容包含非法字符')
  }

  return content
}
