import { tracePath } from '../scenarios'
import type { TraceBundle } from '../types'

/** Load a precomputed trace bundle produced by `packetlab trace`. */
export async function loadTrace(
  id: string,
  signal?: AbortSignal,
): Promise<TraceBundle> {
  const url = new URL(tracePath(id), document.baseURI)
  const response = await fetch(url, { signal })
  if (!response.ok) {
    throw new Error(`could not load ${id} (HTTP ${response.status})`)
  }
  return (await response.json()) as TraceBundle
}
