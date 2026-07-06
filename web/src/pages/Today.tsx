import { useState } from 'react'
import { ListTodo, Plus } from 'lucide-react'
import { AnimatePresence } from 'framer-motion'
import { FilterBar, type Filters } from '../components/FilterBar'
import { TaskCard } from '../components/TaskCard'
import { TaskForm } from '../components/TaskForm'
import { EmptyState } from '../components/EmptyState'
import { useCategories, useCreateTask, useTasks } from '../lib/queries'
import { todayLocal } from '../lib/time'
import type { Task } from '../lib/types'

export function Today() {
  const [filters, setFilters] = useState<Filters>({ date: todayLocal(), status: '', priority: '', category_id: '' })
  const [quickAdd, setQuickAdd] = useState('')
  const [editingTask, setEditingTask] = useState<Task | null>(null)
  const [showNewForm, setShowNewForm] = useState(false)

  const { data: categories = [] } = useCategories()
  const { data: tasks = [], isLoading } = useTasks(filters)
  const createTask = useCreateTask()

  const categoryById = new Map(categories.map((c) => [c.id, c]))

  async function handleQuickAdd(e: React.FormEvent) {
    e.preventDefault()
    if (!quickAdd.trim()) return
    await createTask.mutateAsync({ title: quickAdd.trim(), scheduled_date: filters.date })
    setQuickAdd('')
  }

  return (
    <div className="flex flex-col gap-4 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold text-text">Today</h1>
        <button
          type="button"
          onClick={() => setShowNewForm(true)}
          className="flex items-center gap-1.5 rounded-lg bg-accent px-3 py-1.5 text-sm font-medium text-base hover:opacity-90"
        >
          <Plus size={16} /> New task
        </button>
      </div>

      <FilterBar filters={filters} onChange={setFilters} categories={categories} />

      <form onSubmit={handleQuickAdd} className="flex gap-2">
        <input
          value={quickAdd}
          onChange={(e) => setQuickAdd(e.target.value)}
          placeholder="Quick add a task and press Enter..."
          className="flex-1 rounded-lg border border-surface0 bg-mantle px-3 py-2 text-sm text-text outline-none focus:border-accent"
        />
      </form>

      {isLoading ? (
        <p className="text-sm text-subtext0">Loading...</p>
      ) : tasks.length === 0 ? (
        <EmptyState icon={ListTodo} title="Nothing scheduled" description="Add a task above or create one with full options." />
      ) : (
        <div className="flex flex-col gap-2">
          <AnimatePresence initial={false}>
            {tasks.map((t) => (
              <TaskCard key={t.id} task={t} category={t.category_id ? categoryById.get(t.category_id) : null} onEdit={() => setEditingTask(t)} />
            ))}
          </AnimatePresence>
        </div>
      )}

      {(showNewForm || editingTask) && (
        <TaskForm
          task={editingTask}
          categories={categories}
          defaultDate={filters.date}
          onClose={() => {
            setShowNewForm(false)
            setEditingTask(null)
          }}
        />
      )}
    </div>
  )
}
