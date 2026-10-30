import { SCENARIOS } from '../scenarios'

interface Props {
  active: string
  onSelect: (id: string) => void
}

export function ScenarioPicker({ active, onSelect }: Props) {
  return (
    <aside className="sidebar">
      <section className="side-section">
        <h2 className="side-heading">
          <span>Experiments</span>
          <span>{SCENARIOS.length}</span>
        </h2>
        <div className="scenario-list">
          {SCENARIOS.map((scenario) => (
            <button
              key={scenario.id}
              type="button"
              className={`scenario-item${scenario.id === active ? ' is-active' : ''}`}
              onClick={() => onSelect(scenario.id)}
              aria-pressed={scenario.id === active}
            >
              <span className="scenario-item-name">{scenario.name}</span>
              <span className="scenario-item-blurb">{scenario.blurb}</span>
            </button>
          ))}
        </div>
      </section>

      <section className="side-section">
        <h2 className="side-heading">Legend</h2>
        <div className="legend">
          <div className="legend-row">
            <span className="swatch swatch--green" />
            delivered / echo reply
          </div>
          <div className="legend-row">
            <span className="swatch swatch--red" />
            dropped with a reason
          </div>
          <div className="legend-row">
            <span className="swatch swatch--blue" />
            selected event
          </div>
          <div className="legend-row">
            <span className="swatch swatch--dashed" />
            link down
          </div>
        </div>
      </section>

      <p className="side-note">
        Every row is <strong>engine output</strong>. Replay moves through the
        recorded event log; it does not simulate.
      </p>
    </aside>
  )
}
