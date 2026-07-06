import { useState } from 'react'
import { NotebookPen } from 'lucide-react'
import { JournalEditor } from '../components/JournalEditor'
import { EmptyState } from '../components/EmptyState'
import { useJournal } from '../lib/queries'
import { todayLocal } from '../lib/time'

export function Journal() {
  const [date, setDate] = useState(todayLocal())
  const { data: entries = [] } = useJournal({ date })
  const { data: allEntries = [] } = useJournal()

  return (
    <div className="flex flex-col gap-4 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold text-text">Journal</h1>
        <input
          type="date"
          value={date}
          onChange={(e) => setDate(e.target.value)}
          className="rounded-lg border border-surface0 bg-mantle px-2 py-1.5 text-sm text-text"
        />
      </div>

      <JournalEditor date={date} entry={entries[0] ?? null} />

      <div>
        <p className="mb-2 text-xs font-medium uppercase tracking-wide text-subtext0">All entries</p>
        {allEntries.length === 0 ? (
          <EmptyState icon={NotebookPen} title="No journal entries yet" description="Write your first entry above." />
        ) : (
          <div className="flex flex-col gap-2">
            {allEntries.map((e) => (
              <button
                key={e.id}
                type="button"
                onClick={() => setDate(e.date)}
                className="rounded-lg border border-surface0 bg-mantle p-3 text-left hover:border-surface1"
              >
                <div className="flex items-center justify-between">
                  <span className="font-medium text-text">{e.title || e.date}</span>
                  <span className="text-xs text-subtext0">{e.date}</span>
                </div>
                <p className="mt-1 truncate text-sm text-subtext0">{e.content}</p>
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
