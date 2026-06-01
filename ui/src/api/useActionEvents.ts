import { useState, useEffect, useCallback } from 'react';
import type { ActionEvent, ClusterInfo } from './client';

export type WsMessage = 
  | { type: 'status'; data: ClusterInfo }
  | { type: 'action'; data: ActionEvent };

export function useActionEvents() {
  const [events, setEvents] = useState<ActionEvent[]>([]);
  const [clusterStatus, setClusterStatus] = useState<ClusterInfo | null>(null);
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    // Determine WebSocket URL based on current protocol
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const wsUrl = `${protocol}//${host}/api/ws`;

    let ws: WebSocket;
    let reconnectTimer: number;

    const connect = () => {
      ws = new WebSocket(wsUrl);

      ws.onopen = () => {
        setIsConnected(true);
      };

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data) as WsMessage;
          if (msg.type === 'action') {
            setEvents((prev) => {
              // Keep last 100 events to avoid memory bloat
              const next = [...prev, msg.data];
              if (next.length > 100) return next.slice(next.length - 100);
              return next;
            });
          } else if (msg.type === 'status') {
            setClusterStatus(msg.data);
          }
        } catch (e) {
          console.error('Failed to parse WS message:', e);
        }
      };

      ws.onclose = () => {
        setIsConnected(false);
        // Attempt to reconnect
        reconnectTimer = window.setTimeout(connect, 3000);
      };

      ws.onerror = (err) => {
        console.error('WebSocket error:', err);
      };
    };

    connect();

    return () => {
      clearTimeout(reconnectTimer);
      if (ws) ws.close();
    };
  }, []);

  const clearEvents = useCallback(() => {
    setEvents([]);
  }, []);

  return { events, clusterStatus, isConnected, clearEvents };
}
