import type { LucideIcon } from 'lucide-react'

export function StatCard({
  icon: Icon,
  label,
  value,
  sub,
}: {
  icon: LucideIcon
  label: string
  value: string
  sub?: string
}) {
  return (
    <div className="flex flex-col gap-1 rounded-xl border border-surface0 bg-mantle p-4">
      <div className="flex items-center gap-2 text-subtext0">
        <Icon size={16} />
        <span className="text-xs font-medium uppercase tracking-wide">{label}</span>
      </div>
      <span className="text-2xl font-semibold text-text">{value}</span>
      {sub && <span className="text-xs text-subtext0">{sub}</span>}
    </div>
  )
}
