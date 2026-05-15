import { SportEvent, EventFilters } from '../types/event'
import { mockEvents } from '../data/mockEvents'

const BASE_URL = '/api' // TODO: set real Go backend URL here

// TODO: Replace mock logic with: fetch(`${BASE_URL}/events`)
export async function getEvents(): Promise<SportEvent[]> {
  // Simulating async API call
  return new Promise((resolve) => {
    setTimeout(() => resolve(mockEvents), 300)
  })
}

// TODO: Replace mock logic with: fetch(`${BASE_URL}/events/${id}`)
export async function getEventById(id: string): Promise<SportEvent | undefined> {
  return new Promise((resolve) => {
    setTimeout(() => {
      const event = mockEvents.find((e) => e.id === id)
      resolve(event)
    }, 200)
  })
}

// TODO: Replace mock logic with: fetch(`${BASE_URL}/events?sport=&country=&status=&q=`)
export async function searchEvents(filters: Partial<EventFilters>): Promise<SportEvent[]> {
  return new Promise((resolve) => {
    setTimeout(() => {
      let result = [...mockEvents]

      if (filters.sport && filters.sport !== 'all') {
        result = result.filter((e) =>
          e.sport.toLowerCase() === filters.sport!.toLowerCase()
        )
      }
      if (filters.country && filters.country !== 'all') {
        result = result.filter((e) =>
          e.country.toLowerCase() === filters.country!.toLowerCase()
        )
      }
      if (filters.status && filters.status !== 'all') {
        result = result.filter((e) => e.status === filters.status)
      }
      if (filters.q) {
        result = result.filter((e) =>
          e.title.toLowerCase().includes(filters.q!.toLowerCase())
        )
      }

      resolve(result)
    }, 300)
  })
}
