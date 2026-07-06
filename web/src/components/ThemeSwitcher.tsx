import { ACCENTS, FLAVORS, useTheme, type Flavor } from '../theme/ThemeProvider'
import { ColorSwatch } from './ColorSwatch'

const FLAVOR_LABELS: Record<Flavor, string> = {
  latte: 'Latte',
  frappe: 'Frappé',
  macchiato: 'Macchiato',
  mocha: 'Mocha',
}

export function ThemeSwitcher() {
  const { flavor, accent, setFlavor, setAccent } = useTheme()

  return (
    <div className="flex flex-col gap-4">
      <div>
        <p className="mb-2 text-xs font-medium uppercase tracking-wide text-subtext0">Flavor</p>
        <div className="flex flex-wrap gap-2">
          {FLAVORS.map((f) => (
            <button
              key={f}
              type="button"
              onClick={() => setFlavor(f)}
              className={`rounded-lg border px-3 py-1.5 text-sm transition-colors ${
                flavor === f
                  ? 'border-accent bg-accent/15 text-accent'
                  : 'border-surface0 text-subtext0 hover:border-surface1'
              }`}
            >
              {FLAVOR_LABELS[f]}
            </button>
          ))}
        </div>
      </div>
      <div>
        <p className="mb-2 text-xs font-medium uppercase tracking-wide text-subtext0">Accent</p>
        <div className="flex flex-wrap gap-2">
          {ACCENTS.map((a) => (
            <ColorSwatch key={a} color={a} selected={accent === a} onClick={() => setAccent(a)} />
          ))}
        </div>
      </div>
    </div>
  )
}
