import { X } from 'lucide-react'
import { motion } from 'framer-motion'

export function Modal({
  title,
  onClose,
  children,
}: {
  title: string
  onClose: () => void
  children: React.ReactNode
}) {
  return (
    <div
      className="fixed inset-0 z-20 flex items-center justify-center bg-crust/60 p-4"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose()
      }}
    >
      <motion.div
        initial={{ opacity: 0, scale: 0.96 }}
        animate={{ opacity: 1, scale: 1 }}
        className="flex max-h-[90vh] w-full max-w-lg flex-col gap-4 overflow-y-auto rounded-2xl border border-surface0 bg-mantle p-6"
      >
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-text">{title}</h2>
          <button type="button" onClick={onClose} className="text-subtext0 hover:text-text">
            <X size={18} />
          </button>
        </div>
        {children}
      </motion.div>
    </div>
  )
}
