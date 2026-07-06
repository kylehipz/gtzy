import { useEffect, useState } from 'react'
import { GripVertical, ListTodo, Plus, Search } from 'lucide-react'
import { Reorder, useDragControls } from 'framer-motion'
import { FilterBar, type Filters } from '../components/FilterBar'
import { TaskCard } from '../components/TaskCard'
import { TaskForm } from '../components/TaskForm'
import { EmptyState } from '../components/EmptyState'
import { useCategories, useTasks, useUpdateTask } from '../lib/queries'
import { todayLocal } from '../lib/time'
import type { Category, Task } from '../lib/types'

export function Today() {
  const [filters, setFilters] = useState<Filters>({ date: todayLocal(), status: [], priority: [], category_id: [] })
  const [search, setSearch] = useState('')
  const [editingTask, setEditingTask] = useState<Task | null>(null)
  const [showNewForm, setShowNewForm] = useState(false)

  const { data: categories = [] } = useCategories()
  const { data: allTasks = [], isLoading } = useTasks({ date: filters.date })
  const updateTask = useUpdateTask()

  const categoryById = new Map(categories.map((c) => [c.id, c]))

  const searchTerm = search.trim().toLowerCase()
  const tasks = allTasks.filter(
    (t) =>
      (filters.status.length === 0 || filters.status.includes(t.status)) &&
      (filters.priority.length === 0 || filters.priority.includes(t.priority)) &&
      (filters.category_id.length === 0 || (t.category_id != null && filters.category_id.includes(String(t.category_id)))) &&
      (searchTerm === '' || t.title.toLowerCase().includes(searchTerm) || t.notes.toLowerCase().includes(searchTerm)),
  )

  const [orderedIds, setOrderedIds] = useState<number[]>([])
  const taskIdsKey = tasks.map((t) => t.id).join(',')
  useEffect(() => {
    setOrderedIds(tasks.map((t) => t.id))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [taskIdsKey])

  const taskById = new Map(tasks.map((t) => [t.id, t]))
  const orderedTasks = orderedIds.map((id) => taskById.get(id)).filter((t): t is Task => t != null)

  function persistReorder(newOrder: Task[]) {
    setOrderedIds(newOrder.map((t) => t.id))
    newOrder.forEach((t, i) => {
      if (t.sort_order !== i) updateTask.mutate({ id: t.id, patch: { sort_order: i } })
    })
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

      <div className="relative">
        <Search size={16} className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-subtext0" />
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search tasks…"
          className="w-full rounded-lg border border-surface0 bg-mantle py-2 pl-9 pr-3 text-sm text-text outline-none focus:border-accent"
        />
      </div>

      {isLoading ? (
        <p className="text-sm text-subtext0">Loading...</p>
      ) : orderedTasks.length === 0 ? (
        <EmptyState icon={ListTodo} title="Nothing scheduled" description="Add a task above or create one with full options." />
      ) : (
        <Reorder.Group as="div" axis="y" values={orderedTasks} onReorder={persistReorder} className="flex flex-col gap-2">
          {orderedTasks.map((t) => (
            <SortableTaskRow
              key={t.id}
              task={t}
              category={t.category_id ? categoryById.get(t.category_id) : null}
              onEdit={() => setEditingTask(t)}
            />
          ))}
        </Reorder.Group>
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

function SortableTaskRow({
  task,
  category,
  onEdit,
}: {
  task: Task
  category: Category | null | undefined
  onEdit: () => void
}) {
  const dragControls = useDragControls()
  return (
    <Reorder.Item as="div" value={task} dragListener={false} dragControls={dragControls}>
      <div className="flex items-center gap-2">
        <button
          type="button"
          onPointerDown={(e) => dragControls.start(e)}
          className="flex h-8 w-8 shrink-0 cursor-grab items-center justify-center text-subtext0 active:cursor-grabbing"
          aria-label="Reorder"
        >
          <GripVertical size={14} />
        </button>
        <div className="min-w-0 flex-1">
          <TaskCard task={task} category={category} onEdit={onEdit} />
        </div>
      </div>
    </Reorder.Item>
  )
}
