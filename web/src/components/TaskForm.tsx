import { useEffect, useState } from 'react'
import { X } from 'lucide-react'
import { motion } from 'framer-motion'
import type { Category, Priority, Task } from '../lib/types'
import { todayLocal } from '../lib/time'
import { useCreateRecurrence, useCreateTask, useDeleteTask, useUpdateTask } from '../lib/queries'

const WEEKDAYS = [
  { n: 0, label: 'Sun' },
  { n: 1, label: 'Mon' },
  { n: 2, label: 'Tue' },
  { n: 3, label: 'Wed' },
  { n: 4, label: 'Thu' },
  { n: 5, label: 'Fri' },
  { n: 6, label: 'Sat' },
]

type RepeatMode = 'none' | 'daily' | 'weekly' | 'monthly'

export function TaskForm({
  task,
  categories,
  defaultDate,
  onClose,
}: {
  task: Task | null
  categories: Category[]
  defaultDate: string
  onClose: () => void
}) {
  const isEdit = !!task
  const [title, setTitle] = useState(task?.title ?? '')
  const [notes, setNotes] = useState(task?.notes ?? '')
  const [priority, setPriority] = useState<Priority>(task?.priority ?? 'medium')
  const [categoryId, setCategoryId] = useState<string>(task?.category_id?.toString() ?? '')
  const [estimatedMinutes, setEstimatedMinutes] = useState(task?.estimated_minutes ?? 0)
  const [scheduledDate, setScheduledDate] = useState(task?.scheduled_date ?? defaultDate)
  const [scheduledStart, setScheduledStart] = useState(task?.scheduled_start ?? '')

  const [repeatMode, setRepeatMode] = useState<RepeatMode>('none')
  const [interval, setIntervalValue] = useState(1)
  const [weekdays, setWeekdays] = useState<number[]>([])
  const [dayOfMonth, setDayOfMonth] = useState(1)
  const [endDate, setEndDate] = useState('')

  const createTask = useCreateTask()
  const updateTask = useUpdateTask()
  const deleteTask = useDeleteTask()
  const createRecurrence = useCreateRecurrence()

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [onClose])

  function toggleWeekday(n: number) {
    setWeekdays((prev) => (prev.includes(n) ? prev.filter((d) => d !== n) : [...prev, n].sort()))
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!title.trim()) return

    const category_id = categoryId ? Number(categoryId) : null

    if (isEdit && task) {
      await updateTask.mutateAsync({
        id: task.id,
        patch: {
          title,
          notes,
          priority,
          category_id,
          estimated_minutes: estimatedMinutes,
          scheduled_date: scheduledDate || null,
          scheduled_start: scheduledStart || null,
        },
      })
      onClose()
      return
    }

    if (repeatMode !== 'none') {
      await createRecurrence.mutateAsync({
        title,
        notes,
        priority,
        category_id: category_id ?? undefined,
        estimated_minutes: estimatedMinutes,
        scheduled_start: scheduledStart || undefined,
        freq: repeatMode,
        interval,
        days_of_week: repeatMode === 'weekly' ? weekdays.join(',') : '',
        day_of_month: repeatMode === 'monthly' ? dayOfMonth : undefined,
        start_date: scheduledDate || todayLocal(),
        end_date: endDate || undefined,
      })
    } else {
      await createTask.mutateAsync({
        title,
        notes,
        priority,
        category_id,
        estimated_minutes: estimatedMinutes,
        scheduled_date: scheduledDate || null,
        scheduled_start: scheduledStart || null,
      })
    }
    onClose()
  }

  return (
    <div
      className="fixed inset-0 z-20 flex items-center justify-center bg-crust/60 p-4"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose()
      }}
    >
      <motion.form
        initial={{ opacity: 0, scale: 0.96 }}
        animate={{ opacity: 1, scale: 1 }}
        onSubmit={handleSubmit}
        className="flex max-h-[90vh] w-full max-w-lg flex-col gap-4 overflow-y-auto rounded-2xl border border-surface0 bg-mantle p-6"
      >
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-text">{isEdit ? 'Edit task' : 'New task'}</h2>
          <button type="button" onClick={onClose} className="text-subtext0 hover:text-text">
            <X size={18} />
          </button>
        </div>

        <input
          autoFocus
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Task title"
          className="rounded-lg border border-surface0 bg-base px-3 py-2 text-text outline-none focus:border-accent"
        />
        <textarea
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          placeholder="Notes (optional)"
          rows={2}
          className="rounded-lg border border-surface0 bg-base px-3 py-2 text-sm text-text outline-none focus:border-accent"
        />

        <div className="grid grid-cols-2 gap-3">
          <Field label="Priority">
            <select
              value={priority}
              onChange={(e) => setPriority(e.target.value as Priority)}
              className="w-full rounded-lg border border-surface0 bg-base px-2 py-1.5 text-sm text-text"
            >
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="urgent">Urgent</option>
            </select>
          </Field>
          <Field label="Category">
            <select
              value={categoryId}
              onChange={(e) => setCategoryId(e.target.value)}
              className="w-full rounded-lg border border-surface0 bg-base px-2 py-1.5 text-sm text-text"
            >
              <option value="">None</option>
              {categories.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </Field>
          <Field label="Estimate (min)">
            <input
              type="number"
              min={0}
              value={estimatedMinutes}
              onChange={(e) => setEstimatedMinutes(Number(e.target.value))}
              className="w-full rounded-lg border border-surface0 bg-base px-2 py-1.5 text-sm text-text"
            />
          </Field>
          <Field label="Start time">
            <input
              type="time"
              value={scheduledStart}
              onChange={(e) => setScheduledStart(e.target.value)}
              className="w-full rounded-lg border border-surface0 bg-base px-2 py-1.5 text-sm text-text"
            />
          </Field>
          <Field label="Date">
            <input
              type="date"
              value={scheduledDate}
              onChange={(e) => setScheduledDate(e.target.value)}
              className="w-full rounded-lg border border-surface0 bg-base px-2 py-1.5 text-sm text-text"
            />
          </Field>
        </div>

        {!isEdit && (
          <div className="rounded-xl border border-surface0 p-3">
            <p className="mb-2 text-xs font-medium uppercase tracking-wide text-subtext0">Repeat</p>
            <div className="mb-3 flex gap-2">
              {(['none', 'daily', 'weekly', 'monthly'] as RepeatMode[]).map((m) => (
                <button
                  key={m}
                  type="button"
                  onClick={() => setRepeatMode(m)}
                  className={`rounded-lg border px-2.5 py-1 text-xs capitalize ${
                    repeatMode === m ? 'border-accent bg-accent/15 text-accent' : 'border-surface1 text-subtext0'
                  }`}
                >
                  {m}
                </button>
              ))}
            </div>

            {repeatMode !== 'none' && (
              <div className="flex flex-col gap-3">
                <Field label="Every N">
                  <input
                    type="number"
                    min={1}
                    value={interval}
                    onChange={(e) => setIntervalValue(Number(e.target.value))}
                    className="w-20 rounded-lg border border-surface0 bg-base px-2 py-1.5 text-sm text-text"
                  />
                </Field>

                {repeatMode === 'weekly' && (
                  <div className="flex gap-1">
                    {WEEKDAYS.map((d) => (
                      <button
                        key={d.n}
                        type="button"
                        onClick={() => toggleWeekday(d.n)}
                        className={`h-8 w-10 rounded-lg border text-xs ${
                          weekdays.includes(d.n)
                            ? 'border-accent bg-accent/15 text-accent'
                            : 'border-surface1 text-subtext0'
                        }`}
                      >
                        {d.label}
                      </button>
                    ))}
                  </div>
                )}

                {repeatMode === 'monthly' && (
                  <Field label="Day of month">
                    <input
                      type="number"
                      min={1}
                      max={31}
                      value={dayOfMonth}
                      onChange={(e) => setDayOfMonth(Number(e.target.value))}
                      className="w-20 rounded-lg border border-surface0 bg-base px-2 py-1.5 text-sm text-text"
                    />
                  </Field>
                )}

                <Field label="End date (optional)">
                  <input
                    type="date"
                    value={endDate}
                    onChange={(e) => setEndDate(e.target.value)}
                    className="w-full rounded-lg border border-surface0 bg-base px-2 py-1.5 text-sm text-text"
                  />
                </Field>
              </div>
            )}
          </div>
        )}

        <div className="flex justify-between gap-2">
          {isEdit && task && (
            <button
              type="button"
              onClick={async () => {
                await deleteTask.mutateAsync(task.id)
                onClose()
              }}
              className="rounded-lg px-3 py-2 text-sm text-red hover:bg-red/10"
            >
              Delete
            </button>
          )}
          <div className="ml-auto flex gap-2">
            <button type="button" onClick={onClose} className="rounded-lg px-3 py-2 text-sm text-subtext0 hover:bg-surface0">
              Cancel
            </button>
            <button type="submit" className="rounded-lg bg-accent px-4 py-2 text-sm font-medium text-base hover:opacity-90">
              {isEdit ? 'Save' : 'Create'}
            </button>
          </div>
        </div>
      </motion.form>
    </div>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="flex flex-col gap-1">
      <span className="text-xs text-subtext0">{label}</span>
      {children}
    </label>
  )
}
