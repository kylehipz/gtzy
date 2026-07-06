import type { Category } from '../lib/types'

export function CategoryBadge({ category }: { category: Category | null | undefined }) {
  if (!category) return null
  return (
    <span
      className="rounded-full px-2 py-0.5 text-xs font-medium"
      style={{ backgroundColor: `color-mix(in oklab, var(--ctp-${category.color}) 20%, transparent)`, color: `var(--ctp-${category.color})` }}
    >
      {category.name}
    </span>
  )
}
