import { useEffect, useRef, useState } from 'react'

const WS_BASE = import.meta.env.VITE_WS_URL ?? 'ws://localhost:8080'

/**
 * Connects to /ws/track/{busId}, auto-reconnects with exponential backoff,
 * and calls onMessage(telemetry) for every update.
 * Returns the connection state: 'connecting' | 'open' | 'closed'.
 * onMessage is kept in a ref, so changing it never reconnects the socket.
 */
export function useBusSocket(busId, onMessage) {
  const [status, setStatus] = useState('connecting')
  const handler = useRef(onMessage)

  useEffect(() => {
    handler.current = onMessage
  }, [onMessage])

  useEffect(() => {
    let ws
    let timer
    let attempt = 0
    let stopped = false

    const connect = () => {
      setStatus('connecting')
      ws = new WebSocket(`${WS_BASE}/ws/track/${busId}`)

      ws.onopen = () => {
        attempt = 0
        setStatus('open')
      }
      ws.onmessage = (e) => {
        try {
          handler.current(JSON.parse(e.data))
        } catch {
          /* ignore malformed frames */
        }
      }
      ws.onclose = () => {
        if (stopped) return
        setStatus('closed')
        const delay = Math.min(1000 * 2 ** attempt++, 15000)
        timer = setTimeout(connect, delay)
      }
      ws.onerror = () => ws.close()
    }

    connect()
    return () => {
      stopped = true
      clearTimeout(timer)
      ws?.close()
    }
  }, [busId])

  return status
}