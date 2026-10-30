import { describe, expect, it } from 'vitest'
import { SCENARIOS, tracePath } from './scenarios'

describe('scenario catalogue', () => {
  it('uses unique ids', () => {
    const ids = SCENARIOS.map((scenario) => scenario.id)
    expect(new Set(ids).size).toBe(ids.length)
  })

  it('gives every scenario a name and a blurb', () => {
    for (const scenario of SCENARIOS) {
      expect(scenario.name.length).toBeGreaterThan(0)
      expect(scenario.blurb.length).toBeGreaterThan(0)
    }
  })

  it('maps ids to trace bundle paths', () => {
    expect(tracePath('link-down')).toBe('traces/link-down.json')
  })
})
