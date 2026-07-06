import { useState } from 'react'
import ReactMarkdown from 'react-markdown'
import { Save, Trash2 } from 'lucide-react'
import type { JournalEntry } from '../lib/types'
import { useCreateJournal, useDeleteJournal, useUpdateJournal } from '../lib/queries'

const MOODS = ['great', 'good', 'ok', 'bad'] as const

export function JournalEditor({ date, entry, onSaved }: { date: string; entry: JournalEntry | null; onSaved?: () => void }) {
  const [title, setTitle] = useState(entry?.title ?? '')
  const [content, setContent] = useState(entry?.content ?? '')
  const [mood, setMood] = useState<string | null>(entry?.mood ?? null)
  const [preview, setPreview] = useState(false)

  const createJournal = useCreateJournal()
  const updateJournal = useUpdateJournal()
  const deleteJournal = useDeleteJournal()

  async function handleSave() {
    if (entry) {
      await updateJournal.mutateAsync({ id: entry.id, patch: { title, content, mood } })
    } else {
      await createJournal.mutateAsync({ date, title, content, mood })
      setTitle('')
      setContent('')
      setMood(null)
    }
    onSaved?.()
  }

  return (
    <div className="flex flex-col gap-3 rounded-xl border border-surface0 bg-mantle p-4">
      <div className="flex items-center gap-2">
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Entry title (optional)"
          className="flex-1 rounded-lg border border-surface0 bg-base px-3 py-1.5 text-sm text-text outline-none focus:border-accent"
        />
        <div className="flex gap-1">
          {MOODS.map((m) => (
            <button
              key={m}
              type="button"
              onClick={() => setMood(mood === m ? null : m)}
              className={`rounded-lg border px-2 py-1 text-xs capitalize ${
                mood === m ? 'border-accent bg-accent/15 text-accent' : 'border-surface1 text-subtext0'
              }`}
            >
              {m}
            </button>
          ))}
        </div>
      </div>

      <div className="flex items-center justify-between">
        <button type="button" onClick={() => setPreview((p) => !p)} className="text-xs text-subtext0 hover:text-text">
          {preview ? 'Edit' : 'Preview'}
        </button>
      </div>

      {preview ? (
        <div className="prose prose-invert min-h-32 max-w-none rounded-lg border border-surface0 bg-base p-3 text-sm text-text">
          <ReactMarkdown>{content || '*Nothing written yet*'}</ReactMarkdown>
        </div>
      ) : (
        <textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="Write in markdown..."
          rows={8}
          className="rounded-lg border border-surface0 bg-base px-3 py-2 font-mono text-sm text-text outline-none focus:border-accent"
        />
      )}

      <div className="flex justify-end gap-2">
        {entry && (
          <button
            type="button"
            onClick={() => deleteJournal.mutate(entry.id)}
            className="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm text-red hover:bg-red/10"
          >
            <Trash2 size={14} /> Delete
          </button>
        )}
        <button
          type="button"
          onClick={handleSave}
          className="flex items-center gap-1.5 rounded-lg bg-accent px-3 py-1.5 text-sm font-medium text-base hover:opacity-90"
        >
          <Save size={14} /> Save
        </button>
      </div>
    </div>
  )
}
