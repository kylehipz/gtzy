const SIZE_CLASSES = {
  sm: 'h-4 w-4 rounded-full',
  md: 'h-8 w-8 rounded-full',
} as const

export function ColorSwatch({
  color,
  selected,
  onClick,
  size = 'sm',
}: {
  color: string
  selected: boolean
  onClick: () => void
  size?: keyof typeof SIZE_CLASSES
}) {
  return (
    <button
      type="button"
      title={color}
      onClick={onClick}
      className={`${SIZE_CLASSES[size]} border-2 transition-transform ${
        selected ? 'scale-110 border-text' : 'border-transparent hover:scale-105'
      }`}
      style={{ backgroundColor: `var(--ctp-${color})` }}
    />
  )
}
