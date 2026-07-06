import { useState } from 'react'
import { NotebookPen } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import { JournalEditor } from '../components/JournalEditor'
import { EmptyState } from '../components/EmptyState'
import { useDeleteJournal, useGenerateJournalSummary, useJournal, useSummary } from '../lib/queries'
import { addDaysLocal, todayLocal } from '../lib/time'
import type { JournalEntry } from '../lib/types'

export function Journal() {
  const [date, setDate] = useState(todayLocal())
  const [editing, setEditing] = useState<JournalEntry | null>(null)
  const [from, setFrom] = useState(addDaysLocal(todayLocal(), -6))
  const [to, setTo] = useState(todayLocal())
  const { data: allEntries = [] } = useJournal({ from, to })
  const deleteJournal = useDeleteJournal()

  const summaryKey = `${from}..${to}`
  const { data: summaryData } = useSummary('journal', summaryKey)
  const generateJournalSummary = useGenerateJournalSummary()

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

      <div className="flex flex-col gap-3 rounded-xl border border-surface0 bg-mantle p-4">
        <div className="flex items-center justify-between gap-4">
          <p className="text-sm font-medium text-text">AI Summary</p>
          <div className="flex items-center gap-2">
            <label className="flex items-center gap-1.5 text-xs text-subtext0">
              From
              <input
                type="date"
                value={from}
                onChange={(e) => setFrom(e.target.value)}
                className="rounded-lg border border-surface0 bg-base px-2 py-1 text-sm text-text"
              />
            </label>
            <label className="flex items-center gap-1.5 text-xs text-subtext0">
              To
              <input
                type="date"
                value={to}
                onChange={(e) => setTo(e.target.value)}
                className="rounded-lg border border-surface0 bg-base px-2 py-1 text-sm text-text"
              />
            </label>
          </div>
        </div>

        {!summaryData?.enabled ? (
          <p className="text-sm text-subtext0">Add ANTHROPIC_API_KEY to enable AI-generated journal summaries.</p>
        ) : (
          <div className="flex flex-col gap-2">
            {summaryData.summary ? (
              <div className="prose prose-invert max-w-none text-sm text-text">
                <ReactMarkdown>{summaryData.summary.content}</ReactMarkdown>
              </div>
            ) : (
              <p className="text-sm text-subtext0">No summary generated yet for this range.</p>
            )}
            <button
              type="button"
              onClick={() => generateJournalSummary.mutate({ from, to })}
              disabled={generateJournalSummary.isPending}
              className="self-start rounded-lg bg-accent px-3 py-1.5 text-sm font-medium text-base hover:opacity-90 disabled:opacity-50"
            >
              {generateJournalSummary.isPending ? 'Generating…' : summaryData.summary ? 'Regenerate' : 'Generate summary'}
            </button>
          </div>
        )}
      </div>

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
