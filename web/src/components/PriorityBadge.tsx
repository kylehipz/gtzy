import type { Priority } from '../lib/types'

const STYLES: Record<Priority, string> = {
  urgent: 'bg-red/20 text-red',
  high: 'bg-peach/20 text-peach',
  medium: 'bg-yellow/20 text-yellow',
  low: 'bg-teal/20 text-teal',
}

export function PriorityBadge({ priority }: { priority: Priority }) {
  return (
    <span className={`rounded-full px-2 py-0.5 text-xs font-medium capitalize ${STYLES[priority]}`}>
      {priority}
    </span>
  )
}
