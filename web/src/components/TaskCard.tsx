import { useEffect, useState } from 'react'
import { AlertTriangle, Check, Clock, Pause, Play, Repeat, Timer, Trash2 } from 'lucide-react'
import type { Category, Task } from '../lib/types'
import { elapsedSecondsLive, fmtDuration } from '../lib/time'
import { PriorityBadge } from './PriorityBadge'
import { CategoryBadge } from './CategoryBadge'
import { useCompleteTask, useDeleteTask, usePauseTask, useStartTask } from '../lib/queries'

export function TaskCard({
  task,
  category,
  onEdit,
}: {
  task: Task
  category: Category | null | undefined
  onEdit: () => void
}) {
  const startTask = useStartTask()
  const pauseTask = usePauseTask()
  const completeTask = useCompleteTask()
  const deleteTask = useDeleteTask()
  const [, forceTick] = useState(0)

  useEffect(() => {
    if (!task.is_active) return
    const id = setInterval(() => forceTick((n) => n + 1), 1000)
    return () => clearInterval(id)
  }, [task.is_active])

  const elapsed = elapsedSecondsLive(task.active_started_at, task.actual_seconds)
  const done = task.status === 'done'
  const overEstimate = task.estimated_minutes > 0 && elapsed > task.estimated_minutes * 60

  const tint = category
    ? {
        backgroundColor: `color-mix(in oklab, var(--ctp-${category.color}) 12%, var(--ctp-mantle))`,
        borderColor: `color-mix(in oklab, var(--ctp-${category.color}) 35%, var(--ctp-surface0))`,
      }
    : undefined

  return (
    <div
      style={tint}
      className={`flex items-center gap-3 rounded-xl border p-3 transition-colors ${
        category ? '' : task.is_active ? 'border-accent/50 bg-accent/5' : 'border-surface0 bg-mantle'
      } ${task.is_active ? 'ring-1 ring-accent' : ''} ${done ? 'opacity-60' : ''}`}
    >
      <button
        type="button"
        onClick={() => (done ? undefined : completeTask.mutate(task.id))}
        className={`flex h-5 w-5 shrink-0 items-center justify-center rounded-full border-2 ${
          done ? 'border-green bg-green text-base' : 'border-overlay0 hover:border-green'
        }`}
      >
        {done && <Check size={12} />}
      </button>

      <button type="button" onClick={onEdit} className="flex min-w-0 flex-1 flex-col items-start text-left">
        <span className={`truncate font-medium text-text ${done ? 'line-through' : ''}`}>
          {task.recurrence_id && <Repeat size={12} className="mr-1 inline text-overlay0" />}
          {task.title}
        </span>
        <div className="mt-1 flex flex-wrap items-center gap-1.5">
          <PriorityBadge priority={task.priority} />
          <CategoryBadge category={category} />
          {task.scheduled_start && (
            <span className="flex items-center gap-1 text-xs text-subtext0">
              <Clock size={12} />
              {task.scheduled_start}
            </span>
          )}
          <span className={`ml-2 flex items-center gap-1 text-xs ${overEstimate ? 'text-red' : 'text-subtext0'}`}>
            <Timer size={12} />
            <span className="font-mono">
              {fmtDuration(elapsed)}
              {task.estimated_minutes > 0 && ` / ${fmtDuration(task.estimated_minutes * 60)}`}
            </span>
            {overEstimate && (
              <span title={`Over estimate by ${fmtDuration(elapsed - task.estimated_minutes * 60)}`}>
                <AlertTriangle size={12} />
              </span>
            )}
          </span>
        </div>
      </button>

      <div className="flex shrink-0 gap-1">
        {!done && !task.is_active && (
          <IconButton onClick={() => startTask.mutate(task.id)} label="Start" variant="green">
            <Play size={14} />
          </IconButton>
        )}
        {!done && task.is_active && (
          <IconButton onClick={() => pauseTask.mutate(task.id)} label="Pause" variant="peach">
            <Pause size={14} />
          </IconButton>
        )}
        <IconButton onClick={() => deleteTask.mutate(task.id)} label="Delete" variant="red">
          <Trash2 size={14} />
        </IconButton>
      </div>
    </div>
  )
}

const ICON_BUTTON_VARIANTS = {
  neutral: 'text-subtext0 hover:bg-surface0 hover:text-text',
  green: 'text-green hover:bg-green/10',
  peach: 'text-peach hover:bg-peach/10',
  red: 'text-red hover:bg-red/10',
} as const

export function IconButton({
  children,
  onClick,
  label,
  variant = 'neutral',
}: {
  children: React.ReactNode
  onClick: () => void
  label: string
  variant?: keyof typeof ICON_BUTTON_VARIANTS
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      title={label}
      aria-label={label}
      className={`flex h-8 w-8 items-center justify-center rounded-lg ${ICON_BUTTON_VARIANTS[variant]}`}
    >
      {children}
    </button>
  )
}
