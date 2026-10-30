import type { TraceBundle } from '../types'

interface Props {
  bundle: TraceBundle
}

export function Header({ bundle }: Props) {
  const dropped = bundle.result.dropped
  return (
    <header className="topbar">
      <div className="brand">
        <span className="brand-mark" aria-hidden="true">
          {Array.from({ length: 9 }, (_, i) => (
            <span key={i} />
          ))}
        </span>
        <span>
          <span className="brand-name">PACKETLAB</span>
          <span className="brand-sub">Network workbench</span>
        </span>
      </div>

      <div className="topbar-meta">
        <span className="topbar-title">{bundle.scenario.name}</span>
        {bundle.scenario.description ? (
          <span className="topbar-desc">{bundle.scenario.description}</span>
        ) : null}
      </div>

      <div className="topbar-status">
        <span className="chip">seed {bundle.scenario.seed}</span>
        <span className="chip chip--blue">{bundle.result.records} records</span>
        {dropped > 0 ? (
          <span className="chip chip--danger">
            <span className="chip-dot" />
            {dropped} dropped
          </span>
        ) : (
          <span className="chip chip--success">
            <span className="chip-dot" />
            {bundle.result.delivered} delivered
          </span>
        )}
      </div>
    </header>
  )
}
