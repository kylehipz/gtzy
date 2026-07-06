import { NavLink } from 'react-router-dom'
import { CalendarDays, LayoutDashboard, ListTodo, NotebookPen, Settings } from 'lucide-react'

const LINKS = [
  { to: '/', label: 'Today', icon: ListTodo },
  { to: '/calendar', label: 'Calendar', icon: CalendarDays },
  { to: '/journal', label: 'Journal', icon: NotebookPen },
  { to: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/settings', label: 'Settings', icon: Settings },
]

export function Sidebar() {
  return (
    <nav className="flex w-56 shrink-0 flex-col gap-1 border-r border-surface0 bg-mantle p-4">
      <div className="mb-4 px-2 text-lg font-semibold tracking-tight text-text">gtzy</div>
      {LINKS.map(({ to, label, icon: Icon }) => (
        <NavLink
          key={to}
          to={to}
          end={to === '/'}
          className={({ isActive }) =>
            `flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
              isActive ? 'bg-accent/15 text-accent' : 'text-subtext0 hover:bg-surface0 hover:text-text'
            }`
          }
        >
          <Icon size={18} />
          {label}
        </NavLink>
      ))}
    </nav>
  )
}
