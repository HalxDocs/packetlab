import { useMemo } from 'react'
import { formatTraceTime } from '../lib/format'
import { layoutTopology } from '../lib/layout'
import type { TraceBundle, TraceRecord } from '../types'

interface Props {
  bundle: TraceBundle
  record: TraceRecord | undefined
  selectedNodeId: string | null
  onSelectNode: (id: string | null) => void
}

type MarkerTone = 'active' | 'success' | 'failure'

interface PacketMarker {
  x: number
  y: number
  tone: MarkerTone
  label: string
  labelY: number
  anchor: 'start' | 'middle' | 'end'
}

function delayLabel(ms: number): string {
  return Number.isInteger(ms) ? `${ms}ms` : `${ms.toFixed(3)}ms`
}

/**
 * A schematic of the topology at the current replay position. Nothing here
 * animates: the packet token sits where the selected record says the packet is.
 */
export function TopologyCanvas({
  bundle,
  record,
  selectedNodeId,
  onSelectNode,
}: Props) {
  const layout = useMemo(
    () => layoutTopology(bundle.topology),
    [bundle.topology],
  )

  const activeLink = record?.linkId
    ? layout.links.find((line) => line.id === record.linkId)
    : undefined
  const activeNode = record?.nodeId
    ? layout.nodes.find((box) => box.id === record.nodeId)
    : undefined

  let marker: PacketMarker | null = null
  if (record?.packetId) {
    const tone: MarkerTone =
      record.kind === 'packet.dropped'
        ? 'failure'
        : record.kind === 'packet.delivered' || record.kind === 'ping.reply'
          ? 'success'
          : 'active'

    if (record.kind === 'link.transmit' && activeLink) {
      marker = {
        x: (activeLink.x1 + activeLink.x2) / 2,
        y: (activeLink.y1 + activeLink.y2) / 2,
        tone,
        label: `${record.packetId} · in flight`,
        labelY: (activeLink.y1 + activeLink.y2) / 2 + 17,
        anchor: 'middle',
      }
    } else if (activeNode) {
      marker = {
        x: activeNode.x + activeNode.width - 18,
        y: activeNode.y + 18,
        tone,
        label: record.packetId,
        labelY: activeNode.y + 21,
        anchor: 'end',
      }
    }
  }

  const toggleNode = (id: string) => {
    onSelectNode(selectedNodeId === id ? null : id)
  }

  return (
    <section className="stage" aria-label="Topology">
      <div className="stage-bar">
        <div className="stage-label">
          <span>Topology</span>
          <strong>
            {bundle.topology.nodes.length} nodes ·{' '}
            {bundle.topology.links.length} links
          </strong>
        </div>
        <span className="stage-hint">engine output · replay only</span>
      </div>

      <div className="canvas-wrap">
        <svg
          className="canvas-svg"
          viewBox={`0 0 ${layout.width} ${layout.height}`}
          role="img"
          aria-label="Network topology at the current replay position"
        >
          {layout.links.map((line) => {
            const isActive = activeLink?.id === line.id
            const isFailure =
              isActive && record?.kind === 'packet.dropped' && Boolean(record.reason)
            const midX = (line.x1 + line.x2) / 2
            const midY = (line.y1 + line.y2) / 2
            return (
              <g key={line.id}>
                <title>
                  {line.id}: {line.link.a} ↔ {line.link.b}
                </title>
                <line
                  className={`topo-link${line.link.up ? '' : ' is-down'}${
                    isActive && !isFailure ? ' is-active' : ''
                  }${isFailure ? ' is-failure' : ''}`}
                  x1={line.x1}
                  y1={line.y1}
                  x2={line.x2}
                  y2={line.y2}
                />
                <text
                  className={`topo-link-label${
                    isActive && !isFailure ? ' is-active' : ''
                  }${isFailure ? ' is-failure' : ''}`}
                  x={midX}
                  y={midY - 7}
                  textAnchor="middle"
                >
                  {line.id} · {delayLabel(line.link.delayMs)}
                </text>
              </g>
            )
          })}

          {layout.nodes.map((box) => {
            const node = box.node
            const interfaces = node.interfaces ?? []
            const isActive = record?.nodeId === box.id
            const isSelected = selectedNodeId === box.id
            const isFailure =
              Boolean(record?.reason) && isActive && !record?.linkId
            return (
              <g
                key={box.id}
                className={`topo-node${isActive ? ' is-active' : ''}${
                  isSelected ? ' is-selected' : ''
                }${node.down ? ' is-down' : ''}${
                  isFailure ? ' is-failure' : ''
                }`}
                transform={`translate(${box.x}, ${box.y})`}
                role="button"
                tabIndex={0}
                aria-label={`Inspect ${node.kind} ${node.name ?? node.id}`}
                onClick={() => toggleNode(box.id)}
                onKeyDown={(event) => {
                  if (event.key === 'Enter' || event.key === ' ') {
                    event.preventDefault()
                    toggleNode(box.id)
                  }
                }}
              >
                <rect
                  className="topo-node-box"
                  width={box.width}
                  height={box.height}
                  rx={3}
                />
                <text className="topo-kind" x={14} y={20}>
                  {node.kind}
                </text>
                <text className="topo-name" x={14} y={41}>
                  {node.name ?? node.id}
                </text>
                {interfaces.slice(0, 2).map((iface, i) => (
                  <text
                    key={iface.id}
                    className="topo-ip"
                    x={14}
                    y={59 + i * 14}
                  >
                    {iface.address}
                  </text>
                ))}
                <text className="topo-if" x={14} y={box.height - 9}>
                  {node.id}
                </text>
                {node.down ? (
                  <text
                    className="node-down-label"
                    x={box.width - 14}
                    y={20}
                    textAnchor="end"
                  >
                    DOWN
                  </text>
                ) : null}
              </g>
            )
          })}

          {marker ? (
            <g>
              <circle
                className={`packet-dot-halo is-${marker.tone}`}
                cx={marker.x}
                cy={marker.y}
                r={7}
              />
              <circle
                className={`packet-dot is-${marker.tone}`}
                cx={marker.x}
                cy={marker.y}
                r={3.5}
              />
              <text
                className={`packet-label is-${marker.tone}`}
                x={marker.x + (marker.anchor === 'end' ? -10 : 0)}
                y={marker.labelY}
                textAnchor={marker.anchor}
              >
                {marker.label}
              </text>
            </g>
          ) : null}
        </svg>
      </div>

      <div className="stage-foot">
        <span>
          replay t <b>{formatTraceTime(record?.atMs ?? 0)}</b>
        </span>
        <span>
          event <b>{record ? record.kind : 'none'}</b>
        </span>
        <span>
          packet <b>{record?.packetId ?? '—'}</b>
        </span>
        <span className="spacer" />
        <span>
          {selectedNodeId ? (
            <>
              inspecting <b>{selectedNodeId}</b>
            </>
          ) : (
            'click a node to inspect'
          )}
        </span>
      </div>
    </section>
  )
}
