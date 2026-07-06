import { CheckCircle2, Clock, Flame, Target } from 'lucide-react'
import ReactMarkdown from 'react-markdown'
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
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
import { useGenerateSummary, useMonth, useStats, useSummary } from '../lib/queries'
import { fmtDuration } from '../lib/time'

const CHART_COLORS = ['var(--ctp-mauve)', 'var(--ctp-blue)', 'var(--ctp-green)', 'var(--ctp-peach)', 'var(--ctp-teal)', 'var(--ctp-pink)']

export function Dashboard() {
  const now = new Date()
  const monthKey = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  const from = `${monthKey}-01`
  const to = new Date(now.getFullYear(), now.getMonth() + 1, 0).toISOString().slice(0, 10)

  const { data: stats } = useStats(from, to)
  const { data: days = [] } = useMonth(now.getFullYear(), now.getMonth() + 1)
  const { data: summaryData } = useSummary('month', monthKey)
  const generateSummary = useGenerateSummary()

  const completionTrend = days
    .filter((d) => d.total > 0)
    .map((d) => ({ date: d.date.slice(5), ratio: Math.round(d.ratio * 100) }))

  const estVsActual = (stats?.est_vs_actual ?? []).map((d) => ({
    date: d.date.slice(5),
    Estimated: d.estimated_minutes,
    Actual: Math.round(d.actual_seconds / 60),
  }))

  const categoryTime = (stats?.time_by_category ?? []).filter((c) => c.seconds > 0)

  return (
    <div className="flex flex-col gap-6 p-6">
      <h1 className="text-xl font-semibold text-text">Dashboard</h1>

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
          value={stats ? `${stats.estimated_minutes_total}m / ${Math.round(stats.actual_seconds_total / 60)}m` : '—'}
        />
        <StatCard icon={Clock} label="Total focus time" value={stats ? fmtDuration(stats.actual_seconds_total) : '—'} />
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <ChartCard title="Est vs actual per day">
          <ResponsiveContainer width="100%" height={240}>
            <BarChart data={estVsActual}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--ctp-surface0)" />
              <XAxis dataKey="date" stroke="var(--ctp-subtext0)" fontSize={12} />
              <YAxis stroke="var(--ctp-subtext0)" fontSize={12} />
              <Tooltip contentStyle={{ background: 'var(--ctp-mantle)', border: '1px solid var(--ctp-surface0)', color: 'var(--ctp-text)' }} />
              <Bar dataKey="Estimated" fill="var(--ctp-overlay0)" radius={[4, 4, 0, 0]} />
              <Bar dataKey="Actual" fill="var(--ctp-accent)" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>

        <ChartCard title="Time by category">
          {categoryTime.length === 0 ? (
            <p className="flex h-60 items-center justify-center text-sm text-subtext0">No tracked time yet</p>
          ) : (
            <ResponsiveContainer width="100%" height={240}>
              <PieChart>
                <Pie data={categoryTime} dataKey="seconds" nameKey="category_name" innerRadius={60} outerRadius={90} paddingAngle={2}>
                  {categoryTime.map((_, i) => (
                    <Cell key={i} fill={CHART_COLORS[i % CHART_COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip contentStyle={{ background: 'var(--ctp-mantle)', border: '1px solid var(--ctp-surface0)', color: 'var(--ctp-text)' }} />
              </PieChart>
            </ResponsiveContainer>
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
                <p className="text-sm text-subtext0">No summary generated yet for this month.</p>
              )}
              <button
                type="button"
                onClick={() => generateSummary.mutate({ period: 'month', key: monthKey })}
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
