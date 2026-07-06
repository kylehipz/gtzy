import type { LucideIcon } from 'lucide-react'

export function EmptyState({
  icon: Icon,
  title,
  description,
}: {
  icon: LucideIcon
  title: string
  description?: string
}) {
  return (
    <div className="flex flex-col items-center justify-center gap-2 rounded-xl border border-dashed border-surface1 py-16 text-center text-subtext0">
      <Icon size={32} className="text-overlay0" />
      <p className="font-medium text-text">{title}</p>
      {description && <p className="max-w-sm text-sm">{description}</p>}
    </div>
  )
}
