import type { ReactNode } from 'react'
import { formatTraceTime, packetJourney, reasonLabel } from '../lib/format'
import type { TraceBundle, TraceRecord } from '../types'

interface Props {
  bundle: TraceBundle
  record: TraceRecord | undefined
  records: TraceRecord[]
  nodeId?: string
  onSelect: (index: number) => void
}

function Row({ label, children }: { label: string; children: ReactNode }) {
  return (
    <>
      <span className="kv-key">{label}</span>
      <span className="kv-val">{children}</span>
    </>
  )
}

/**
 * The inspector explains the current record: its inputs, the packet's whole
 * journey, and the static tables of the node involved. Every value comes from
 * the trace bundle.
 */
export function PacketInspector({
  bundle,
  record,
  records,
  nodeId,
  onSelect,
}: Props) {
  if (!record) {
    return (
      <aside className="inspector" aria-label="Inspector">
        <div className="empty-state">
          <strong>No event selected.</strong>
          <br />
          Choose a row in the event log or press play.
        </div>
      </aside>
    )
  }

  const journey = packetJourney(records, record.packetId)
  const node = nodeId
    ? bundle.topology.nodes.find((candidate) => candidate.id === nodeId)
    : undefined
  const details = Object.entries(record.details ?? {}).sort(([a], [b]) =>
    a.localeCompare(b),
  )

  return (
    <aside className="inspector" aria-label="Inspector">
      <div className="inspector-head">
        <div className="inspector-kicker">
          <span>
            Event {record.index + 1} / {records.length}
          </span>
          <span>·</span>
          <span>{record.kind}</span>
        </div>
        <h2 className="inspector-title">{record.summary}</h2>
        <div className="inspector-sub">
          t = {formatTraceTime(record.atMs)}
          {record.nodeId ? ` · ${record.nodeId}` : ''}
        </div>
        {record.reason ? (
          <p style={{ margin: '8px 0 0' }}>
            <span className="reason">
              <span className="reason-dot" />
              {reasonLabel(record.reason)}
            </span>
          </p>
        ) : null}
      </div>

      <section className="inspector-section">
        <h3 className="section-title">Event</h3>
        <div className="kv">
          <Row label="time">{formatTraceTime(record.atMs)}</Row>
          <Row label="kind">{record.kind}</Row>
          {record.nodeId ? <Row label="node">{record.nodeId}</Row> : null}
          {record.linkId ? <Row label="link">{record.linkId}</Row> : null}
          {record.packetId ? (
            <Row label="packet">{record.packetId}</Row>
          ) : null}
          {record.reason ? (
            <Row label="reason">
              <span className="mono">{record.reason}</span>
            </Row>
          ) : null}
        </div>
      </section>

      {journey.length > 1 ? (
        <section className="inspector-section">
          <h3 className="section-title">
            <span>Packet {record.packetId}</span>
            <span className="muted">{journey.length} events</span>
          </h3>
          <div className="journey">
            {journey.map((item) => (
              <button
                key={item.index}
                type="button"
                className={`journey-item${
                  item.index === record.index ? ' is-current' : ''
                }${item.reason ? ' is-failure' : ''}`}
                onClick={() => onSelect(item.index)}
              >
                <span className="journey-time">
                  {formatTraceTime(item.atMs)}
                </span>
                <span className="journey-text">{item.summary}</span>
              </button>
            ))}
          </div>
        </section>
      ) : null}

      {node ? (
        <section className="inspector-section">
          <h3 className="section-title">
            <span>Node · {node.id}</span>
            {node.down ? (
              <span className="chip chip--danger">
                <span className="chip-dot" />
                down
              </span>
            ) : (
              <span className="muted">{node.kind}</span>
            )}
          </h3>
          <div className="kv">
            <Row label="name">{node.name ?? node.id}</Row>
            <Row label="kind">{node.kind}</Row>
            <Row label="status">{node.down ? 'down' : 'up'}</Row>
            {(node.interfaces ?? []).map((iface) => (
              <Row key={iface.id} label={iface.id}>
                {iface.address}
                <span className="muted"> · {iface.subnet}</span>
              </Row>
            ))}
          </div>

          <h3 className="section-title" style={{ marginTop: '14px' }}>
            <span>Routes</span>
            <span className="muted">{node.routes?.length ?? 0}</span>
          </h3>
          {node.routes && node.routes.length > 0 ? (
            <table className="route-table">
              <thead>
                <tr>
                  <th>Network</th>
                  <th>Next hop</th>
                  <th>Interface</th>
                </tr>
              </thead>
              <tbody>
                {node.routes.map((route, i) => (
                  <tr key={`${route.network}-${i}`}>
                    <td title={route.metric ? `metric ${route.metric}` : undefined}>
                      {route.network}
                    </td>
                    <td>{route.nextHop ?? 'direct'}</td>
                    <td>{route.interface}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <p className="muted" style={{ margin: 0, fontSize: '11px' }}>
              No static routes.
            </p>
          )}
        </section>
      ) : null}

      <section className="inspector-section">
        <h3 className="section-title">Details</h3>
        {details.length > 0 ? (
          <div className="kv">
            {details.map(([key, value]) => (
              <Row key={key} label={key}>
                {value}
              </Row>
            ))}
          </div>
        ) : (
          <p className="muted" style={{ margin: 0, fontSize: '11px' }}>
            No additional details.
          </p>
        )}
      </section>
    </aside>
  )
}
