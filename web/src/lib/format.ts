import type { TraceRecord } from '../types'

/** Render virtual milliseconds the way the engine trace does. */
export function formatTraceTime(ms: number): string {
  if (!Number.isFinite(ms)) {
    return '—'
  }
  if (ms < 0) {
    return `-${formatTraceTime(-ms)}`
  }
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(3)}s`
  }
  return `${ms.toFixed(3)}ms`
}

const REASONS: Record<string, string> = {
  no_route: 'no route to destination',
  link_down: 'link is down',
  ttl_expired: 'TTL expired',
  no_interface: 'route has no usable interface',
  unknown_destination: 'destination address is not local',
  node_down: 'node is down',
  queue_full: 'queue is full',
  link_loss: 'link loss model dropped the packet',
}

export function reasonLabel(reason: string | undefined): string {
  if (!reason) {
    return ''
  }
  return REASONS[reason] ?? reason
}

export type Filter = 'all' | 'packets' | 'routing' | 'links' | 'failures'

export const FILTERS: { id: Filter; label: string }[] = [
  { id: 'all', label: 'All' },
  { id: 'packets', label: 'Packets' },
  { id: 'routing', label: 'Routing' },
  { id: 'links', label: 'Links' },
  { id: 'failures', label: 'Failures' },
]

export function matchesFilter(record: TraceRecord, filter: Filter): boolean {
  switch (filter) {
    case 'all':
      return true
    case 'packets':
      return record.kind.startsWith('packet.') || record.kind.startsWith('ping.')
    case 'routing':
      return record.kind.startsWith('route.')
    case 'links':
      return record.kind.startsWith('link.')
    case 'failures':
      return record.kind === 'packet.dropped' || Boolean(record.reason)
    default:
      return true
  }
}

export function filterRecords(
  records: TraceRecord[],
  filter: Filter,
): TraceRecord[] {
  return filter === 'all'
    ? records
    : records.filter((record) => matchesFilter(record, filter))
}

export type Tone = 'neutral' | 'success' | 'failure'

export function recordTone(record: TraceRecord): Tone {
  if (record.kind === 'packet.dropped' || record.reason) {
    return 'failure'
  }
  if (record.kind === 'packet.delivered' || record.kind === 'ping.reply') {
    return 'success'
  }
  return 'neutral'
}

export function packetJourney(
  records: TraceRecord[],
  packetId: string | undefined,
): TraceRecord[] {
  if (!packetId) {
    return []
  }
  return records.filter((record) => record.packetId === packetId)
}

export function countByFilter(
  records: TraceRecord[],
  filter: Filter,
): number {
  return filterRecords(records, filter).length
}

export function nodeLabel(id: string | undefined): string {
  return id ?? '—'
}
