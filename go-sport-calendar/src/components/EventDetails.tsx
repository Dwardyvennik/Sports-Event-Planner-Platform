import React from 'react'
import { SportEvent } from '../types/event'

interface EventDetailsProps {
  event: SportEvent | null
  onClose: () => void
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-GB', { day: '2-digit', month: 'long', year: 'numeric' })
}

const EventDetails: React.FC<EventDetailsProps> = ({ event, onClose }) => {
  if (!event) return null

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <button className="modal-close" onClick={onClose}>✕</button>

        <div className="modal-header">
          <span className="modal-sport">{event.sport}</span>
          <h2 className="modal-title">{event.title}</h2>
          <span
            className={`badge ${
              event.status === 'live'
                ? 'badge--live'
                : event.status === 'upcoming'
                ? 'badge--upcoming'
                : 'badge--finished'
            }`}
          >
            {event.status === 'live' && <span className="pulse" />}
            {event.status.toUpperCase()}
          </span>
        </div>

        <div className="modal-body">
          <div className="modal-info-grid">
            <div className="modal-info-item">
              <span className="info-label">📍 Location</span>
              <span className="info-value">{event.city}, {event.country}</span>
            </div>
            <div className="modal-info-item">
              <span className="info-label">📅 Start Date</span>
              <span className="info-value">{formatDate(event.start_date)}</span>
            </div>
            <div className="modal-info-item">
              <span className="info-label">🏁 End Date</span>
              <span className="info-value">{formatDate(event.end_date)}</span>
            </div>
          </div>

          <div className="modal-description">
            <span className="info-label">📝 Description</span>
            <p>{event.description}</p>
          </div>

          <div className="modal-links">
            {event.website_url && (
              <a
                href={event.website_url}
                target="_blank"
                rel="noreferrer"
                className="modal-link-btn"
              >
                🌐 Official Website
              </a>
            )}
            {event.live_url && (
              <a
                href={event.live_url}
                target="_blank"
                rel="noreferrer"
                className="modal-link-btn modal-link-btn--live"
              >
                ▶ Watch Live
              </a>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default EventDetails
