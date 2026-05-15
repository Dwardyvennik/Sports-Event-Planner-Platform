export type EventStatus = 'upcoming' | 'live' | 'finished'

export interface SportEvent {
  id: string
  title: string
  sport: string
  country: string
  city: string
  start_date: string   // "YYYY-MM-DD"
  end_date: string     // "YYYY-MM-DD"
  status: EventStatus
  description: string
  website_url: string
  live_url: string
}

export interface EventFilters {
  sport: string
  country: string
  status: string
  q: string
}
