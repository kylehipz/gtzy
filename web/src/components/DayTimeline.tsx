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
const MOVE_THRESHOLD_PX = 4

// Inverse of topFor: clamp minutes-since-midnight into the day and format "HH:MM".
function minutesToHHMM(mins: number): string {
  const max = (END_HOUR - START_HOUR) * 60 - RESIZE_STEP_MINUTES
  const clamped = Math.max(0, Math.min(Math.round(mins), max))
  const h = Math.floor(clamped / 60) + START_HOUR
  const m = clamped % 60
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`
}

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
    const dur = t.estimated_minutes || 30
    const startMin = t.scheduled_start ? topFor(t.scheduled_start) / PX_PER_MIN : cursor
    if (!t.scheduled_start) cursor += dur
    return { task: t, top: startMin * PX_PER_MIN, startMin, endMin: startMin + dur, colIndex: 0, colCount: 1 }
  })

  // Google-Calendar-style side-by-side layout: blocks whose time intervals
  // overlap split the horizontal width. Grouped into clusters of transitively
  // overlapping intervals; greedy column packing within each cluster.
  {
    let cluster: typeof blocks = []
    let clusterEnd = -1
    const flush = (c: typeof blocks) => {
      const colEnds: number[] = [] // last endMin per column
      for (const b of c) {
        let col = colEnds.findIndex((e) => e <= b.startMin)
        if (col === -1) {
          col = colEnds.length
          colEnds.push(b.endMin)
        } else {
          colEnds[col] = b.endMin
        }
        b.colIndex = col
      }
      for (const b of c) b.colCount = colEnds.length
    }
    for (const b of [...blocks].sort((a, z) => a.startMin - z.startMin || a.endMin - z.endMin)) {
      if (cluster.length > 0 && b.startMin >= clusterEnd) {
        flush(cluster)
        cluster = []
        clusterEnd = -1
      }
      cluster.push(b)
      clusterEnd = Math.max(clusterEnd, b.endMin)
    }
    if (cluster.length > 0) flush(cluster)
  }

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

        {blocks.map(({ task: t, top, colIndex, colCount }) => (
          <TimelineBlock
            key={t.id}
            task={t}
            top={top}
            colIndex={colIndex}
            colCount={colCount}
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
  colIndex,
  colCount,
  category,
  onSelectTask,
}: {
  task: Task
  top: number
  colIndex: number
  colCount: number
  category: Category | null | undefined
  onSelectTask: (t: Task) => void
}) {
  const updateTask = useUpdateTask()
  const [previewMinutes, setPreviewMinutes] = useState<number | null>(null)
  const dragState = useRef<{ startY: number; startMinutes: number } | null>(null)

  const [previewStartMinutes, setPreviewStartMinutes] = useState<number | null>(null)
  const moveState = useRef<{ startY: number; startMinutes: number; moved: boolean } | null>(null)
  const suppressClickRef = useRef(false)

  const baseMinutes = task.estimated_minutes || 30
  const minutes = previewMinutes ?? baseMinutes
  const height = Math.max(minutes * PX_PER_MIN, 24)

  const effectiveTop = previewStartMinutes !== null ? previewStartMinutes * PX_PER_MIN : top

  // Horizontal position: overlapping blocks split the band into equal columns.
  // Band matches the former `left-16 right-2` (64px gutter, 8px right margin).
  const COLUMN_GAP = 2
  const left = `calc(64px + (100% - 72px) * ${colIndex} / ${colCount})`
  const width = `calc((100% - 72px) / ${colCount} - ${COLUMN_GAP}px)`

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

  // Move drag: dragging the block body vertically sets its scheduled_start.
  function snappedStart(state: { startY: number; startMinutes: number }, clientY: number): number {
    const deltaMinutes = (clientY - state.startY) / PX_PER_MIN
    const stepped = Math.round((state.startMinutes + deltaMinutes) / RESIZE_STEP_MINUTES) * RESIZE_STEP_MINUTES
    const max = (END_HOUR - START_HOUR) * 60 - RESIZE_STEP_MINUTES
    return Math.max(0, Math.min(stepped, max))
  }

  function handleMovePointerMove(e: PointerEvent) {
    const state = moveState.current
    if (!state) return
    if (!state.moved && Math.abs(e.clientY - state.startY) < MOVE_THRESHOLD_PX) return
    state.moved = true
    setPreviewStartMinutes(snappedStart(state, e.clientY))
  }

  function handleMovePointerUp(e: PointerEvent) {
    window.removeEventListener('pointermove', handleMovePointerMove)
    window.removeEventListener('pointerup', handleMovePointerUp)
    const state = moveState.current
    moveState.current = null
    setPreviewStartMinutes(null)
    if (!state || !state.moved) return
    suppressClickRef.current = true
    const next = minutesToHHMM(snappedStart(state, e.clientY))
    if (next !== task.scheduled_start) {
      updateTask.mutate({ id: task.id, patch: { scheduled_start: next } })
    }
  }

  function handleMovePointerDown(e: React.PointerEvent) {
    moveState.current = { startY: e.clientY, startMinutes: top / PX_PER_MIN, moved: false }
    window.addEventListener('pointermove', handleMovePointerMove)
    window.addEventListener('pointerup', handleMovePointerUp)
  }

  function handleClick() {
    if (suppressClickRef.current) {
      suppressClickRef.current = false
      return
    }
    onSelectTask(task)
  }

  return (
    <button
      type="button"
      onClick={handleClick}
      onPointerDown={handleMovePointerDown}
      className={`absolute touch-none cursor-grab overflow-hidden rounded-lg border px-2 py-1 text-left text-xs text-text active:cursor-grabbing ${statusClass}`}
      style={{ top: effectiveTop, left, width, height, backgroundColor }}
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
