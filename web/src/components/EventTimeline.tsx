import { useEffect, useRef, type ReactNode } from 'react'
import {
  FILTERS,
  countByFilter,
  filterRecords,
  formatTraceTime,
  recordTone,
} from '../lib/format'
import type { Filter } from '../lib/format'
import type { TraceRecord } from '../types'

interface Props {
  records: TraceRecord[]
  index: number
  filter: Filter
  onFilterChange: (filter: Filter) => void
  onSelect: (index: number) => void
  controls: ReactNode
}

/**
 * The timeline is a view over the engine's append-only event log. Filtering
 * only hides rows; it never changes their recorded order or timestamps.
 */
export function EventTimeline({
  records,
  index,
  filter,
  onFilterChange,
  onSelect,
  controls,
}: Props) {
  const scrollRef = useRef<HTMLDivElement>(null)
  const visible = filterRecords(records, filter)

  // Keep the replay cursor visible as it advances.
  useEffect(() => {
    const container = scrollRef.current
    if (!container) {
      return
    }
    const row = container.querySelector<HTMLElement>(
      `[data-record-index="${index}"]`,
    )
    row?.scrollIntoView({ block: 'nearest' })
  }, [index, filter])

  return (
    <section className="timeline" aria-label="Event log">
      <div className="timeline-head">
        <div className="timeline-title">
          <span>Event log</span>
          <b>{records.length}</b>
        </div>

        <div className="filters" role="group" aria-label="Filter events">
          {FILTERS.map((entry) => (
            <button
              key={entry.id}
              type="button"
              className={`filter${entry.id === filter ? ' is-active' : ''}`}
              aria-pressed={entry.id === filter}
              onClick={() => onFilterChange(entry.id)}
            >
              {entry.label}
              <span className="filter-count">
                {countByFilter(records, entry.id)}
              </span>
            </button>
          ))}
        </div>

        <span className="spacer" />
        {controls}
      </div>

      <div className="timeline-scroll" ref={scrollRef}>
        {visible.length === 0 ? (
          <p className="timeline-empty">
            No events match this filter. Choose “All” to see the full log.
          </p>
        ) : (
          visible.map((record) => {
            const tone = recordTone(record)
            const selected = record.index === index
            return (
              <button
                key={record.index}
                type="button"
                data-record-index={record.index}
                className={`event-row${selected ? ' is-selected' : ''}${
                  tone === 'failure' ? ' is-failure' : ''
                }${tone === 'success' ? ' is-success' : ''}`}
                aria-current={selected ? 'true' : undefined}
                onClick={() => onSelect(record.index)}
              >
                <span className="event-time">
                  {formatTraceTime(record.atMs)}
                </span>
                <span className="event-kind">{record.kind}</span>
                <span className="event-node">{record.nodeId ?? '—'}</span>
                <span className="event-packet">{record.packetId ?? '—'}</span>
                <span className="event-summary">{record.summary}</span>
              </button>
            )
          })
        )}
      </div>
    </section>
  )
}
