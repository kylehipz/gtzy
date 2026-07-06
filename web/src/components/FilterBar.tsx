import type { Category } from '../lib/types'

export interface Filters {
  date: string
  status: string[]
  priority: string[]
  category_id: string[]
}

const STATUS_OPTIONS = [
  { value: 'todo', label: 'Todo' },
  { value: 'in_progress', label: 'In progress' },
  { value: 'paused', label: 'Paused' },
  { value: 'done', label: 'Done' },
]

const PRIORITY_OPTIONS = [
  { value: 'urgent', label: 'Urgent', color: 'red' },
  { value: 'high', label: 'High', color: 'peach' },
  { value: 'medium', label: 'Medium', color: 'yellow' },
  { value: 'low', label: 'Low', color: 'teal' },
]

function toggle(values: string[], value: string): string[] {
  return values.includes(value) ? values.filter((v) => v !== value) : [...values, value]
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
    <div className="flex flex-col items-start gap-3">
      <input
        type="date"
        value={filters.date}
        onChange={(e) => onChange({ ...filters, date: e.target.value })}
        className="rounded-lg border border-surface0 bg-mantle px-2 py-1.5 text-sm text-text"
      />

      <div className="inline-flex flex-wrap items-center gap-1 rounded-lg border border-surface0 bg-mantle p-1">
        <ToggleGroup
          options={STATUS_OPTIONS}
          selected={filters.status}
          onToggle={(v) => onChange({ ...filters, status: toggle(filters.status, v) })}
        />
      </div>

      <div className="inline-flex flex-wrap items-center gap-1 rounded-lg border border-surface0 bg-mantle p-1">
        <ToggleGroup
          options={PRIORITY_OPTIONS}
          selected={filters.priority}
          onToggle={(v) => onChange({ ...filters, priority: toggle(filters.priority, v) })}
        />
      </div>

      <div className="inline-flex flex-wrap items-center gap-1 rounded-lg border border-surface0 bg-mantle p-1">
        {categories.map((c) => {
          const active = filters.category_id.includes(String(c.id))
          return (
            <button
              key={c.id}
              type="button"
              onClick={() => onChange({ ...filters, category_id: toggle(filters.category_id, String(c.id)) })}
              className={`flex items-center gap-1.5 rounded-lg border px-2.5 py-1 text-sm transition-colors ${
                active ? 'border-accent bg-accent/15 text-accent' : 'border-surface0 text-subtext0 hover:border-surface1'
              }`}
            >
              <span className="h-2.5 w-2.5 rounded-full" style={{ backgroundColor: `var(--ctp-${c.color})` }} />
              {c.name}
            </button>
          )
        })}
      </div>
    </div>
  )
}

function ToggleGroup({
  options,
  selected,
  onToggle,
}: {
  options: { value: string; label: string; color?: string }[]
  selected: string[]
  onToggle: (value: string) => void
}) {
  return (
    <div className="flex flex-wrap gap-1">
      {options.map((o) => {
        const active = selected.includes(o.value)
        const activeColorStyle =
          active && o.color
            ? {
                borderColor: `var(--ctp-${o.color})`,
                color: `var(--ctp-${o.color})`,
                backgroundColor: `color-mix(in oklab, var(--ctp-${o.color}) 15%, transparent)`,
              }
            : undefined
        return (
          <button
            key={o.value}
            type="button"
            onClick={() => onToggle(o.value)}
            style={activeColorStyle}
            className={`rounded-lg border px-2.5 py-1 text-sm transition-colors ${
              active && o.color
                ? ''
                : active
                  ? 'border-accent bg-accent/15 text-accent'
                  : 'border-surface0 text-subtext0 hover:border-surface1'
            }`}
          >
            {o.color ? (
              <span className="flex items-center gap-1.5">
                <span className="h-2 w-2 rounded-full" style={{ backgroundColor: `var(--ctp-${o.color})` }} />
                {o.label}
              </span>
            ) : (
              o.label
            )}
          </button>
        )
      })}
    </div>
  )
}
