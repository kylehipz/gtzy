import type { Category } from '../lib/types'

export interface Filters {
  date: string
  status: string
  priority: string
  category_id: string
}

export function FilterBar({
  filters,
  onChange,
  categories,
}: {
  filters: Filters
  onChange: (f: Filters) => void
  categories: Category[]
}) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      <input
        type="date"
        value={filters.date}
        onChange={(e) => onChange({ ...filters, date: e.target.value })}
        className="rounded-lg border border-surface0 bg-mantle px-2 py-1.5 text-sm text-text"
      />
      <select
        value={filters.status}
        onChange={(e) => onChange({ ...filters, status: e.target.value })}
        className="rounded-lg border border-surface0 bg-mantle px-2 py-1.5 text-sm text-text"
      >
        <option value="">All statuses</option>
        <option value="todo">Todo</option>
        <option value="in_progress">In progress</option>
        <option value="paused">Paused</option>
        <option value="done">Done</option>
      </select>
      <select
        value={filters.priority}
        onChange={(e) => onChange({ ...filters, priority: e.target.value })}
        className="rounded-lg border border-surface0 bg-mantle px-2 py-1.5 text-sm text-text"
      >
        <option value="">All priorities</option>
        <option value="urgent">Urgent</option>
        <option value="high">High</option>
        <option value="medium">Medium</option>
        <option value="low">Low</option>
      </select>
      <select
        value={filters.category_id}
        onChange={(e) => onChange({ ...filters, category_id: e.target.value })}
        className="rounded-lg border border-surface0 bg-mantle px-2 py-1.5 text-sm text-text"
      >
        <option value="">All categories</option>
        {categories.map((c) => (
          <option key={c.id} value={c.id}>
            {c.name}
          </option>
        ))}
      </select>
    </div>
  )
}
