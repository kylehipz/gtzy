import { useState } from 'react'
import { AlertTriangle, CheckCircle2, Clock, Flame, Target } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Legend,
  Line,
  LineChart,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { StatCard } from '../components/StatCard'
import { useCategories, useGenerateSummary, useStats, useSummary, useTasks } from '../lib/queries'
import { fmtDuration, isoWeekKey, periodRangeLocal, todayLocal } from '../lib/time'

const CHART_COLORS = ['var(--ctp-mauve)', 'var(--ctp-blue)', 'var(--ctp-green)', 'var(--ctp-peach)', 'var(--ctp-teal)', 'var(--ctp-pink)']
const TOOLTIP_STYLE = { background: 'var(--ctp-mantle)', border: '1px solid var(--ctp-surface0)', color: 'var(--ctp-text)' }
const SUMMARY_PERIODS = [
  { value: 'day', label: 'Day' },
  { value: 'week', label: 'Week' },
  { value: 'month', label: 'Month' },
] as const

export function Dashboard() {
  const now = new Date()
  const monthKey = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  const [categoryFilter, setCategoryFilter] = useState<string>('')
  const [summaryPeriod, setSummaryPeriod] = useState<(typeof SUMMARY_PERIODS)[number]['value']>('week')

  const { from: periodFrom, to: periodTo } = periodRangeLocal(summaryPeriod, now)

  const { data: categories = [] } = useCategories()
  const { data: stats } = useStats(periodFrom, periodTo, categoryFilter || undefined)
  const summaryKey = summaryPeriod === 'day' ? todayLocal() : summaryPeriod === 'week' ? isoWeekKey(now) : monthKey
  const { data: summaryData } = useSummary(summaryPeriod, summaryKey)
  const { data: todayTasks = [] } = useTasks({ date: todayLocal() })
  const generateSummary = useGenerateSummary()

  const categoriesById = new Map(categories.map((c) => [c.id, c]))

  const statsDays = (stats?.est_vs_actual ?? []).filter((d) => d.total > 0)

  const completionTrend = statsDays.map((d) => ({ date: d.date.slice(5), ratio: Math.round((d.done / d.total) * 100) }))

  const estVsActual = (stats?.est_vs_actual ?? []).map((d) => ({
    date: d.date.slice(5),
    Estimated: Math.round(d.estimated_seconds / 60),
    Actual: Math.round(d.actual_seconds / 60),
  }))

  const estVsActualToday = todayTasks.map((t) => ({
    name: t.title,
    Estimated: Math.round(t.estimated_seconds / 60),
    Actual: Math.round(t.elapsed_seconds / 60),
    EstSec: t.estimated_seconds,
    ActualSec: t.elapsed_seconds,
  }))

  const categoryTime = (stats?.time_by_category ?? []).filter((c) => c.seconds > 0)
  const maxCategorySeconds = Math.max(1, ...categoryTime.map((c) => c.seconds))

  const todayMax = Math.max(1, ...estVsActualToday.map((t) => Math.max(t.Estimated, t.Actual)))

  const taskVolume = statsDays.map((d) => ({ date: d.date.slice(5), Completed: d.done, Incomplete: d.total - d.done }))

  return (
    <div className="flex flex-col gap-6 p-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold text-text">Dashboard</h1>
        <div className="flex gap-1">
          {SUMMARY_PERIODS.map((p) => (
            <button
              key={p.value}
              type="button"
              onClick={() => setSummaryPeriod(p.value)}
              className={`rounded-lg border px-2.5 py-1 text-xs transition-colors ${
                summaryPeriod === p.value ? 'border-accent bg-accent/15 text-accent' : 'border-surface0 text-subtext0 hover:border-surface1'
              }`}
            >
              {p.label}
            </button>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
        <StatCard
          icon={CheckCircle2}
          label="Completion rate"
          value={stats ? `${Math.round(stats.completion_rate * 100)}%` : '—'}
          sub={stats ? `${stats.tasks_completed}/${stats.tasks_total} tasks` : undefined}
        />
        <StatCard icon={Flame} label="Streak" value={stats ? `${stats.current_streak}d` : '—'} />
        <StatCard
          icon={Target}
          label="Est. vs actual"
          value={stats ? `${fmtDuration(stats.estimated_seconds_total)} / ${fmtDuration(stats.actual_seconds_total)}` : '—'}
        />
        <StatCard icon={Clock} label="Total focus time" value={stats ? fmtDuration(stats.actual_seconds_total) : '—'} />
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <ChartCard title="Est vs actual per day">
          <div className="mb-3 flex flex-wrap gap-1.5">
            <button
              type="button"
              onClick={() => setCategoryFilter('')}
              className={`rounded-lg border px-2.5 py-1 text-xs transition-colors ${
                categoryFilter === '' ? 'border-accent bg-accent/15 text-accent' : 'border-surface0 text-subtext0 hover:border-surface1'
              }`}
            >
              All categories
            </button>
            {categories.map((c) => (
              <button
                key={c.id}
                type="button"
                onClick={() => setCategoryFilter(String(c.id))}
                className={`flex items-center gap-1.5 rounded-lg border px-2.5 py-1 text-xs transition-colors ${
                  categoryFilter === String(c.id)
                    ? 'border-accent bg-accent/15 text-accent'
                    : 'border-surface0 text-subtext0 hover:border-surface1'
                }`}
              >
                <span className="h-2 w-2 rounded-full" style={{ backgroundColor: `var(--ctp-${c.color})` }} />
                {c.name}
              </button>
            ))}
          </div>
          <ResponsiveContainer width="100%" height={240}>
            <BarChart data={estVsActual}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--ctp-surface0)" />
              <XAxis dataKey="date" stroke="var(--ctp-subtext0)" fontSize={12} />
              <YAxis stroke="var(--ctp-subtext0)" fontSize={12} />
              <Tooltip contentStyle={TOOLTIP_STYLE} />
              <Legend wrapperStyle={{ fontSize: 12, color: 'var(--ctp-subtext0)' }} />
              <Bar dataKey="Estimated" fill="var(--ctp-overlay0)" radius={[4, 4, 0, 0]} maxBarSize={32} />
              <Bar dataKey="Actual" fill="var(--ctp-accent)" radius={[4, 4, 0, 0]} maxBarSize={32} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>

        <ChartCard title="Est vs actual today">
          {estVsActualToday.length === 0 ? (
            <p className="flex h-60 items-center justify-center text-sm text-subtext0">No tasks scheduled today</p>
          ) : (
            <div className="flex max-h-60 flex-col gap-2.5 overflow-y-auto">
              {estVsActualToday.map((t) => {
                const over = t.Estimated > 0 && t.Actual > t.Estimated
                const actualPct = (t.Actual / todayMax) * 100
                const estPct = (t.Estimated / todayMax) * 100
                return (
                  <div key={t.name} className="flex flex-col gap-1 text-xs">
                    <span className="truncate text-text">{t.name}</span>
                    <div className="flex items-center gap-2">
                      <div className="relative h-2 flex-1 rounded-full bg-surface0">
                        <div
                          className="absolute inset-y-0 left-0 rounded-full"
                          style={{ width: `${actualPct}%`, backgroundColor: over ? 'var(--ctp-red)' : 'var(--ctp-accent)' }}
                        />
                        {t.Estimated > 0 && (
                          <span
                            className="absolute inset-y-0 w-px bg-subtext0"
                            style={{ left: `${Math.min(estPct, 100)}%` }}
                          />
                        )}
                      </div>
                      <span className={`flex shrink-0 items-center gap-1 font-mono ${over ? 'text-red' : 'text-subtext0'}`}>
                        {fmtDuration(t.ActualSec)} / {fmtDuration(t.EstSec)}
                        {over && <AlertTriangle size={11} />}
                      </span>
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </ChartCard>

        <ChartCard title="Time by category">
          {categoryTime.length === 0 ? (
            <p className="flex h-60 items-center justify-center text-sm text-subtext0">No tracked time yet</p>
          ) : (
            <div className="flex flex-col items-center gap-4">
              <ResponsiveContainer width="100%" height={200} className="max-w-[280px]">
                <PieChart>
                  <Pie data={categoryTime} dataKey="seconds" nameKey="category_name" innerRadius={60} outerRadius={90} paddingAngle={2}>
                    {categoryTime.map((c, i) => {
                      const cat = c.category_id ? categoriesById.get(c.category_id) : null
                      return <Cell key={i} fill={cat ? `var(--ctp-${cat.color})` : CHART_COLORS[i % CHART_COLORS.length]} />
                    })}
                  </Pie>
                  <Tooltip contentStyle={TOOLTIP_STYLE} />
                </PieChart>
              </ResponsiveContainer>
              <div className="flex w-full flex-col gap-2">
                {categoryTime.map((c, i) => {
                  const cat = c.category_id ? categoriesById.get(c.category_id) : null
                  const color = cat ? `var(--ctp-${cat.color})` : CHART_COLORS[i % CHART_COLORS.length]
                  return (
                    <div key={c.category_id ?? 'none'} className="flex items-center gap-2 text-xs">
                      <span className="h-2.5 w-2.5 shrink-0 rounded-full" style={{ backgroundColor: color }} />
                      <span className="w-20 shrink-0 truncate text-text">{c.category_name}</span>
                      <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-surface0">
                        <div
                          className="h-full rounded-full"
                          style={{ width: `${(c.seconds / maxCategorySeconds) * 100}%`, backgroundColor: color }}
                        />
                      </div>
                      <span className="w-14 shrink-0 text-right font-mono text-subtext0">{fmtDuration(c.seconds)}</span>
                    </div>
                  )
                })}
              </div>
            </div>
          )}
        </ChartCard>

        <ChartCard title="Completion trend">
          <ResponsiveContainer width="100%" height={240}>
            <LineChart data={completionTrend}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--ctp-surface0)" />
              <XAxis dataKey="date" stroke="var(--ctp-subtext0)" fontSize={12} />
              <YAxis stroke="var(--ctp-subtext0)" fontSize={12} domain={[0, 100]} />
              <Tooltip contentStyle={{ background: 'var(--ctp-mantle)', border: '1px solid var(--ctp-surface0)', color: 'var(--ctp-text)' }} />
              <Line type="monotone" dataKey="ratio" stroke="var(--ctp-green)" strokeWidth={2} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </ChartCard>

        <ChartCard title="Tasks completed per day">
          <ResponsiveContainer width="100%" height={240}>
            <BarChart data={taskVolume}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--ctp-surface0)" />
              <XAxis dataKey="date" stroke="var(--ctp-subtext0)" fontSize={12} />
              <YAxis stroke="var(--ctp-subtext0)" fontSize={12} allowDecimals={false} />
              <Tooltip contentStyle={TOOLTIP_STYLE} />
              <Legend wrapperStyle={{ fontSize: 12, color: 'var(--ctp-subtext0)' }} />
              <Bar dataKey="Completed" stackId="a" fill="var(--ctp-green)" maxBarSize={32} />
              <Bar dataKey="Incomplete" stackId="a" fill="var(--ctp-surface1)" radius={[4, 4, 0, 0]} maxBarSize={32} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>

        <ChartCard title="AI Growth Summary">
          {!summaryData?.enabled ? (
            <div className="flex h-60 flex-col items-center justify-center gap-2 text-center text-sm text-subtext0">
              <p>Add ANTHROPIC_API_KEY to enable AI-generated growth summaries.</p>
            </div>
          ) : (
            <div className="flex h-60 flex-col gap-2 overflow-y-auto">
              {summaryData.summary ? (
                <div className="prose prose-invert max-w-none text-sm text-text">
                  <ReactMarkdown>{summaryData.summary.content}</ReactMarkdown>
                </div>
              ) : (
                <p className="text-sm text-subtext0">No summary generated yet for this period.</p>
              )}
              <button
                type="button"
                onClick={() => generateSummary.mutate({ period: summaryPeriod, key: summaryKey })}
                disabled={generateSummary.isPending}
                className="mt-auto self-start rounded-lg bg-accent px-3 py-1.5 text-sm font-medium text-base hover:opacity-90 disabled:opacity-50"
              >
                {generateSummary.isPending ? 'Generating…' : summaryData.summary ? 'Regenerate' : 'Generate summary'}
              </button>
            </div>
          )}
        </ChartCard>
      </div>
    </div>
  )
}

function ChartCard({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-xl border border-surface0 bg-mantle p-4">
      <p className="mb-2 text-sm font-medium text-text">{title}</p>
      {children}
    </div>
  )
}
