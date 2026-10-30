/** The canonical experiments shipped with the repository. */
export interface ScenarioEntry {
  id: string
  name: string
  blurb: string
}

export const SCENARIOS: ScenarioEntry[] = [
  {
    id: 'two-hosts-one-router',
    name: 'First ping',
    blurb: 'Host A reaches Host B through one router. 1 ms links.',
  },
  {
    id: 'missing-route',
    name: 'Missing route',
    blurb: 'Host A has no default route, so forwarding never starts.',
  },
  {
    id: 'link-down',
    name: 'Broken link',
    blurb: 'The route is valid, but Link B is down.',
  },
  {
    id: 'ttl-loop',
    name: 'TTL loop',
    blurb: 'Two routers point at each other until the TTL expires.',
  },
]

export function tracePath(id: string): string {
  return `traces/${id}.json`
}
