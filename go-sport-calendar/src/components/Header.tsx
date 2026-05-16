import React from 'react'

interface HeaderProps {
  onSearch: (q: string) => void
}

const Header: React.FC<HeaderProps> = ({ onSearch }) => {
  return (
    <header className="header">
      <div className="header-inner">
        <div className="logo">
          <span className="logo-icon"></span>
          <span className="logo-text">Sport Calendar</span>
        </div>

        <nav className="nav">
          <a href="#" className="nav-link">Calendar</a>
          <a href="#" className="nav-link">Sports</a>
          <a href="#" className="nav-link">Countries</a>
          <a href="#" className="nav-link">Games</a>
          <a href="#" className="nav-link nav-live">Live</a>
          <a href="#" className="nav-link">About</a>
        </nav>

        <div className="header-search">
          <input
            type="text"
            placeholder="Search events..."
            className="search-input"
            onChange={(e) => onSearch(e.target.value)}
          />
        </div>
      </div>
    </header>
  )
}

export default Header
