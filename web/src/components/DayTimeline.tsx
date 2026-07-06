import { useEffect, useRef, useState } from 'react'
import type { Category, Task } from '../lib/types'
import { useUpdateTask } from '../lib/queries'
import { nowMinutes, todayLocal } from '../lib/time'

const START_HOUR = 0
const END_HOUR = 24
const PX_PER_MIN = 1.2
const MIN_ESTIMATE_MINUTES = 5
const RESIZE_STEP_MINUTES = 5
const DEFAULT_SCROLL_HOUR = 8

export function DayTimeline({
  tasks,
  categoryById,
  onSelectTask,
  date,
}: {
  tasks: Task[]
  categoryById: Map<number, Category>
  onSelectTask: (t: Task) => void
  date?: string
}) {
  const totalMinutes = (END_HOUR - START_HOUR) * 60
  const isToday = date === todayLocal()
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const el = scrollRef.current
    if (!el) return
    const targetMinutes = isToday ? nowMinutes() : DEFAULT_SCROLL_HOUR * 60
    const centered = targetMinutes * PX_PER_MIN - el.clientHeight / 2
    el.scrollTop = Math.max(0, centered)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [date])

  function topFor(start: string): number {
    const [h, m] = start.split(':').map(Number)
    return (h * 60 + m - START_HOUR * 60) * PX_PER_MIN
  }

  // Scheduled tasks render at their start time; tasks without a start time
  // stack sequentially from 00:00 using a running cursor.
  let cursor = 0
  const blocks = tasks.map((t) => {
    if (t.scheduled_start) return { task: t, top: topFor(t.scheduled_start) }
    const top = cursor * PX_PER_MIN
    cursor += t.estimated_minutes || 30
    return { task: t, top }
  })

  return (
    <div ref={scrollRef} className="max-h-[70vh] overflow-y-auto rounded-xl border border-surface0 bg-mantle">
      <div className="relative" style={{ height: totalMinutes * PX_PER_MIN }}>
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

        {isToday && (
          <div className="pointer-events-none absolute left-0 right-0 z-10" style={{ top: nowMinutes() * PX_PER_MIN }}>
            <div className="relative h-px bg-accent">
              <div className="absolute -left-1 -top-[3px] h-2 w-2 rounded-full bg-accent" />
            </div>
          </div>
        )}

        {blocks.map(({ task: t, top }) => (
          <TimelineBlock
            key={t.id}
            task={t}
            top={top}
            category={t.category_id ? categoryById.get(t.category_id) : null}
            onSelectTask={onSelectTask}
          />
        ))}
      </div>
    </div>
  )
}

function TimelineBlock({
  task,
  top,
  category,
  onSelectTask,
}: {
  task: Task
  top: number
  category: Category | null | undefined
  onSelectTask: (t: Task) => void
}) {
  const updateTask = useUpdateTask()
  const [previewMinutes, setPreviewMinutes] = useState<number | null>(null)
  const dragState = useRef<{ startY: number; startMinutes: number } | null>(null)

  const baseMinutes = task.estimated_minutes || 30
  const minutes = previewMinutes ?? baseMinutes
  const height = Math.max(minutes * PX_PER_MIN, 24)

  const tint = category ? `color-mix(in oklab, var(--ctp-${category.color}) 25%, var(--ctp-surface0))` : 'var(--ctp-surface0)'

  const statusClass =
    task.status === 'done' ? 'border-green bg-green/20' : task.is_active ? 'border-accent bg-accent/20 animate-pulse' : 'border-overlay0'
  const backgroundColor = task.status === 'done' || task.is_active ? undefined : tint

  function handlePointerMove(e: PointerEvent) {
    if (!dragState.current) return
    const deltaMinutes = (e.clientY - dragState.current.startY) / PX_PER_MIN
    const stepped = Math.round((dragState.current.startMinutes + deltaMinutes) / RESIZE_STEP_MINUTES) * RESIZE_STEP_MINUTES
    setPreviewMinutes(Math.max(stepped, MIN_ESTIMATE_MINUTES))
  }

  function handlePointerUp() {
    window.removeEventListener('pointermove', handlePointerMove)
    window.removeEventListener('pointerup', handlePointerUp)
    dragState.current = null
    setPreviewMinutes((current) => {
      if (current !== null && current !== baseMinutes) {
        updateTask.mutate({ id: task.id, patch: { estimated_minutes: current } })
      }
      return null
    })
  }

  function handlePointerDown(e: React.PointerEvent) {
    e.stopPropagation()
    dragState.current = { startY: e.clientY, startMinutes: baseMinutes }
    window.addEventListener('pointermove', handlePointerMove)
    window.addEventListener('pointerup', handlePointerUp)
  }

  return (
    <button
      type="button"
      onClick={() => onSelectTask(task)}
      className={`absolute left-16 right-2 overflow-hidden rounded-lg border px-2 py-1 text-left text-xs text-text ${statusClass}`}
      style={{ top, height, backgroundColor }}
    >
      <div className="truncate font-medium">{task.title}</div>
      <div className="text-subtext0">{task.scheduled_start ?? 'Unscheduled'}</div>
      <div
        onPointerDown={handlePointerDown}
        onClick={(e) => e.stopPropagation()}
        className="absolute inset-x-0 bottom-0 h-2 cursor-ns-resize rounded-b-lg hover:bg-overlay0/40"
      />
    </button>
  )
}
