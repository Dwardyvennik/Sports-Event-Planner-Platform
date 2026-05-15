import React, { useEffect, useState, useCallback } from 'react'
import Header from './components/Header'
import Filters from './components/Filters'
import CalendarTable from './components/CalendarTable'
import EventDetails from './components/EventDetails'
import { SportEvent, EventFilters } from './types/event'
import { searchEvents } from './api/events'
import { mockEvents } from './data/mockEvents'

const defaultFilters: EventFilters = {
  sport: 'all',
  country: 'all',
  status: 'all',
  q: '',
}

// Get unique values from mock data for dropdowns
const allSports = [...new Set(mockEvents.map((e) => e.sport))].sort()
const allCountries = [...new Set(mockEvents.map((e) => e.country))].sort()

function App() {
  const [events, setEvents] = useState<SportEvent[]>([])
  const [filters, setFilters] = useState<EventFilters>(defaultFilters)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedEvent, setSelectedEvent] = useState<SportEvent | null>(null)

  const fetchEvents = useCallback(async (currentFilters: EventFilters) => {
    setLoading(true)
    setError(null)
    try {
      const data = await searchEvents(currentFilters)
      setEvents(data)
    } catch (err) {
      setError('Failed to load events. Please try again.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchEvents(filters)
  }, [filters, fetchEvents])

  const handleSearch = (q: string) => {
    setFilters((prev) => ({ ...prev, q }))
  }

  const handleWeekChange = (direction: 'prev' | 'current' | 'next') => {
    // TODO: implement week-based filtering when connecting to backend
    console.log('Week change:', direction)
  }

  return (
    <div className="app">
      <Header onSearch={handleSearch} />

      <main className="main-content">
        <div className="page-title-section">
          <h1 className="page-title">Sports Calendar</h1>
          <p className="page-subtitle">
            Track upcoming, live, and finished sports events worldwide
          </p>
        </div>

        <Filters
          filters={filters}
          onChange={setFilters}
          onWeekChange={handleWeekChange}
          sports={allSports}
          countries={allCountries}
        />

        <CalendarTable
          events={events}
          onEventClick={setSelectedEvent}
          loading={loading}
          error={error}
        />
      </main>

      <EventDetails
        event={selectedEvent}
        onClose={() => setSelectedEvent(null)}
      />
    </div>
  )
}

export default App
