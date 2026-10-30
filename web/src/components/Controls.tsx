import type { Replay } from '../hooks/useReplay'

interface Props {
  replay: Replay
  total: number
}

const SPEEDS = [0.5, 1, 2, 4]

export function Controls({ replay, total }: Props) {
  const atStart = replay.index <= 0
  const atEnd = total === 0 || replay.index >= total - 1
  return (
    <div className="controls">
      <button
        type="button"
        className="control-btn"
        onClick={replay.first}
        disabled={atStart}
        title="First event"
        aria-label="First event"
      >
        |◀
      </button>
      <button
        type="button"
        className="control-btn"
        onClick={replay.prev}
        disabled={atStart}
        title="Previous event"
        aria-label="Previous event"
      >
        ◀
      </button>
      <button
        type="button"
        className="control-btn control-btn--primary"
        onClick={replay.toggle}
        disabled={total === 0}
        title={replay.playing ? 'Pause replay' : 'Play replay'}
        aria-label={replay.playing ? 'Pause replay' : 'Play replay'}
      >
        {replay.playing ? '❚❚' : '▶'}
      </button>
      <button
        type="button"
        className="control-btn"
        onClick={replay.next}
        disabled={atEnd}
        title="Next event"
        aria-label="Next event"
      >
        ▶
      </button>
      <button
        type="button"
        className="control-btn"
        onClick={replay.last}
        disabled={atEnd}
        title="Last event"
        aria-label="Last event"
      >
        ▶|
      </button>

      <select
        className="speed"
        value={replay.speed}
        onChange={(event) => replay.setSpeed(Number(event.target.value))}
        aria-label="Replay speed"
      >
        {SPEEDS.map((speed) => (
          <option key={speed} value={speed}>
            {speed}×
          </option>
        ))}
      </select>

      <input
        className="scrub"
        type="range"
        min={0}
        max={Math.max(0, total - 1)}
        value={replay.index}
        onChange={(event) => replay.setIndex(Number(event.target.value))}
        aria-label="Replay position"
      />

      <span className="progress">
        <b>{total === 0 ? 0 : replay.index + 1}</b> / {total}
      </span>
    </div>
  )
}
