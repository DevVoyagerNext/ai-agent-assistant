<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { Graph } from '@antv/x6'
import { useRouter } from 'vue-router'

import type { Node as NodeData } from '../types/userProgress'

const props = defineProps<{
  nodes: NodeData[]
}>()

const container = ref<HTMLElement | null>(null)
let graph: Graph | null = null
const router = useRouter()

const getStatusStyles = (status: NodeData['status']) => {
  switch (status) {
    case 'completed':
      return {
        fill: '#4ade80', // green-400
        stroke: '#22c55e',
        labelColor: '#ffffff'
      }
    case 'learning':
      return {
        fill: '#60a5fa', // blue-400
        stroke: '#3b82f6',
        labelColor: '#ffffff',
        className: 'breathing-node'
      }
    case 'unstarted':
    default:
      return {
        fill: '#e5e7eb', // gray-200
        stroke: '#9ca3af',
        labelColor: '#6b7280'
      }
  }
}

const initGraph = () => {
  if (!container.value) return

  graph = new Graph({
    container: container.value,
    autoResize: true,
    background: {
      color: '#f8fafc',
    },
    grid: {
      visible: true,
      type: 'doubleMesh',
      args: [
        { color: '#eee', thickness: 1 },
        { color: '#ddd', thickness: 1, factor: 4 },
      ],
    },
    interacting: {
      nodeMovable: true,
    },
  })

  renderData()

  graph.on('node:click', ({ node }) => {
    const data = node.getData() as NodeData
    if (data.status !== 'unstarted') {
      router.push(`/study/${data.id}`)
    } else {
      alert('该知识点尚未解锁，请先完成前置内容')
    }
  })
}

const renderData = () => {
  if (!graph) return
  graph.clearCells()

  const cells: any[] = []

  // Create Nodes
  props.nodes.forEach((node) => {
    const styles = getStatusStyles(node.status)
    const x6Node = graph!.createNode({
      id: node.id,
      shape: 'rect',
      x: node.x || 100,
      y: node.y || 100,
      width: 120,
      height: 45,
      label: node.name,
      attrs: {
        body: {
          fill: styles.fill,
          stroke: styles.stroke,
          strokeWidth: 2,
          rx: 10,
          ry: 10,
          class: styles.className
        },
        label: {
          fill: styles.labelColor,
          fontSize: 14,
          fontWeight: 'bold',
        },
      },
      data: node
    })
    cells.push(x6Node)
  })

  // Create Edges
  props.nodes.forEach((node) => {
    if (node.parentId) {
      const edge = graph!.createEdge({
        source: node.parentId,
        target: node.id,
        attrs: {
          line: {
            stroke: '#94a3b8',
            strokeWidth: 2,
            targetMarker: 'classic',
          },
        },
        zIndex: 0,
      })
      cells.push(edge)
    }
  })

  graph.addCell(cells)
  graph.centerContent()
}

onMounted(() => {
  initGraph()
})

watch(() => props.nodes, () => {
  renderData()
}, { deep: true })
</script>

<template>
  <div class="knowledge-tree-container">
    <div ref="container" class="x6-container"></div>
  </div>
</template>

<style scoped>
.knowledge-tree-container {
  width: 100%;
  height: 100%;
  position: relative;
}

.x6-container {
  width: 100%;
  height: 100%;
}

:deep(.breathing-node) {
  animation: breathing 2s ease-in-out infinite;
}

@keyframes breathing {
  0% {
    filter: drop-shadow(0 0 2px rgba(96, 165, 250, 0.5));
    opacity: 0.8;
  }
  50% {
    filter: drop-shadow(0 0 10px rgba(96, 165, 250, 0.8));
    opacity: 1;
  }
  100% {
    filter: drop-shadow(0 0 2px rgba(96, 165, 250, 0.5));
    opacity: 0.8;
  }
}
</style>
