import { Check, X } from 'lucide-react'
import type { CalendarDay } from '../lib/types'

const STATE_STYLES: Record<CalendarDay['state'], string> = {
  empty: 'bg-mantle border-surface0',
  complete: 'bg-green/20 border-green/40',
  partial: 'bg-yellow/20 border-yellow/40',
  missed: 'bg-red/20 border-red/40',
}

export function MonthGrid({
  year,
  month,
  days,
  onSelectDay,
}: {
  year: number
  month: number
  days: CalendarDay[]
  onSelectDay: (date: string) => void
}) {
  const firstOfMonth = new Date(year, month - 1, 1)
  const leadingBlanks = firstOfMonth.getDay()

  return (
    <div className="grid grid-cols-7 gap-2">
      {['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'].map((d) => (
        <div key={d} className="text-center text-xs font-medium text-subtext0">
          {d}
        </div>
      ))}
      {Array.from({ length: leadingBlanks }).map((_, i) => (
        <div key={`blank-${i}`} />
      ))}
      {days.map((day) => {
        const dayNum = Number(day.date.split('-')[2])
        return (
          <button
            key={day.date}
            type="button"
            onClick={() => onSelectDay(day.date)}
            className={`flex aspect-square flex-col items-center justify-center rounded-lg border text-sm text-text transition-transform hover:scale-105 ${STATE_STYLES[day.state]}`}
            style={day.total > 0 ? { opacity: 0.5 + day.ratio * 0.5 } : undefined}
          >
            <span className="flex items-center gap-1">
              {dayNum}
              {day.state === 'complete' && <Check size={12} className="text-green" />}
              {day.state === 'missed' && <X size={12} className="text-red" />}
            </span>
            {day.total > 0 && (
              <span className="text-[10px] text-subtext0">
                {day.done}/{day.total}
              </span>
            )}
          </button>
        )
      })}
    </div>
  )
}
