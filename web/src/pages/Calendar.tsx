import { useState } from 'react'
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
  const [selectedDate, setSelectedDate] = useState<string | null>(null)
  const [editingTask, setEditingTask] = useState<Task | null>(null)

  const { data: days = [] } = useMonth(year, month)
  const { data: categories = [] } = useCategories()
  const { data: dayTasks = [] } = useTasks({ date: selectedDate ?? '' })
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
          <button type="button" onClick={() => setSelectedDate(null)} className="text-sm text-accent">
            Back to month
          </button>
        )}
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
          <MonthGrid year={year} month={month} days={days} onSelectDay={setSelectedDate} />
        </>
      ) : (
        <>
          <h2 className="text-lg font-medium text-text">{selectedDate}</h2>
          <DayTimeline tasks={dayTasks} categoryById={categoryById} onSelectTask={setEditingTask} />
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
