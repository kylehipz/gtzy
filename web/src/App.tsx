import { Route, Routes } from 'react-router-dom'
import { Sidebar } from './components/Sidebar'
import { ActiveTimerBar } from './components/ActiveTimerBar'
import { Today } from './pages/Today'
import { Calendar } from './pages/Calendar'
import { Journal } from './pages/Journal'
import { Dashboard } from './pages/Dashboard'
import { Settings } from './pages/Settings'

function App() {
  return (
    <div className="flex h-full bg-base">
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col overflow-y-auto">
        <ActiveTimerBar />
        <Routes>
          <Route path="/" element={<Today />} />
          <Route path="/calendar" element={<Calendar />} />
          <Route path="/journal" element={<Journal />} />
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/settings" element={<Settings />} />
        </Routes>
      </div>
    </div>
  )
}

export default App
