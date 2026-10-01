/**
 * Thin wrapper over the Traccia tag that Layout.astro loads in production.
 *
 * The tag defines `window.traccia` once it has executed. In dev builds, when
 * PUBLIC_TRACCIA_PROJECT_ID is unset, or on pages that opt out of analytics,
 * it is never defined — so every call here is a no-op instead of a crash.
 */

type Metadata = Record<string, string | number | boolean>

interface TracciaClient {
  track(name: string, metadata?: Metadata): void
}

declare global {
  interface Window {
    traccia?: TracciaClient
  }
}

export function track(name: string, metadata?: Metadata): void {
  if (typeof window === 'undefined') return
  window.traccia?.track(name, metadata)
}
