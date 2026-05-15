import React from 'react'
import { SportEvent } from '../types/event'

interface CalendarTableProps {
  events: SportEvent[]
  onEventClick: (event: SportEvent) => void
  loading: boolean
  error: string | null
}

function daysLeft(endDate: string): number {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const end = new Date(endDate)
  const diff = Math.ceil((end.getTime() - today.getTime()) / (1000 * 60 * 60 * 24))
  return diff
}

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

function StatusBadge({ status }: { status: string }) {
  if (status === 'live') {
    return <span className="badge badge--live"><span className="pulse" />LIVE</span>
  }
  if (status === 'upcoming') {
    return <span className="badge badge--upcoming">Upcoming</span>
  }
  return <span className="badge badge--finished">Finished</span>
}

const CalendarTable: React.FC<CalendarTableProps> = ({
  events,
  onEventClick,
  loading,
  error,
}) => {
  if (loading) {
    return <div className="state-box">⏳ Loading events...</div>
  }

  if (error) {
    return <div className="state-box state-box--error">❌ Error: {error}</div>
  }

  if (events.length === 0) {
    return (
      <div className="state-box">
        <div className="empty-icon">📭</div>
        <p>No events found matching your filters.</p>
        <small>Try changing your Sport, Country, or Status filter.</small>
      </div>
    )
  }

  return (
    <div className="table-wrapper">
      <table className="calendar-table">
        <thead>
          <tr>
            <th>Sport</th>
            <th>Event</th>
            <th>Location</th>
            <th>Start</th>
            <th>End</th>
            <th>Status</th>
            <th>Days Left</th>
            <th>Links</th>
          </tr>
        </thead>
        <tbody>
          {events.map((event) => {
            const left = daysLeft(event.end_date)
            const isLive = event.status === 'live'
            const isFinished = event.status === 'finished'

            return (
              <tr
                key={event.id}
                className={`table-row ${isLive ? 'table-row--live' : ''}`}
              >
                <td>
                  <span className="sport-tag">{event.sport}</span>
                </td>
                <td>
                  <button
                    className="event-title-btn"
                    onClick={() => onEventClick(event)}
                  >
                    {event.title}
                  </button>
                </td>
                <td className="location-cell">
                  <span>{event.country}</span>
                  <small>{event.city}</small>
                </td>
                <td>{formatDate(event.start_date)}</td>
                <td>{formatDate(event.end_date)}</td>
                <td>
                  <StatusBadge status={event.status} />
                </td>
                <td className="days-left-cell">
                  {isFinished
                    ? <span className="days-done">Done</span>
                    : left <= 0
                    ? <span className="days-today">Today</span>
                    : <span className={left <= 7 ? 'days-soon' : ''}>{left}d</span>
                  }
                </td>
                <td>
                  <div className="action-btns">
                    {event.website_url && (
                      <a
                        href={event.website_url}
                        target="_blank"
                        rel="noreferrer"
                        className="action-btn"
                      >
                        Web
                      </a>
                    )}
                    {isLive && event.live_url && (
                      <a
                        href={event.live_url}
                        target="_blank"
                        rel="noreferrer"
                        className="action-btn action-btn--live"
                      >
                        ▶ Live
                      </a>
                    )}
                    <button
                      className="action-btn action-btn--details"
                      onClick={() => onEventClick(event)}
                    >
                      Details
                    </button>
                  </div>
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}

export default CalendarTable
