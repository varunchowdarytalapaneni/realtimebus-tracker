import { useEffect, useState } from 'react'

const TONES = {
  green: 'bg-green-100 text-green-800',
  amber: 'bg-amber-100 text-amber-800',
  red: 'bg-red-100 text-red-800',
  slate: 'bg-slate-100 text-slate-700',
}

function liveStatus(conn, ageSec) {
  if (conn === 'connecting') return { label: 'Connecting…', tone: 'amber' }
  if (conn === 'closed') return { label: 'Reconnecting…', tone: 'amber' }
  if (ageSec == null) return { label: 'Waiting for bus…', tone: 'slate' }
  if (ageSec < 15) return { label: 'Live', tone: 'green' }
  if (ageSec < 60) return { label: 'Delayed', tone: 'amber' }
  return { label: 'Signal lost', tone: 'red' }
}

function etaText(info, ageSec) {
  if (!info || ageSec == null || ageSec >= 60) return '—'
  if (info.distance < 50) return 'Arriving'
  if (info.avgSpeed < 1) return 'Stopped'
  const minutes = Math.max(1, Math.round(info.distance / info.avgSpeed / 60))
  return `~${minutes} min`
}

function Row({ label, value }) {
  return (
    <div className="flex items-baseline justify-between py-1.5">
      <dt className="text-xs uppercase tracking-wide text-slate-500">{label}</dt>
      <dd className="text-sm font-semibold text-slate-900">{value}</dd>
    </div>
  )
}

/** Only this component ticks every second; the map is untouched. */
export default function InfoOverlay({ info, conn, busNumber, destinationName, follow, onToggleFollow }) {
  const [now, setNow] = useState(() => Date.now())
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(id)
  }, [])

  const t = info?.t
  const ageSec = t ? Math.max(0, Math.round((now - Date.parse(t.timestamp)) / 1000)) : null
  const status = liveStatus(conn, ageSec)

  return (
    <div className="absolute left-4 top-4 z-1100 w-72 rounded-lg border border-slate-200 bg-white/95 p-4 shadow-md backdrop-blur">
      <div className="mb-2 flex items-center justify-between">
        <h2 className="text-base font-semibold text-slate-900">Bus {busNumber}</h2>
        <span className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${TONES[status.tone]}`}>
          {status.label}
        </span>
      </div>

      <dl className="divide-y divide-slate-100">
        <Row label="Speed" value={t ? `${Math.round(t.speed * 3.6)} km/h` : '—'} />
        <Row
          label="Last updated"
          value={t ? (ageSec < 5 ? 'just now' : `${ageSec}s ago`) : '—'}
        />
        <Row label={`ETA to ${destinationName}`} value={etaText(info, ageSec)} />
        <Row label="Distance" value={info ? `${(info.distance / 1000).toFixed(1)} km` : '—'} />
      </dl>

      <label className="mt-3 flex cursor-pointer items-center gap-2 text-sm text-slate-700">
        <input type="checkbox" checked={follow} onChange={onToggleFollow} />
        Follow bus
      </label>
      <p className="mt-2 text-[11px] leading-snug text-slate-400">
        ETA is a straight-line estimate from current speed.
      </p>
    </div>
  )
}