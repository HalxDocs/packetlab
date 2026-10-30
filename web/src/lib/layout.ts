import type { TraceLink, TraceNode, TraceTopology } from '../types'

export interface NodeBox {
  id: string
  node: TraceNode
  x: number
  y: number
  width: number
  height: number
}

export interface LinkLine {
  id: string
  link: TraceLink
  aNode: string
  bNode: string
  x1: number
  y1: number
  x2: number
  y2: number
}

export interface TopologyLayout {
  width: number
  height: number
  nodes: NodeBox[]
  links: LinkLine[]
}

const NODE_WIDTH = 168
const NODE_HEIGHT = 100
const COL_GAP = 76
const ROW_GAP = 56
const MARGIN = 32

/**
 * Lay a topology out left to right by BFS depth. The result is deterministic:
 * nodes are sorted by ID, neighbours by link ID, and rows are centred within
 * their column. The UI never invents positions from animation state.
 */
export function layoutTopology(topology: TraceTopology): TopologyLayout {
  const nodes = [...topology.nodes].sort((a, b) => a.id.localeCompare(b.id))
  const byID = new Map(nodes.map((node) => [node.id, node]))

  const interfaceNode = new Map<string, string>()
  for (const node of nodes) {
    for (const iface of node.interfaces) {
      interfaceNode.set(iface.id, node.id)
    }
  }

  const adjacency = new Map<string, string[]>()
  for (const node of nodes) {
    adjacency.set(node.id, [])
  }
  const links = [...topology.links].sort((a, b) => a.id.localeCompare(b.id))
  for (const link of links) {
    const a = interfaceNode.get(link.a)
    const b = interfaceNode.get(link.b)
    if (!a || !b || a === b) {
      continue
    }
    adjacency.get(a)?.push(b)
    adjacency.get(b)?.push(a)
  }
  for (const neighbours of adjacency.values()) {
    neighbours.sort((a, b) => a.localeCompare(b))
  }

  const depth = new Map<string, number>()
  for (const start of nodes.map((node) => node.id)) {
    if (depth.has(start)) {
      continue
    }
    depth.set(start, 0)
    const queue = [start]
    while (queue.length > 0) {
      const current = queue.shift() as string
      const currentDepth = depth.get(current) ?? 0
      for (const next of adjacency.get(current) ?? []) {
        if (!depth.has(next)) {
          depth.set(next, currentDepth + 1)
          queue.push(next)
        }
      }
    }
  }

  const columns = new Map<number, string[]>()
  for (const node of nodes) {
    const d = depth.get(node.id) ?? 0
    const column = columns.get(d) ?? []
    column.push(node.id)
    columns.set(d, column)
  }
  const depths = [...columns.keys()].sort((a, b) => a - b)
  const maxDepth = depths.length > 0 ? Math.max(...depths) : 0
  const maxRows = Math.max(...depths.map((d) => columns.get(d)?.length ?? 0), 1)

  const width = MARGIN * 2 + (maxDepth + 1) * NODE_WIDTH + maxDepth * COL_GAP
  const height =
    MARGIN * 2 + maxRows * NODE_HEIGHT + Math.max(0, maxRows - 1) * ROW_GAP

  const boxes: NodeBox[] = []
  for (const d of depths) {
    const ids = columns.get(d) ?? []
    const offset =
      ((maxRows - ids.length) * (NODE_HEIGHT + ROW_GAP)) / 2
    ids.forEach((id, row) => {
      const node = byID.get(id)
      if (!node) {
        return
      }
      boxes.push({
        id,
        node,
        x: MARGIN + d * (NODE_WIDTH + COL_GAP),
        y: MARGIN + offset + row * (NODE_HEIGHT + ROW_GAP),
        width: NODE_WIDTH,
        height: NODE_HEIGHT,
      })
    })
  }

  const boxByID = new Map(boxes.map((box) => [box.id, box]))
  const lines: LinkLine[] = []
  for (const link of links) {
    const a = interfaceNode.get(link.a)
    const b = interfaceNode.get(link.b)
    if (!a || !b) {
      continue
    }
    const from = boxByID.get(a)
    const to = boxByID.get(b)
    if (!from || !to) {
      continue
    }
    const anchors = anchor(from, to)
    lines.push({ id: link.id, link, aNode: a, bNode: b, ...anchors })
  }

  return { width, height, nodes: boxes, links: lines }
}

function anchor(from: NodeBox, to: NodeBox) {
  const fromCenterY = from.y + from.height / 2
  const toCenterY = to.y + to.height / 2

  if (to.x >= from.x + from.width) {
    return {
      x1: from.x + from.width,
      y1: fromCenterY,
      x2: to.x,
      y2: toCenterY,
    }
  }
  if (to.x + to.width <= from.x) {
    return {
      x1: from.x,
      y1: fromCenterY,
      x2: to.x + to.width,
      y2: toCenterY,
    }
  }
  if (to.y > from.y) {
    return {
      x1: from.x + from.width / 2,
      y1: from.y + from.height,
      x2: to.x + to.width / 2,
      y2: to.y,
    }
  }
  return {
    x1: from.x + from.width / 2,
    y1: from.y,
    x2: to.x + to.width / 2,
    y2: to.y + to.height,
  }
}
