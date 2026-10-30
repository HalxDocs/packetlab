/** Types mirroring the Go trace bundle. Keep field names identical. */

export interface TraceRecord {
  index: number
  atMs: number
  atNs: number
  kind: string
  nodeId?: string
  linkId?: string
  packetId?: string
  reason?: string
  summary: string
  details?: Record<string, string>
}

export interface TraceInterface {
  id: string
  node: string
  address: string
  subnet: string
}

export interface TraceRoute {
  network: string
  nextHop?: string
  interface: string
  metric?: number
}

export interface TraceNode {
  id: string
  name?: string
  kind: string
  down?: boolean
  interfaces: TraceInterface[]
  routes?: TraceRoute[]
}

export interface TraceLink {
  id: string
  a: string
  b: string
  delayMs: number
  bandwidthBps?: number
  up: boolean
}

export interface TraceTopology {
  nodes: TraceNode[]
  links: TraceLink[]
}

export interface TraceResult {
  processed: number
  delivered: number
  dropped: number
  durationMs: number
  completed: boolean
  limitReached: boolean
  records: number
}

export interface TraceScenario {
  schemaVersion: string
  id: string
  name: string
  description?: string
  seed: number
}

export interface TraceBundle {
  schemaVersion: string
  scenario: TraceScenario
  topology: TraceTopology
  records: TraceRecord[]
  result: TraceResult
}
