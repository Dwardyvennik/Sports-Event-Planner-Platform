import React from 'react'
import { EventFilters } from '../types/event'

interface FiltersProps {
  filters: EventFilters
  onChange: (filters: EventFilters) => void
  onWeekChange: (direction: 'prev' | 'current' | 'next') => void
  sports: string[]
  countries: string[]
}

const Filters: React.FC<FiltersProps> = ({
  filters,
  onChange,
  onWeekChange,
  sports,
  countries,
}) => {
  const update = (key: keyof EventFilters, value: string) => {
    onChange({ ...filters, [key]: value })
  }

  return (
    <div className="filters-section">
      {/* Week selector */}
      <div className="week-selector">
        <button className="week-btn" onClick={() => onWeekChange('prev')}>
          ← Previous Week
        </button>
        <button className="week-btn week-btn--active" onClick={() => onWeekChange('current')}>
          Current Week
        </button>
        <button className="week-btn" onClick={() => onWeekChange('next')}>
          Next Week →
        </button>
      </div>

      {/* Filter dropdowns */}
      <div className="filter-row">
        <select
          className="filter-select"
          value={filters.sport}
          onChange={(e) => update('sport', e.target.value)}
        >
          <option value="all">All Sports</option>
          {sports.map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>

        <select
          className="filter-select"
          value={filters.country}
          onChange={(e) => update('country', e.target.value)}
        >
          <option value="all">All Countries</option>
          {countries.map((c) => (
            <option key={c} value={c}>{c}</option>
          ))}
        </select>

        <select
          className="filter-select"
          value={filters.status}
          onChange={(e) => update('status', e.target.value)}
        >
          <option value="all">All Statuses</option>
          <option value="upcoming">Upcoming</option>
          <option value="live">Live</option>
          <option value="finished">Finished</option>
        </select>

        <input
          type="text"
          className="filter-search"
          placeholder="Search by event name..."
          value={filters.q}
          onChange={(e) => update('q', e.target.value)}
        />
      </div>
    </div>
  )
}

export default Filters
