import { useState } from 'react'
import { Droplet, RefreshCw } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import {
  CartesianGrid,
  Line,
  LineChart,
  ReferenceArea,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { StatCard } from '../components/StatCard'
import { EmptyState } from '../components/EmptyState'
import {
  useBloodSugar,
  useCreateReading,
  useDeleteReading,
  useGenerateBloodSugarSummary,
  useSummary,
  useSyncMeter,
} from '../lib/queries'
import { addDaysLocal, todayLocal } from '../lib/time'
import type { BloodSugarReading, MealTag } from '../lib/types'

const MEAL_TAGS: { value: MealTag; label: string }[] = [
  { value: '', label: 'No tag' },
  { value: 'fasting', label: 'Fasting' },
  { value: 'pre_meal', label: 'Pre-meal' },
  { value: 'post_meal', label: 'Post-meal' },
  { value: 'bedtime', label: 'Bedtime' },
  { value: 'random', label: 'Random' },
]
const MEAL_LABELS: Record<string, string> = Object.fromEntries(MEAL_TAGS.map((t) => [t.value, t.label]))
const TOOLTIP_STYLE = { background: 'var(--ctp-mantle)', border: '1px solid var(--ctp-surface0)', color: 'var(--ctp-text)' }

// Target range for time-in-range and the chart band (mg/dL).
const LOW = 70
const HIGH = 180

function localDateTimeNow(): string {
  const d = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function valueColor(v: number): string {
  if (v < LOW) return 'var(--ctp-red)'
  if (v > HIGH) return 'var(--ctp-peach)'
  return 'var(--ctp-green)'
}

export function BloodSugar() {
  const [from, setFrom] = useState(addDaysLocal(todayLocal(), -13))
  const [to, setTo] = useState(todayLocal())
  const { data: readings = [] } = useBloodSugar(from, to + 'T23:59:59Z')

  const [value, setValue] = useState('')
  const [takenAt, setTakenAt] = useState(localDateTimeNow())
  const [mealTag, setMealTag] = useState<MealTag>('')
  const [notes, setNotes] = useState('')

  const createReading = useCreateReading()
  const deleteReading = useDeleteReading()
  const syncMeter = useSyncMeter()

  const summaryKey = `${from}..${to}`
  const { data: summaryData } = useSummary('blood_sugar', summaryKey)
  const generateSummary = useGenerateBloodSugarSummary()

  const stats = computeStats(readings)
  const chartData = [...readings]
    .sort((a, b) => a.taken_at.localeCompare(b.taken_at))
    .map((r) => ({
      label: new Date(r.taken_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric' }),
      value: r.value_mgdl,
    }))

  function submitReading(e: React.FormEvent) {
    e.preventDefault()
    const v = Number(value)
    if (!Number.isFinite(v) || v <= 0) return
    createReading.mutate(
      { value_mgdl: v, taken_at: new Date(takenAt).toISOString(), meal_tag: mealTag, notes },
      {
        onSuccess: () => {
          setValue('')
          setNotes('')
          setTakenAt(localDateTimeNow())
        },
      },
    )
  }

  return (
    <div className="flex flex-col gap-4 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold text-text">Blood Sugar</h1>
        <button
          type="button"
          onClick={() => syncMeter.mutate()}
          disabled={syncMeter.isPending}
          className="flex items-center gap-2 rounded-lg border border-surface0 bg-mantle px-3 py-1.5 text-sm font-medium text-text hover:border-surface1 disabled:opacity-50"
        >
          <RefreshCw size={15} className={syncMeter.isPending ? 'animate-spin' : ''} />
          {syncMeter.isPending ? 'Syncing…' : 'Sync meter'}
        </button>
      </div>

      {syncMeter.isError && (
        <p className="rounded-lg border border-red/40 bg-red/10 px-3 py-2 text-sm text-red">
          {(syncMeter.error as Error).message}
        </p>
      )}
      {syncMeter.isSuccess && (
        <p className="rounded-lg border border-green/40 bg-green/10 px-3 py-2 text-sm text-green">
          Synced {syncMeter.data.synced} new reading(s) ({syncMeter.data.fetched} fetched from meter).
        </p>
      )}

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <StatCard icon={Droplet} label="Readings" value={String(stats.count)} sub={`${from} → ${to}`} />
        <StatCard icon={Droplet} label="Average" value={stats.count ? `${stats.mean.toFixed(0)}` : '—'} sub="mg/dL" />
        <StatCard icon={Droplet} label="Est. A1C" value={stats.count ? `${stats.estA1c.toFixed(1)}%` : '—'} sub="ADAG estimate" />
        <StatCard icon={Droplet} label="In range" value={stats.count ? `${stats.inRangePct.toFixed(0)}%` : '—'} sub={`${LOW}–${HIGH} mg/dL`} />
      </div>

      <div className="flex flex-col gap-3 rounded-xl border border-surface0 bg-mantle p-4">
        <p className="text-sm font-medium text-text">Readings over time</p>
        {chartData.length === 0 ? (
          <p className="text-sm text-subtext0">No readings in this range yet.</p>
        ) : (
          <ResponsiveContainer width="100%" height={240}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--ctp-surface0)" />
              <XAxis dataKey="label" stroke="var(--ctp-subtext0)" fontSize={12} />
              <YAxis stroke="var(--ctp-subtext0)" fontSize={12} domain={[0, 'auto']} />
              <ReferenceArea y1={LOW} y2={HIGH} fill="var(--ctp-green)" fillOpacity={0.12} />
              <Tooltip contentStyle={TOOLTIP_STYLE} formatter={(v) => [`${v} mg/dL`, 'Glucose']} />
              <Line type="monotone" dataKey="value" stroke="var(--ctp-mauve)" strokeWidth={2} dot={{ r: 3 }} />
            </LineChart>
          </ResponsiveContainer>
        )}
      </div>

      <form onSubmit={submitReading} className="flex flex-wrap items-end gap-3 rounded-xl border border-surface0 bg-mantle p-4">
        <label className="flex flex-col gap-1 text-xs text-subtext0">
          Value (mg/dL)
          <input
            type="number"
            min={1}
            step={1}
            value={value}
            onChange={(e) => setValue(e.target.value)}
            required
            className="w-28 rounded-lg border border-surface0 bg-base px-2 py-1.5 text-sm text-text"
          />
        </label>
        <label className="flex flex-col gap-1 text-xs text-subtext0">
          When
          <input
            type="datetime-local"
            value={takenAt}
            onChange={(e) => setTakenAt(e.target.value)}
            className="rounded-lg border border-surface0 bg-base px-2 py-1.5 text-sm text-text"
          />
        </label>
        <label className="flex flex-col gap-1 text-xs text-subtext0">
          Tag
          <select
            value={mealTag}
            onChange={(e) => setMealTag(e.target.value as MealTag)}
            className="rounded-lg border border-surface0 bg-base px-2 py-1.5 text-sm text-text"
          >
            {MEAL_TAGS.map((t) => (
              <option key={t.value || 'none'} value={t.value}>
                {t.label}
              </option>
            ))}
          </select>
        </label>
        <label className="flex min-w-40 flex-1 flex-col gap-1 text-xs text-subtext0">
          Notes
          <input
            type="text"
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            className="rounded-lg border border-surface0 bg-base px-2 py-1.5 text-sm text-text"
          />
        </label>
        <button
          type="submit"
          disabled={createReading.isPending}
          className="rounded-lg bg-accent px-3 py-1.5 text-sm font-medium text-base hover:opacity-90 disabled:opacity-50"
        >
          Add reading
        </button>
      </form>

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
          <p className="text-sm text-subtext0">Add ANTHROPIC_API_KEY to enable AI-generated blood-sugar summaries.</p>
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
              onClick={() => generateSummary.mutate({ from, to })}
              disabled={generateSummary.isPending}
              className="self-start rounded-lg bg-accent px-3 py-1.5 text-sm font-medium text-base hover:opacity-90 disabled:opacity-50"
            >
              {generateSummary.isPending ? 'Generating…' : summaryData.summary ? 'Regenerate' : 'Generate summary'}
            </button>
            <p className="text-xs text-subtext0">Informational only — not a substitute for advice from your care team.</p>
          </div>
        )}
      </div>

      <div>
        <p className="mb-2 text-xs font-medium uppercase tracking-wide text-subtext0">All readings</p>
        {readings.length === 0 ? (
          <EmptyState icon={Droplet} title="No readings yet" description="Add one above or sync your meter." />
        ) : (
          <div className="flex flex-col gap-2">
            {readings.map((r) => (
              <div key={r.id} className="flex items-center gap-3 rounded-lg border border-surface0 bg-mantle p-3 hover:border-surface1">
                <span className="w-16 text-lg font-semibold" style={{ color: valueColor(r.value_mgdl) }}>
                  {r.value_mgdl.toFixed(0)}
                </span>
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-text">{new Date(r.taken_at).toLocaleString()}</span>
                    {r.meal_tag && (
                      <span className="rounded-full bg-surface0 px-2 py-0.5 text-xs text-subtext0">{MEAL_LABELS[r.meal_tag] ?? r.meal_tag}</span>
                    )}
                    {r.source === 'meter' && (
                      <span className="rounded-full bg-blue/15 px-2 py-0.5 text-xs text-blue">meter</span>
                    )}
                  </div>
                  {r.notes && <p className="mt-1 truncate text-sm text-subtext0">{r.notes}</p>}
                </div>
                <button
                  type="button"
                  onClick={() => deleteReading.mutate(r.id)}
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

interface BSStats {
  count: number
  mean: number
  estA1c: number
  inRangePct: number
}

function computeStats(readings: BloodSugarReading[]): BSStats {
  if (readings.length === 0) return { count: 0, mean: 0, estA1c: 0, inRangePct: 0 }
  const sum = readings.reduce((acc, r) => acc + r.value_mgdl, 0)
  const inRange = readings.filter((r) => r.value_mgdl >= LOW && r.value_mgdl <= HIGH).length
  const mean = sum / readings.length
  return {
    count: readings.length,
    mean,
    estA1c: (mean + 46.7) / 28.7,
    inRangePct: (inRange / readings.length) * 100,
  }
}
