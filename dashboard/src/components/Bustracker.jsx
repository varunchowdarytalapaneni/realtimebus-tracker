import { useCallback, useRef, useState } from 'react'
import LiveMap from './Livemap.jsx'
import InfoOverlay from './Infooverlay.jsx'
import { useBusSocket } from '../hooks/Usebussocket.js'
import { haversineMeters } from '../utils/Geo.js'

// Module-level constants keep props stable, so LiveMap (memo) never re-renders.
const DEFAULT_CENTER = [22.5726, 88.3639] // Kolkata
const DEFAULT_ZOOM = 13
// Demo destination. Later: replace with the next stop from the bus_stops API.
const DESTINATION = { name: 'School', lat: 22.58, lng: 88.37 }

export default function BusTracker({ busId = 1, busNumber = 'WB-001' }) {
  // Ensure numeric ID is always used for the WebSocket connection
  const numericBusId = Number(busId) || 1;

  const mapRef = useRef(null)
  const avgSpeed = useRef(null) // smoothed speed (EMA) for a steadier ETA
  const [info, setInfo] = useState(null)
  const [follow, setFollow] = useState(true)

  const onMessage = useCallback((t) => {
    // 1) Move the marker imperatively (no React render).
    if (!mapRef.current?.update(t)) return // stale/duplicate: ignore
    // 2) Update only the small overlay.
    avgSpeed.current =
      avgSpeed.current == null ? t.speed : 0.7 * avgSpeed.current + 0.3 * t.speed
    setInfo({
      t,
      avgSpeed: avgSpeed.current,
      distance: haversineMeters(t.latitude, t.longitude, DESTINATION.lat, DESTINATION.lng),
    })
  }, [])

  const conn = useBusSocket(numericBusId, onMessage)

  return (
    <div className="relative h-full w-full">
      <LiveMap
        ref={mapRef}
        center={DEFAULT_CENTER}
        zoom={DEFAULT_ZOOM}
        follow={follow}
        destination={DESTINATION}
      />
      <InfoOverlay
        info={info}
        conn={conn}
        busNumber={busNumber}
        destinationName={DESTINATION.name}
        follow={follow}
        onToggleFollow={() => setFollow((f) => !f)}
      />
    </div>
  )
}