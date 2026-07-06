import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { MonthGrid } from '../components/MonthGrid'
import { DayTimeline } from '../components/DayTimeline'
import { TaskForm } from '../components/TaskForm'
import { useCategories, useMonth, useTasks } from '../lib/queries'
import { todayLocal } from '../lib/time'
import type { Task } from '../lib/types'

export function Calendar() {
  const today = new Date()
  const [year, setYear] = useState(today.getFullYear())
  const [month, setMonth] = useState(today.getMonth() + 1)
  const { date: selectedDate } = useParams<{ date?: string }>()
  const navigate = useNavigate()
  const [editingTask, setEditingTask] = useState<Task | null>(null)
  const [categoryFilter, setCategoryFilter] = useState('')

  const { data: days = [] } = useMonth(year, month, categoryFilter || undefined)
  const { data: categories = [] } = useCategories()
  const { data: dayTasks = [] } = useTasks({ date: selectedDate ?? '', category_id: categoryFilter || undefined })
  const categoryById = new Map(categories.map((c) => [c.id, c]))

  function shiftMonth(delta: number) {
    let m = month + delta
    let y = year
    if (m > 12) {
      m = 1
      y++
    } else if (m < 1) {
      m = 12
      y--
    }
    setMonth(m)
    setYear(y)
  }

  const monthLabel = new Date(year, month - 1, 1).toLocaleDateString(undefined, { month: 'long', year: 'numeric' })

  return (
    <div className="flex flex-col gap-4 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold text-text">Calendar</h1>
        {selectedDate && (
          <button type="button" onClick={() => navigate('/calendar')} className="text-sm text-accent">
            Back to month
          </button>
        )}
      </div>

      <div className="flex flex-wrap gap-1.5">
        <button
          type="button"
          onClick={() => setCategoryFilter('')}
          className={`rounded-lg border px-2.5 py-1 text-xs transition-colors ${
            categoryFilter === '' ? 'border-accent bg-accent/15 text-accent' : 'border-surface0 text-subtext0 hover:border-surface1'
          }`}
        >
          All categories
        </button>
        {categories.map((c) => (
          <button
            key={c.id}
            type="button"
            onClick={() => setCategoryFilter(String(c.id))}
            className={`flex items-center gap-1.5 rounded-lg border px-2.5 py-1 text-xs transition-colors ${
              categoryFilter === String(c.id)
                ? 'border-accent bg-accent/15 text-accent'
                : 'border-surface0 text-subtext0 hover:border-surface1'
            }`}
          >
            <span className="h-2 w-2 rounded-full" style={{ backgroundColor: `var(--ctp-${c.color})` }} />
            {c.name}
          </button>
        ))}
      </div>

      {!selectedDate ? (
        <>
          <div className="flex items-center justify-center gap-4">
            <button type="button" onClick={() => shiftMonth(-1)} className="rounded-lg p-1.5 hover:bg-surface0">
              <ChevronLeft size={18} />
            </button>
            <span className="font-medium text-text">{monthLabel}</span>
            <button type="button" onClick={() => shiftMonth(1)} className="rounded-lg p-1.5 hover:bg-surface0">
              <ChevronRight size={18} />
            </button>
          </div>
          <MonthGrid year={year} month={month} days={days} onSelectDay={(d) => navigate(`/calendar/${d}`)} />
        </>
      ) : (
        <>
          <h2 className="text-lg font-medium text-text">{selectedDate}</h2>
          <DayTimeline tasks={dayTasks} categoryById={categoryById} onSelectTask={setEditingTask} date={selectedDate} />
        </>
      )}

      {editingTask && (
        <TaskForm
          task={editingTask}
          categories={categories}
          defaultDate={selectedDate ?? todayLocal()}
          onClose={() => setEditingTask(null)}
        />
      )}
    </div>
  )
}
