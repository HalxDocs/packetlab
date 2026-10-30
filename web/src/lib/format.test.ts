import { describe, expect, it } from 'vitest'
import {
  countByFilter,
  filterRecords,
  formatTraceTime,
  matchesFilter,
  packetJourney,
  reasonLabel,
  recordTone,
} from './format'
import type { TraceRecord } from '../types'

function record(partial: Partial<TraceRecord>): TraceRecord {
  return {
    index: 0,
    atMs: 0,
    atNs: 0,
    kind: 'packet.created',
    summary: 'created packet',
    ...partial,
  }
}

describe('formatTraceTime', () => {
  it('renders milliseconds with three decimals', () => {
    expect(formatTraceTime(0)).toBe('0.000ms')
    expect(formatTraceTime(1)).toBe('1.000ms')
    expect(formatTraceTime(1.5)).toBe('1.500ms')
    expect(formatTraceTime(999.999)).toBe('999.999ms')
  })

  it('switches to seconds at one second', () => {
    expect(formatTraceTime(1000)).toBe('1.000s')
    expect(formatTraceTime(2500)).toBe('2.500s')
  })

  it('handles negative and invalid values', () => {
    expect(formatTraceTime(-1)).toBe('-1.000ms')
    expect(formatTraceTime(Number.NaN)).toBe('—')
  })
})

describe('reasonLabel', () => {
  it('translates stable reason codes', () => {
    expect(reasonLabel('no_route')).toBe('no route to destination')
    expect(reasonLabel('link_down')).toBe('link is down')
    expect(reasonLabel('ttl_expired')).toBe('TTL expired')
  })

  it('falls back to the raw code and handles empty input', () => {
    expect(reasonLabel('something_new')).toBe('something_new')
    expect(reasonLabel(undefined)).toBe('')
  })
})

describe('filters', () => {
  const records = [
    record({ kind: 'packet.created' }),
    record({ kind: 'route.selected' }),
    record({ kind: 'link.transmit' }),
    record({ kind: 'packet.dropped', reason: 'link_down' }),
    record({ kind: 'ping.reply' }),
  ]

  it('matches each category', () => {
    expect(matchesFilter(records[0], 'packets')).toBe(true)
    expect(matchesFilter(records[1], 'routing')).toBe(true)
    expect(matchesFilter(records[2], 'links')).toBe(true)
    expect(matchesFilter(records[3], 'failures')).toBe(true)
    expect(matchesFilter(records[4], 'packets')).toBe(true)
    expect(matchesFilter(records[4], 'failures')).toBe(false)
  })

  it('counts and filters', () => {
    expect(countByFilter(records, 'failures')).toBe(1)
    expect(filterRecords(records, 'routing')).toHaveLength(1)
    expect(filterRecords(records, 'all')).toHaveLength(5)
  })
})

describe('recordTone', () => {
  it('marks delivery as success and drops as failure', () => {
    expect(recordTone(record({ kind: 'packet.delivered' }))).toBe('success')
    expect(recordTone(record({ kind: 'ping.reply' }))).toBe('success')
    expect(recordTone(record({ kind: 'packet.dropped' }))).toBe('failure')
    expect(recordTone(record({ kind: 'packet.created' }))).toBe('neutral')
  })
})

describe('packetJourney', () => {
  it('returns every record for one packet', () => {
    const records = [
      record({ index: 1, packetId: 'p-001' }),
      record({ index: 2, packetId: 'p-002' }),
      record({ index: 3, packetId: 'p-001' }),
    ]
    expect(packetJourney(records, 'p-001').map((r) => r.index)).toEqual([1, 3])
    expect(packetJourney(records, undefined)).toEqual([])
  })
})
