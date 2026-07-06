import { useEffect, useState } from 'react'
import { Check, Pause, Play, Repeat, Trash2 } from 'lucide-react'
import { motion } from 'framer-motion'
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

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.97 }}
      className={`flex items-center gap-3 rounded-xl border p-3 transition-colors ${
        task.is_active ? 'border-accent/50 bg-accent/5' : 'border-surface0 bg-mantle'
      } ${done ? 'opacity-60' : ''}`}
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
          {task.scheduled_start && <span className="text-xs text-subtext0">{task.scheduled_start}</span>}
          <span className="font-mono text-xs text-subtext0">
            {fmtDuration(elapsed)}
            {task.estimated_minutes > 0 && ` / ${fmtDuration(task.estimated_minutes * 60)}`}
          </span>
        </div>
      </button>

      <div className="flex shrink-0 gap-1">
        {!done && !task.is_active && (
          <IconButton onClick={() => startTask.mutate(task.id)} label="Start">
            <Play size={14} />
          </IconButton>
        )}
        {!done && task.is_active && (
          <IconButton onClick={() => pauseTask.mutate(task.id)} label="Pause">
            <Pause size={14} />
          </IconButton>
        )}
        <IconButton onClick={() => deleteTask.mutate(task.id)} label="Delete">
          <Trash2 size={14} />
        </IconButton>
      </div>
    </motion.div>
  )
}

function IconButton({ children, onClick, label }: { children: React.ReactNode; onClick: () => void; label: string }) {
  return (
    <button
      type="button"
      onClick={onClick}
      title={label}
      aria-label={label}
      className="flex h-8 w-8 items-center justify-center rounded-lg text-subtext0 hover:bg-surface0 hover:text-text"
    >
      {children}
    </button>
  )
}
