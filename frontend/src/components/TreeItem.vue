<script setup lang="ts">
import { type PropType } from 'vue'
import { 
  ChevronRight, ChevronDown, Loader2 
} from 'lucide-vue-next'
import type { SubjectNode } from '../types/node'

interface TreeNode extends SubjectNode {
  expanded?: boolean;
  children?: TreeNode[];
  loadingChildren?: boolean;
}

const props = defineProps({
  node: { type: Object as PropType<TreeNode>, required: true },
  level: { type: Number, required: true },
  activeId: { type: [Number, Object] as PropType<number | null>, required: false, default: null }
})

const emit = defineEmits(['nodeClick', 'toggleExpand'])

const handleToggleExpand = (e: MouseEvent) => {
  e.stopPropagation()
  emit('toggleExpand', props.node)
}

const handleNodeClick = () => {
  emit('nodeClick', props.node)
}
</script>

<template>
  <li class="tree-item">
    <div 
      class="node-content" 
      :class="{ active: activeId === node.id, 'is-parent': node.isLeaf === 0 }"
      :style="{ paddingLeft: `${16 + level * 18}px` }"
      @click="handleNodeClick"
    >
      <span v-if="node.isLeaf === 0" class="expand-icon" @click="handleToggleExpand">
        <ChevronDown v-if="node.expanded" :size="14" />
        <ChevronRight v-else :size="14" />
      </span>
      
      <span class="node-name">{{ node.name }}</span>
    </div>

    <ul v-if="node.expanded && node.children && node.children.length > 0" class="tree-list sub-list">
      <TreeItem
        v-for="child in node.children"
        :key="child.id"
        :node="child"
        :level="level + 1"
        :active-id="activeId"
        @node-click="(n) => emit('nodeClick', n)"
        @toggle-expand="(n) => emit('toggleExpand', n)"
      />
    </ul>

    <div v-if="node.loadingChildren" class="loading-sub" :style="{ paddingLeft: `${16 + (level + 1) * 18}px` }">
      <Loader2 class="spin" :size="14" />
    </div>
  </li>
</template>

<style scoped>
.tree-item { list-style: none; }
.tree-list { list-style: none; padding: 0; margin: 0; }
.sub-list { background: rgba(248, 250, 252, 0.5); }

.node-content {
  display: flex;
  align-items: center;
  padding: 10px 16px;
  cursor: pointer;
  transition: all 0.2s;
  color: #475569;
  font-size: 14px;
  border-left: 3px solid transparent;
}

.node-content:hover { background: #f1f5f9; color: #0f172a; }
.node-content.active {
  background: #eff6ff;
  color: #2563eb;
  font-weight: 600;
  border-left-color: #3b82f6;
}

.expand-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  color: #94a3b8;
  margin-right: 6px;
  flex-shrink: 0;
}

.node-name { flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.loading-sub { padding: 8px 16px; display: flex; color: #3b82f6; }

.spin { animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
</style>
