import { describe, expect, it } from 'vitest'
import { layoutTopology } from './layout'
import type { TraceTopology } from '../types'

const topology: TraceTopology = {
  nodes: [
    {
      id: 'host-a',
      kind: 'host',
      interfaces: [
        {
          id: 'host-a:eth0',
          node: 'host-a',
          address: '10.0.1.10/24',
          subnet: '10.0.1.0/24',
        },
      ],
    },
    {
      id: 'router-r1',
      kind: 'router',
      interfaces: [
        {
          id: 'router-r1:eth0',
          node: 'router-r1',
          address: '10.0.1.1/24',
          subnet: '10.0.1.0/24',
        },
        {
          id: 'router-r1:eth1',
          node: 'router-r1',
          address: '10.0.2.1/24',
          subnet: '10.0.2.0/24',
        },
      ],
    },
    {
      id: 'host-b',
      kind: 'host',
      interfaces: [
        {
          id: 'host-b:eth0',
          node: 'host-b',
          address: '10.0.2.20/24',
          subnet: '10.0.2.0/24',
        },
      ],
    },
  ],
  links: [
    { id: 'link-a', a: 'host-a:eth0', b: 'router-r1:eth0', delayMs: 1, up: true },
    { id: 'link-b', a: 'router-r1:eth1', b: 'host-b:eth0', delayMs: 1, up: true },
  ],
}

describe('layoutTopology', () => {
  it('places one box per node and one line per link', () => {
    const layout = layoutTopology(topology)
    expect(layout.nodes).toHaveLength(3)
    expect(layout.links).toHaveLength(2)
    expect(layout.width).toBeGreaterThan(0)
    expect(layout.height).toBeGreaterThan(0)
  })

  it('is deterministic', () => {
    const first = JSON.stringify(layoutTopology(topology))
    const second = JSON.stringify(layoutTopology(topology))
    expect(first).toBe(second)
  })

  it('orders columns by BFS depth', () => {
    const layout = layoutTopology(topology)
    const x = (id: string) =>
      layout.nodes.find((node) => node.id === id)?.x ?? -1
    expect(x('host-a')).toBeLessThan(x('router-r1'))
    expect(x('router-r1')).toBeLessThan(x('host-b'))
  })

  it('never overlaps nodes', () => {
    const layout = layoutTopology(topology)
    for (let i = 0; i < layout.nodes.length; i += 1) {
      for (let j = i + 1; j < layout.nodes.length; j += 1) {
        const a = layout.nodes[i]
        const b = layout.nodes[j]
        const overlaps =
          a.x < b.x + b.width &&
          a.x + a.width > b.x &&
          a.y < b.y + b.height &&
          a.y + a.height > b.y
        expect(overlaps).toBe(false)
      }
    }
  })

  it('anchors links on box edges', () => {
    const layout = layoutTopology(topology)
    for (const line of layout.links) {
      for (const value of [line.x1, line.y1, line.x2, line.y2]) {
        expect(Number.isFinite(value)).toBe(true)
      }
      expect(line.x1).not.toBe(line.x2)
    }
  })
})
