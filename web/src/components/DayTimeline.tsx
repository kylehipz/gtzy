import type { Category, Task } from '../lib/types'
import { PriorityBadge } from './PriorityBadge'
import { CategoryBadge } from './CategoryBadge'

const START_HOUR = 6
const END_HOUR = 24
const PX_PER_MIN = 1.2

export function DayTimeline({
  tasks,
  categoryById,
  onSelectTask,
}: {
  tasks: Task[]
  categoryById: Map<number, Category>
  onSelectTask: (t: Task) => void
}) {
  const scheduled = tasks.filter((t) => t.scheduled_start)
  const unscheduled = tasks.filter((t) => !t.scheduled_start)
  const totalMinutes = (END_HOUR - START_HOUR) * 60

  function topFor(start: string): number {
    const [h, m] = start.split(':').map(Number)
    return (h * 60 + m - START_HOUR * 60) * PX_PER_MIN
  }

  return (
    <div className="flex gap-4">
      <div className="relative flex-1 rounded-xl border border-surface0 bg-mantle" style={{ height: totalMinutes * PX_PER_MIN }}>
        {Array.from({ length: END_HOUR - START_HOUR + 1 }).map((_, i) => {
          const hour = START_HOUR + i
          return (
            <div
              key={hour}
              className="absolute left-0 right-0 flex items-start border-t border-surface0 pl-2 text-xs text-subtext0"
              style={{ top: i * 60 * PX_PER_MIN }}
            >
              {String(hour % 24).padStart(2, '0')}:00
            </div>
          )
        })}

        {scheduled.map((t) => {
          const height = Math.max((t.estimated_minutes || 30) * PX_PER_MIN, 24)
          const statusStyle =
            t.status === 'done'
              ? 'border-green bg-green/20'
              : t.is_active
                ? 'border-accent bg-accent/20 animate-pulse'
                : 'border-overlay0 bg-surface0'
          return (
            <button
              key={t.id}
              type="button"
              onClick={() => onSelectTask(t)}
              className={`absolute left-16 right-2 overflow-hidden rounded-lg border px-2 py-1 text-left text-xs text-text ${statusStyle}`}
              style={{ top: topFor(t.scheduled_start!), height }}
            >
              <div className="truncate font-medium">{t.title}</div>
              <div className="text-subtext0">{t.scheduled_start}</div>
            </button>
          )
        })}
      </div>

      <div className="w-56 shrink-0">
        <p className="mb-2 text-xs font-medium uppercase tracking-wide text-subtext0">Unscheduled</p>
        <div className="flex flex-col gap-2">
          {unscheduled.map((t) => (
            <button
              key={t.id}
              type="button"
              onClick={() => onSelectTask(t)}
              className={`rounded-lg border border-surface0 bg-mantle p-2 text-left text-xs ${t.status === 'done' ? 'opacity-50 line-through' : ''}`}
            >
              <div className="font-medium text-text">{t.title}</div>
              <div className="mt-1 flex gap-1">
                <PriorityBadge priority={t.priority} />
                <CategoryBadge category={t.category_id ? categoryById.get(t.category_id) : null} />
              </div>
            </button>
          ))}
          {unscheduled.length === 0 && <p className="text-xs text-subtext0">None</p>}
        </div>
      </div>
    </div>
  )
}
