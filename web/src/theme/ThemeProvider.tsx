import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'

export type Flavor = 'latte' | 'frappe' | 'macchiato' | 'mocha'
export const FLAVORS: Flavor[] = ['latte', 'frappe', 'macchiato', 'mocha']

export const ACCENTS = [
  'rosewater', 'flamingo', 'pink', 'mauve', 'red', 'maroon', 'peach',
  'yellow', 'green', 'teal', 'sky', 'sapphire', 'blue', 'lavender',
] as const
export type Accent = (typeof ACCENTS)[number]

interface ThemeContextValue {
  flavor: Flavor
  accent: Accent
  setFlavor: (f: Flavor) => void
  setAccent: (a: Accent) => void
}

const ThemeContext = createContext<ThemeContextValue | null>(null)

const FLAVOR_KEY = 'gtzy-flavor'
const ACCENT_KEY = 'gtzy-accent'

function readStored<T extends string>(key: string, allowed: readonly T[], fallback: T): T {
  const v = localStorage.getItem(key)
  return (allowed as readonly string[]).includes(v ?? '') ? (v as T) : fallback
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [flavor, setFlavorState] = useState<Flavor>(() => readStored(FLAVOR_KEY, FLAVORS, 'macchiato'))
  const [accent, setAccentState] = useState<Accent>(() => readStored(ACCENT_KEY, ACCENTS, 'mauve'))

  useEffect(() => {
    document.documentElement.setAttribute('data-flavor', flavor)
    localStorage.setItem(FLAVOR_KEY, flavor)
  }, [flavor])

  useEffect(() => {
    document.documentElement.style.setProperty('--ctp-accent', `var(--ctp-${accent})`)
    localStorage.setItem(ACCENT_KEY, accent)
  }, [accent])

  const value = useMemo(
    () => ({ flavor, accent, setFlavor: setFlavorState, setAccent: setAccentState }),
    [flavor, accent],
  )

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}

export function useTheme() {
  const ctx = useContext(ThemeContext)
  if (!ctx) throw new Error('useTheme must be used within a ThemeProvider')
  return ctx
}
