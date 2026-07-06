export function ColorSwatch({
  color,
  selected,
  onClick,
}: {
  color: string
  selected: boolean
  onClick: () => void
}) {
  return (
    <button
      type="button"
      title={color}
      onClick={onClick}
      className={`h-5 w-5 rounded-full border-2 transition-transform ${
        selected ? 'scale-110 border-text' : 'border-transparent hover:scale-105'
      }`}
      style={{ backgroundColor: `var(--ctp-${color})` }}
    />
  )
}
