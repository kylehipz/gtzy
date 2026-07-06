import { useState } from 'react'
import { Pause, Play, Plus, Trash2 } from 'lucide-react'
import { ThemeSwitcher } from '../components/ThemeSwitcher'
import { ACCENTS } from '../theme/ThemeProvider'
import {
  useCategories,
  useCreateCategory,
  useDeleteCategory,
  useDeleteRecurrence,
  useRecurrences,
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
  const createCategory = useCreateCategory()
  const deleteCategory = useDeleteCategory()
  const [newCategoryName, setNewCategoryName] = useState('')
  const [newCategoryColor, setNewCategoryColor] = useState<string>('mauve')

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
        <form
          onSubmit={async (e) => {
            e.preventDefault()
            if (!newCategoryName.trim()) return
            await createCategory.mutateAsync({ name: newCategoryName.trim(), color: newCategoryColor })
            setNewCategoryName('')
          }}
          className="flex flex-wrap items-center gap-2"
        >
          <input
            value={newCategoryName}
            onChange={(e) => setNewCategoryName(e.target.value)}
            placeholder="New category name"
            className="rounded-lg border border-surface0 bg-mantle px-3 py-1.5 text-sm text-text outline-none focus:border-accent"
          />
          <select
            value={newCategoryColor}
            onChange={(e) => setNewCategoryColor(e.target.value)}
            className="rounded-lg border border-surface0 bg-mantle px-2 py-1.5 text-sm text-text"
          >
            {ACCENTS.map((a) => (
              <option key={a} value={a}>
                {a}
              </option>
            ))}
          </select>
          <button type="submit" className="flex items-center gap-1.5 rounded-lg bg-accent px-3 py-1.5 text-sm font-medium text-base">
            <Plus size={14} /> Add
          </button>
        </form>

        <div className="flex flex-col gap-2">
          {categories.map((c) => (
            <div key={c.id} className="flex items-center justify-between rounded-lg border border-surface0 bg-mantle p-2.5">
              <div className="flex items-center gap-2">
                <span className="h-3 w-3 rounded-full" style={{ backgroundColor: `var(--ctp-${c.color})` }} />
                <span className="text-sm text-text">{c.name}</span>
              </div>
              <button type="button" onClick={() => deleteCategory.mutate(c.id)} className="text-subtext0 hover:text-red">
                <Trash2 size={14} />
              </button>
            </div>
          ))}
        </div>
      </Section>

      <Section title="Recurring tasks">
        {recurrences.length === 0 ? (
          <p className="text-sm text-subtext0">No recurring rules yet — create one from the New Task form.</p>
        ) : (
          <div className="flex flex-col gap-2">
            {recurrences.map((r) => (
              <div key={r.id} className="flex items-center justify-between rounded-lg border border-surface0 bg-mantle p-2.5">
                <div>
                  <p className={`text-sm font-medium text-text ${!r.active ? 'opacity-50' : ''}`}>{r.title}</p>
                  <p className="text-xs text-subtext0">{repeatSummary(r)}</p>
                </div>
                <div className="flex gap-1">
                  <button
                    type="button"
                    title={r.active ? 'Pause' : 'Resume'}
                    onClick={() => updateRecurrence.mutate({ id: r.id, patch: { active: !r.active } })}
                    className="flex h-8 w-8 items-center justify-center rounded-lg text-subtext0 hover:bg-surface0 hover:text-text"
                  >
                    {r.active ? <Pause size={14} /> : <Play size={14} />}
                  </button>
                  <button
                    type="button"
                    title="Delete rule + future instances"
                    onClick={() => deleteRecurrence.mutate({ id: r.id, hard: true })}
                    className="flex h-8 w-8 items-center justify-center rounded-lg text-subtext0 hover:bg-surface0 hover:text-red"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              </div>
            ))}
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
