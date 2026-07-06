import { useState } from 'react'
import { NotebookPen } from 'lucide-react'
import { JournalEditor } from '../components/JournalEditor'
import { EmptyState } from '../components/EmptyState'
import { useDeleteJournal, useJournal } from '../lib/queries'
import { todayLocal } from '../lib/time'
import type { JournalEntry } from '../lib/types'

export function Journal() {
  const [date, setDate] = useState(todayLocal())
  const [editing, setEditing] = useState<JournalEntry | null>(null)
  const { data: allEntries = [] } = useJournal()
  const deleteJournal = useDeleteJournal()

  return (
    <div className="flex flex-col gap-4 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold text-text">Journal</h1>
        <div className="flex items-center gap-2">
          {editing && (
            <button type="button" onClick={() => setEditing(null)} className="text-sm text-accent">
              New entry
            </button>
          )}
          <input
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            className="rounded-lg border border-surface0 bg-mantle px-2 py-1.5 text-sm text-text"
          />
        </div>
      </div>

      <JournalEditor
        key={editing?.id ?? 'new'}
        date={editing?.date ?? date}
        entry={editing}
        onSaved={() => setEditing(null)}
      />

      <div>
        <p className="mb-2 text-xs font-medium uppercase tracking-wide text-subtext0">All entries</p>
        {allEntries.length === 0 ? (
          <EmptyState icon={NotebookPen} title="No journal entries yet" description="Write your first entry above." />
        ) : (
          <div className="flex flex-col gap-2">
            {allEntries.map((e) => (
              <div key={e.id} className="flex items-center gap-2 rounded-lg border border-surface0 bg-mantle p-3 hover:border-surface1">
                <button type="button" onClick={() => setEditing(e)} className="min-w-0 flex-1 text-left">
                  <div className="flex items-center justify-between">
                    <span className="font-medium text-text">{e.title || e.date}</span>
                    <span className="text-xs text-subtext0">{e.date}</span>
                  </div>
                  <p className="mt-1 truncate text-sm text-subtext0">{e.content}</p>
                </button>
                <button
                  type="button"
                  onClick={() => {
                    deleteJournal.mutate(e.id)
                    if (editing?.id === e.id) setEditing(null)
                  }}
                  className="shrink-0 rounded-lg px-2 py-1 text-xs text-red hover:bg-red/10"
                >
                  Delete
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
