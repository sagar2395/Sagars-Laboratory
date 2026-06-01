import type { Notification } from '../types'

interface NotificationListProps {
  items: Notification[]
  onDismiss: (id: number) => void
}

export function NotificationList({ items, onDismiss }: NotificationListProps) {
  if (items.length === 0) return null

  return (
    <div className="notif-container">
      {items.map(n => (
        <div key={n.id} className={`notif notif-${n.level}`}>
          <div className="notif-header">
            <span className="notif-title">{n.title}</span>
            <button className="notif-close" onClick={() => onDismiss(n.id)}>×</button>
          </div>
          {n.detail && <div className="notif-detail">{n.detail}</div>}
        </div>
      ))}
    </div>
  )
}
