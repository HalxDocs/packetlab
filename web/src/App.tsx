import { useEffect, useState } from 'react'
import { Controls } from './components/Controls'
import { EventTimeline } from './components/EventTimeline'
import { Header } from './components/Header'
import { PacketInspector } from './components/PacketInspector'
import { ScenarioPicker } from './components/ScenarioPicker'
import { TopologyCanvas } from './components/TopologyCanvas'
import { useReplay } from './hooks/useReplay'
import { loadTrace } from './lib/api'
import type { Filter } from './lib/format'
import { SCENARIOS } from './scenarios'
import type { TraceBundle } from './types'

export default function App() {
  const [scenarioId, setScenarioId] = useState(SCENARIOS[0].id)
  const [bundle, setBundle] = useState<TraceBundle | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [filter, setFilter] = useState<Filter>('all')
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)

  const replay = useReplay(bundle?.records.length ?? 0, scenarioId)

  useEffect(() => {
    const controller = new AbortController()
    setLoading(true)
    setError(null)
    setBundle(null)
    setSelectedNodeId(null)
    setFilter('all')

    loadTrace(scenarioId, controller.signal)
      .then((next) => {
        setBundle(next)
        setLoading(false)
      })
      .catch((cause: unknown) => {
        if (controller.signal.aborted) {
          return
        }
        setError(cause instanceof Error ? cause.message : String(cause))
        setLoading(false)
      })

    return () => controller.abort()
  }, [scenarioId])

  useEffect(() => {
    if (bundle) {
      document.title = `${bundle.scenario.name} · PacketLab`
    }
  }, [bundle])

  if (loading) {
    return <div className="loading">loading trace…</div>
  }

  if (error || !bundle) {
    return (
      <div className="error">
        <div className="error-box">
          <h1>Could not load the trace</h1>
          <p>{error ?? 'The trace bundle is missing.'}</p>
          <p>
            Build the bundles with <span className="mono">make traces</span> and
            reload this page.
          </p>
        </div>
      </div>
    )
  }

  const record = bundle.records[replay.index]

  const selectEvent = (index: number) => {
    if (replay.playing) {
      replay.toggle()
    }
    replay.setIndex(index)
    setSelectedNodeId(null)
  }

  return (
    <div className="workbench">
      <Header bundle={bundle} />
      <ScenarioPicker
        active={scenarioId}
        onSelect={(id) => {
          if (id !== scenarioId) {
            setScenarioId(id)
          }
        }}
      />
      <TopologyCanvas
        bundle={bundle}
        record={record}
        selectedNodeId={selectedNodeId}
        onSelectNode={setSelectedNodeId}
      />
      <PacketInspector
        bundle={bundle}
        record={record}
        records={bundle.records}
        nodeId={selectedNodeId ?? record?.nodeId}
        onSelect={selectEvent}
      />
      <EventTimeline
        records={bundle.records}
        index={replay.index}
        filter={filter}
        onFilterChange={setFilter}
        onSelect={selectEvent}
        controls={<Controls replay={replay} total={bundle.records.length} />}
      />
    </div>
  )
}
