import { forwardRef, memo, useEffect, useImperativeHandle, useRef } from 'react'
import { CircleMarker, MapContainer, TileLayer, Tooltip, useMap } from 'react-leaflet'
import L from 'leaflet'

const busIcon = L.divIcon({
  className: '',
  html: '<div class="bus-pin"><div class="bus-arrow"></div></div>',
  iconSize: [28, 28],
  iconAnchor: [14, 14],
})

/**
 * Owns the bus marker and moves it imperatively (setLatLng inside
 * requestAnimationFrame). React never re-renders for position updates.
 */
const BusLayer = forwardRef(function BusLayer({ follow }, ref) {
  const map = useMap()
  const marker = useRef(null)
  const current = useRef(null) // position currently drawn
  const raf = useRef(0)
  const lastTs = useRef(0)
  const lastArrival = useRef(0)
  const followRef = useRef(follow)

  useEffect(() => {
    followRef.current = follow
  }, [follow])

  useEffect(
    () => () => {
      cancelAnimationFrame(raf.current)
      marker.current?.remove()
      marker.current = null
    },
    [],
  )

  useImperativeHandle(ref, () => ({
    /** Returns true if the update was accepted (newer than what we've shown). */
    update(t) {
      const ts = Date.parse(t.timestamp)
      if (!(ts > lastTs.current)) return false // stale or duplicate (e.g. snapshot after live ping)
      lastTs.current = ts

      const target = L.latLng(t.latitude, t.longitude)
      const now = performance.now()
      const gap = lastArrival.current ? now - lastArrival.current : 0
      lastArrival.current = now
      const duration = Math.min(Math.max(gap, 500), 3000) // glide for about one ping interval

      if (!marker.current) {
        marker.current = L.marker(target, { icon: busIcon, zIndexOffset: 1000 }).addTo(map)
        current.current = target
        map.setView(target, 15)
      } else {
        cancelAnimationFrame(raf.current)
        const from = current.current
        const start = performance.now()
        const m = marker.current
        const step = (time) => {
          const k = Math.min((time - start) / duration, 1)
          const pos = L.latLng(
            from.lat + (target.lat - from.lat) * k,
            from.lng + (target.lng - from.lng) * k,
          )
          current.current = pos
          m.setLatLng(pos)
          if (k < 1) raf.current = requestAnimationFrame(step)
        }
        raf.current = requestAnimationFrame(step)

        if (followRef.current) {
          map.panTo(target, { animate: true, duration: duration / 1000, easeLinearity: 1 })
        }
      }

      // Heading is unreliable when (almost) stationary, so keep the last one then.
      if (t.speed > 0.5) {
        const arrow = marker.current.getElement()?.querySelector('.bus-arrow')
        if (arrow) arrow.style.transform = `rotate(${t.heading}deg)`
      }
      return true
    },
  }))

  return null
})

/** memo + stable props => the map component itself does not re-render on telemetry. */
const LiveMap = memo(
  forwardRef(function LiveMap({ center, zoom, follow, destination }, ref) {
    return (
      <MapContainer center={center} zoom={zoom} className="h-full w-full">
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        {destination && (
          <CircleMarker
            center={[destination.lat, destination.lng]}
            radius={8}
            pathOptions={{ color: '#b45309', fillColor: '#f59e0b', fillOpacity: 0.9 }}
          >
            <Tooltip permanent direction="top">
              {destination.name}
            </Tooltip>
          </CircleMarker>
        )}
        <BusLayer ref={ref} follow={follow} />
      </MapContainer>
    )
  }),
)

export default LiveMap