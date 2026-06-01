import { useEffect, useRef, useCallback } from 'react'
import type { WSMessage, ActionEvent } from '../types'

export type WSStatus = 'connecting' | 'connected' | 'disconnected'

interface UseWebSocketOptions {
  onActionEvent: (event: ActionEvent) => void
  onStatusChange: (status: WSStatus) => void
}

/** Opens and manages a persistent WebSocket connection to /api/ws.
 *  Reconnects automatically after 3 s on close. */
export function useWebSocket({ onActionEvent, onStatusChange }: UseWebSocketOptions) {
  const wsRef = useRef<WebSocket | null>(null)
  const retryRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const mountedRef = useRef(true)

  const connect = useCallback(() => {
    if (!mountedRef.current) return
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const ws = new WebSocket(`${proto}//${location.host}/api/ws`)
    wsRef.current = ws

    ws.onopen = () => {
      if (mountedRef.current) onStatusChange('connected')
    }

    ws.onmessage = (e: MessageEvent<string>) => {
      try {
        const msg = JSON.parse(e.data) as WSMessage
        if (msg.type === 'action' && msg.data) {
          onActionEvent(msg.data as ActionEvent)
        }
      } catch { /* ignore malformed frames */ }
    }

    ws.onclose = () => {
      if (!mountedRef.current) return
      onStatusChange('disconnected')
      retryRef.current = setTimeout(connect, 3000)
    }

    ws.onerror = () => ws.close()
  }, [onActionEvent, onStatusChange])

  useEffect(() => {
    mountedRef.current = true
    onStatusChange('connecting')
    connect()
    return () => {
      mountedRef.current = false
      if (retryRef.current) clearTimeout(retryRef.current)
      wsRef.current?.close()
    }
  }, [connect, onStatusChange])
}
