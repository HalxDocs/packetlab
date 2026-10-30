import { useCallback, useEffect, useState } from 'react'

export interface Replay {
  index: number
  playing: boolean
  speed: number
  setIndex: (index: number) => void
  next: () => void
  prev: () => void
  first: () => void
  last: () => void
  toggle: () => void
  setSpeed: (speed: number) => void
}

const BASE_DELAY_MS = 420

/**
 * Replay state over an already-recorded event log. This never simulates: it
 * only moves a cursor through engine output.
 */
export function useReplay(total: number, resetKey?: string): Replay {
  const [index, setIndexState] = useState(0)
  const [playing, setPlaying] = useState(false)
  const [speed, setSpeed] = useState(1)

  const clamp = useCallback(
    (value: number) => Math.max(0, Math.min(value, Math.max(0, total - 1))),
    [total],
  )

  const setIndex = useCallback(
    (value: number) => setIndexState(clamp(value)),
    [clamp],
  )

  const next = useCallback(
    () => setIndexState((current) => clamp(current + 1)),
    [clamp],
  )

  const prev = useCallback(
    () => setIndexState((current) => clamp(current - 1)),
    [clamp],
  )

  const first = useCallback(() => {
    setPlaying(false)
    setIndexState(0)
  }, [])

  const last = useCallback(() => {
    setPlaying(false)
    setIndexState(clamp(Number.POSITIVE_INFINITY))
  }, [clamp])

  const toggle = useCallback(() => {
    if (playing) {
      setPlaying(false)
      return
    }
    if (index >= total - 1) {
      setIndexState(0)
    }
    setPlaying(true)
  }, [playing, index, total])

  // Reset when a different trace is loaded.
  useEffect(() => {
    setIndexState(0)
    setPlaying(false)
  }, [total, resetKey])

  // Advance while playing. Pacing is fixed wall-clock time per event; virtual
  // timestamps remain the source of truth for what is displayed.
  useEffect(() => {
    if (!playing || total === 0) {
      return
    }
    if (index >= total - 1) {
      setPlaying(false)
      return
    }
    const timer = window.setTimeout(() => {
      setIndexState((current) => clamp(current + 1))
    }, BASE_DELAY_MS / speed)
    return () => window.clearTimeout(timer)
  }, [playing, index, total, speed, clamp])

  return {
    index,
    playing,
    speed,
    setIndex,
    next,
    prev,
    first,
    last,
    toggle,
    setSpeed,
  }
}
