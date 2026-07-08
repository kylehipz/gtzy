import { useState } from 'react'
import { Pause, Play, Plus, Repeat, Trash2 } from 'lucide-react'
import { ThemeSwitcher } from '../components/ThemeSwitcher'
import { ColorSwatch } from '../components/ColorSwatch'
import { CategoryBadge } from '../components/CategoryBadge'
import { PriorityBadge } from '../components/PriorityBadge'
import { IconButton } from '../components/TaskCard'
import { Modal } from '../components/Modal'
import { ACCENTS } from '../theme/ThemeProvider'
import {
  useCategories,
  useCreateCategory,
  useDeleteCategory,
  useDeleteRecurrence,
  useRecurrences,
  useUpdateCategory,
  useUpdateRecurrence,
} from '../lib/queries'

function repeatSummary(r: { freq: string; interval: number; days_of_week: string; day_of_month: number | null }): string {
  if (r.freq === 'daily') return r.interval === 1 ? 'Daily' : `Every ${r.interval} days`
  if (r.freq === 'weekly') {
    const days = r.days_of_week
      .split(',')
      .filter(Boolean)
      .map((n) => ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'][Number(n)])
      .join('/')
    return r.interval === 1 ? `Weekly on ${days}` : `Every ${r.interval} weeks on ${days}`
  }
  if (r.freq === 'monthly') {
    return r.interval === 1 ? `Monthly on day ${r.day_of_month}` : `Every ${r.interval} months on day ${r.day_of_month}`
  }
  return r.freq
}

export function Settings() {
  const { data: categories = [] } = useCategories()
  const categoryById = new Map(categories.map((c) => [c.id, c]))
  const createCategory = useCreateCategory()
  const updateCategory = useUpdateCategory()
  const deleteCategory = useDeleteCategory()
  const [newCategoryName, setNewCategoryName] = useState('')
  const [newCategoryColor, setNewCategoryColor] = useState<string>('mauve')
  const [showAddCategory, setShowAddCategory] = useState(false)

  const { data: recurrences = [] } = useRecurrences()
  const updateRecurrence = useUpdateRecurrence()
  const deleteRecurrence = useDeleteRecurrence()

  return (
    <div className="flex flex-col gap-8 p-6">
      <h1 className="text-xl font-semibold text-text">Settings</h1>

      <Section title="Theme">
        <ThemeSwitcher />
      </Section>

      <Section title="Categories">
        <button
          type="button"
          onClick={() => setShowAddCategory(true)}
          className="flex items-center gap-1.5 self-start rounded-lg bg-accent px-3 py-1.5 text-sm font-medium text-base hover:opacity-90"
        >
          <Plus size={14} /> Add category
        </button>

        <div className="flex flex-col gap-2">
          {categories.map((c) => (
            <div
              key={c.id}
              className="flex flex-wrap items-center gap-2 rounded-lg border border-surface0 bg-mantle p-2.5"
            >
              <span className="text-sm text-text">{c.name}</span>
              <div className="ml-auto flex flex-wrap justify-end gap-1">
                {ACCENTS.map((a) => (
                  <ColorSwatch
                    key={a}
                    color={a}
                    size="sm"
                    selected={c.color === a}
                    onClick={() => updateCategory.mutate({ id: c.id, patch: { color: a } })}
                  />
                ))}
              </div>
              <IconButton onClick={() => deleteCategory.mutate(c.id)} label="Delete category" variant="red">
                <Trash2 size={14} />
              </IconButton>
            </div>
          ))}
        </div>
      </Section>

      {showAddCategory && (
        <Modal title="Add category" onClose={() => setShowAddCategory(false)}>
          <form
            onSubmit={async (e) => {
              e.preventDefault()
              if (!newCategoryName.trim()) return
              await createCategory.mutateAsync({ name: newCategoryName.trim(), color: newCategoryColor })
              setNewCategoryName('')
              setShowAddCategory(false)
            }}
            className="flex flex-col gap-4"
          >
            <input
              autoFocus
              value={newCategoryName}
              onChange={(e) => setNewCategoryName(e.target.value)}
              placeholder="Category name"
              className="rounded-lg border border-surface0 bg-base px-3 py-2 text-sm text-text outline-none focus:border-accent"
            />
            <div className="flex flex-wrap gap-2">
              {ACCENTS.map((a) => (
                <ColorSwatch key={a} color={a} selected={newCategoryColor === a} onClick={() => setNewCategoryColor(a)} />
              ))}
            </div>
            <button
              type="submit"
              className="self-start rounded-lg bg-accent px-3 py-1.5 text-sm font-medium text-base hover:opacity-90"
            >
              Add
            </button>
          </form>
        </Modal>
      )}

      <Section title="Recurring tasks">
        {recurrences.length === 0 ? (
          <p className="text-sm text-subtext0">No recurring rules yet — create one from the New Task form.</p>
        ) : (
          <div className="flex flex-col gap-2">
            {recurrences.map((r) => {
              const category = r.category_id ? categoryById.get(r.category_id) : null
              const tint = category
                ? {
                    backgroundColor: `color-mix(in oklab, var(--ctp-${category.color}) 12%, var(--ctp-mantle))`,
                    borderColor: `color-mix(in oklab, var(--ctp-${category.color}) 35%, var(--ctp-surface0))`,
                  }
                : undefined
              return (
              <div
                key={r.id}
                style={tint}
                className={`flex items-center gap-3 rounded-xl border p-3 ${category ? '' : 'border-surface0 bg-mantle'}`}
              >
                <Repeat size={16} className="shrink-0 text-overlay0" />
                <div className="min-w-0 flex-1">
                  <p className={`truncate text-sm font-medium text-text ${!r.active ? 'opacity-50' : ''}`}>{r.title}</p>
                  <div className="mt-1 flex flex-wrap items-center gap-1.5">
                    <PriorityBadge priority={r.priority} />
                    <CategoryBadge category={category} />
                    <span className="text-xs text-subtext0">{repeatSummary(r)}</span>
                  </div>
                </div>
                <div className="flex shrink-0 gap-1">
                  <IconButton
                    onClick={() => updateRecurrence.mutate({ id: r.id, patch: { active: !r.active } })}
                    label={r.active ? 'Pause' : 'Resume'}
                    variant={r.active ? 'peach' : 'green'}
                  >
                    {r.active ? <Pause size={14} /> : <Play size={14} />}
                  </IconButton>
                  <IconButton
                    onClick={() => deleteRecurrence.mutate({ id: r.id, hard: true })}
                    label="Delete rule + future instances"
                    variant="red"
                  >
                    <Trash2 size={14} />
                  </IconButton>
                </div>
              </div>
              )
            })}
          </div>
        )}
      </Section>

      <Section title="Server">
        <p className="text-sm text-subtext0">
          API base: <code className="rounded bg-surface0 px-1.5 py-0.5 text-xs">{window.location.origin}/api</code>
        </p>
      </Section>
    </div>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="flex flex-col gap-3">
      <h2 className="text-sm font-semibold uppercase tracking-wide text-subtext0">{title}</h2>
      {children}
    </section>
  )
}
