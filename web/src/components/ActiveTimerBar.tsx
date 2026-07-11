import { useEffect, useState } from 'react'
import { Pause, Check, Timer } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { useCompleteTask, useCurrentTimer, usePauseCurrent } from '../lib/queries'
import { elapsedSecondsLive, fmtDuration } from '../lib/time'

export function ActiveTimerBar() {
  const { data } = useCurrentTimer()
  const pauseCurrent = usePauseCurrent()
  const completeTask = useCompleteTask()
  const [, forceTick] = useState(0)

  const current = data?.current ?? null

  useEffect(() => {
    if (!current) return
    const id = setInterval(() => forceTick((n) => n + 1), 1000)
    return () => clearInterval(id)
  }, [current])

  return (
    <AnimatePresence>
      {current && (
        <motion.div
          initial={{ y: -40, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          exit={{ y: -40, opacity: 0 }}
          className="sticky top-0 z-10 flex items-center justify-between gap-4 border-b border-surface0 bg-mantle/95 px-6 py-3 backdrop-blur"
        >
          <div className="flex items-center gap-3 overflow-hidden">
            <span className="relative flex h-2.5 w-2.5 shrink-0">
              <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-green opacity-75" />
              <span className="relative inline-flex h-2.5 w-2.5 rounded-full bg-green" />
            </span>
            <Timer size={16} className="shrink-0 text-subtext0" />
            <span className="truncate font-medium text-text">{current.title}</span>
            <span className="shrink-0 font-mono text-sm text-subtext0">
              {fmtDuration(elapsedSecondsLive(current.active_started_at, current.actual_seconds))}
              {current.estimated_seconds > 0 && ` / ${fmtDuration(current.estimated_seconds)}`}
            </span>
          </div>
          <div className="flex shrink-0 gap-2">
            <button
              type="button"
              onClick={() => pauseCurrent.mutate()}
              className="flex items-center gap-1.5 rounded-lg border border-surface1 px-3 py-1.5 text-sm text-text hover:bg-surface0"
            >
              <Pause size={14} /> Pause
            </button>
            <button
              type="button"
              onClick={() => completeTask.mutate(current.id)}
              className="flex items-center gap-1.5 rounded-lg bg-accent px-3 py-1.5 text-sm text-base hover:opacity-90"
            >
              <Check size={14} /> Complete
            </button>
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  )
}
