<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useTheme } from '../composables/useTheme'
import { Lock, Download, Copy, Check, AlertTriangle, Shield, Eye, EyeOff, Clock, FileText, Trash2, Terminal, X, Zap, UserX, Timer } from 'lucide-vue-next'

const { current } = useTheme()

type State = 'loading' | 'password' | 'decrypting' | 'success' | 'destroyed' | 'error'

const props = defineProps<{
  demo?: boolean
}>()

const state = ref<State>('loading')
const errorType = ref<'NOT_FOUND' | 'INVALID_KEY' | 'WRONG_PASSWORD' | 'EXPIRED' | 'SERVER_ERROR'>('NOT_FOUND')
const password = ref('')
const showPassword = ref(false)
const passwordError = ref(false)
const filename = ref('.env')
const readsLeft = ref(0)
const wasLastRead = ref(false)
const content = ref('')
const copied = ref(false)
const copiedInstall = ref(false)
const revealContent = ref(false)
const showInstallModal = ref(false)

const installMethods = [
  { label: 'Homebrew', command: 'brew install envshare' },
  { label: 'Go', command: 'go install github.com/antoniojosev/envshare@latest' },
  { label: 'curl', command: 'curl -fsSL https://envshare.dev/install.sh | sh' },
]
const activeInstall = ref(0)

function copyInstallCommand() {
  navigator.clipboard.writeText(installMethods[activeInstall.value].command).then(() => {
    copiedInstall.value = true
    setTimeout(() => (copiedInstall.value = false), 2000)
  })
}

// Parsed env vars for display
const envLines = computed(() => {
  if (!content.value) return []
  return content.value.split('\n').map((line) => {
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#')) {
      return { type: 'comment' as const, text: line }
    }
    const eqIdx = line.indexOf('=')
    if (eqIdx === -1) return { type: 'text' as const, text: line }
    return {
      type: 'var' as const,
      key: line.slice(0, eqIdx),
      value: line.slice(eqIdx + 1),
    }
  })
})

function maskValue(val: string): string {
  if (val.length <= 4) return '•'.repeat(val.length || 4)
  return val.slice(0, 2) + '•'.repeat(Math.min(val.length - 4, 20)) + val.slice(-2)
}

function copyContent() {
  navigator.clipboard.writeText(content.value).then(() => {
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
  })
}

function downloadFile() {
  const blob = new Blob([content.value], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename.value
  a.click()
  URL.revokeObjectURL(url)
}

async function submitPassword() {
  passwordError.value = false
  state.value = 'decrypting'

  if (props.demo) {
    await fakePause(800)
    state.value = 'success'
    return
  }

  // Real decrypt would go here
  try {
    const { decryptSecret } = await import('../lib/crypto')
    const hash = window.location.hash.slice(1)
    const secretData = (window as any).__SECRET_DATA__
    const result = await decryptSecret(hash, secretData, password.value)
    content.value = result
    state.value = 'success'
  } catch {
    passwordError.value = true
    state.value = 'password'
  }
}

function fakePause(ms: number) {
  return new Promise((r) => setTimeout(r, ms))
}

async function runDemo() {
  // Simulate the full flow
  state.value = 'loading'
  await fakePause(1200)

  // Simulate password prompt
  state.value = 'password'

  // Content will be set after password
  filename.value = '.env.production'
  readsLeft.value = 0
  wasLastRead.value = true
  content.value = `# Database
DATABASE_URL=postgresql://admin:s3cret@db.internal:5432/envshare_prod
DATABASE_POOL_SIZE=25

# Redis
REDIS_URL=redis://:r3d1s_p4ss@redis.internal:6379/0

# API Keys
STRIPE_SECRET_KEY=sk_live_51N8x•••••••••••••••kQ7
STRIPE_WEBHOOK_SECRET=whsec_•••••••••••••••

# Auth
JWT_SECRET=eyJhbGciOiJIUzI1NiJ9.••••••
SESSION_SECRET=a7f2e9c1d4b8••••••••

# AWS
AWS_ACCESS_KEY_ID=AKIA••••••••••••
AWS_SECRET_ACCESS_KEY=wJal••••••••••••••••••••
AWS_REGION=us-east-1
S3_BUCKET=envshare-uploads-prod

# Monitoring
SENTRY_DSN=https://abc123@sentry.io/456
LOG_LEVEL=warn

# App
PORT=8080
NODE_ENV=production
API_BASE_URL=https://api.envshare.dev`
}

async function runRealDecrypt() {
  try {
    const { parseSecretUrl, fetchSecret, decryptSecret } = await import('../lib/crypto')
    const parsed = parseSecretUrl()
    if (!parsed) {
      errorType.value = 'INVALID_KEY'
      state.value = 'error'
      return
    }

    const apiBase = import.meta.env.PUBLIC_API_URL || ''
    const secret = await fetchSecret(apiBase, parsed.id)

    // Store for password flow
    ;(window as any).__SECRET_DATA__ = secret
    filename.value = secret.filename || '.env'
    readsLeft.value = secret.reads_left
    wasLastRead.value = secret.reads_left === 0

    if (secret.password_protected) {
      state.value = 'password'
      return
    }

    state.value = 'decrypting'
    const result = await decryptSecret(parsed.key, secret)
    content.value = result
    state.value = 'success'
  } catch (e: any) {
    if (e?.message === 'NOT_FOUND') {
      errorType.value = 'NOT_FOUND'
    } else if (e?.message === 'INVALID_KEY') {
      errorType.value = 'INVALID_KEY'
    } else {
      errorType.value = 'SERVER_ERROR'
    }
    state.value = 'error'
  }
}

onMounted(() => {
  if (props.demo) {
    runDemo()
  } else {
    runRealDecrypt()
  }
})
</script>

<template>
  <div class="min-h-screen flex flex-col font-sans" :style="{ background: `var(--t-s1)`, color: `var(--t-t1)` }">
    <!-- Nav -->
    <nav class="px-4 sm:px-6 lg:px-12 flex items-center justify-between border-b border-b1 gap-3" style="height: 3.5rem; min-height: 3.5rem;">
      <a href="/" class="flex items-center gap-2 font-mono font-bold text-sm tracking-wide text-t1 shrink-0">
        <span class="rounded-full bg-hi pulse-glow" style="width: .5rem; height: .5rem;" />
        envshare
      </a>
      <div class="flex items-center gap-3 sm:gap-4 min-w-0">
        <div class="hidden md:flex items-center gap-1.5 text-xs text-t3 font-mono">
          <Shield :size="12" class="text-hi shrink-0" />
          Zero-knowledge · Decrypted in your browser
        </div>
        <!-- Header CTA: Install button -->
        <button
          @click="showInstallModal = true"
          class="flex items-center font-mono font-semibold text-s1 bg-hi rounded-lg transition-all hover:opacity-90 whitespace-nowrap cursor-pointer"
          style="gap: 0.375rem; padding: 0.375rem 0.75rem; font-size: 0.6875rem;"
        >
          <Terminal :size="12" />
          Install CLI
        </button>
      </div>
    </nav>

    <!-- Content -->
    <div class="flex-1 flex items-center justify-center px-4 sm:px-6 py-8 sm:py-12">
      <div style="width: 100%; max-width: 42rem;">

        <!-- ═══ LOADING STATE ═══ -->
        <div v-if="state === 'loading'" class="text-center">
          <div class="rounded-2xl bg-s2 border border-b1 flex items-center justify-center mx-auto mb-6" style="width: 4rem; height: 4rem;">
            <div class="border-2 border-hi/30 border-t-hi rounded-full animate-spin" style="width: 1.5rem; height: 1.5rem;" />
          </div>
          <h2 class="font-display font-bold text-2xl mb-2">Fetching encrypted secret...</h2>
          <p class="text-t3 text-sm">The decryption key never leaves your browser.</p>
        </div>

        <!-- ═══ PASSWORD STATE ═══ -->
        <div v-else-if="state === 'password'" style="max-width: 28rem; margin-left: auto; margin-right: auto;">
          <div class="rounded-2xl bg-hi/8 border border-hi/20 flex items-center justify-center mx-auto mb-5 sm:mb-6" style="width: 4rem; height: 4rem;">
            <Lock :size="28" class="text-hi" />
          </div>
          <h2 class="font-display font-bold text-xl sm:text-2xl text-center mb-2">Password required</h2>
          <p class="text-t3 text-sm text-center mb-6 sm:mb-8 leading-relaxed">This secret is password-protected. Enter the password shared with you.</p>

          <form @submit.prevent="submitPassword">
            <div class="relative">
              <input
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                placeholder="Enter password"
                autofocus
                class="w-full bg-s2 border rounded-xl font-mono text-sm placeholder-t3 focus:outline-none focus:border-hi/40 transition-colors"
                :class="passwordError ? 'border-red' : 'border-b1'"
                style="padding: 0.875rem 3rem 0.875rem 1rem;"
              />
              <button
                type="button"
                @click="showPassword = !showPassword"
                class="absolute text-t3 hover:text-t2 transition-colors"
                style="right: 0.875rem; top: 50%; transform: translateY(-50%); padding: 0.25rem;"
              >
                <EyeOff v-if="showPassword" :size="18" />
                <Eye v-else :size="18" />
              </button>
            </div>
            <p v-if="passwordError" class="text-red text-xs flex items-center gap-1.5" style="margin-top: 0.5rem;">
              <AlertTriangle :size="12" class="shrink-0" />
              Wrong password. Please try again.
            </p>
            <button
              type="submit"
              class="w-full bg-hi text-s1 font-semibold rounded-xl transition-all cursor-pointer"
              style="padding: 0.875rem; margin-top: 1rem; opacity: 0.9;"
              @mouseenter="($event.target as HTMLElement).style.opacity = '1'"
              @mouseleave="($event.target as HTMLElement).style.opacity = '0.9'"
            >
              Decrypt secret
            </button>
          </form>

          <div class="flex items-center justify-center gap-2 text-t3 font-mono text-center" style="margin-top: 1.25rem; font-size: 0.7rem;">
            <Shield :size="12" class="text-hi shrink-0" />
            Password is used locally — never sent to server
          </div>
        </div>

        <!-- ═══ DECRYPTING STATE ═══ -->
        <div v-else-if="state === 'decrypting'" class="text-center">
          <div class="rounded-2xl bg-hi/8 border border-hi/20 flex items-center justify-center mx-auto mb-6" style="width: 4rem; height: 4rem;">
            <div class="border-2 border-hi/30 border-t-hi rounded-full animate-spin" style="width: 1.5rem; height: 1.5rem;" />
          </div>
          <h2 class="font-display font-bold text-2xl mb-2">Decrypting...</h2>
          <p class="text-t3 text-sm">AES-256-GCM decryption running in your browser.</p>

          <!-- Fake progress steps -->
          <div class="mt-8 text-left space-y-2 font-mono text-xs" style="max-width: 20rem; margin-left: auto; margin-right: auto;">
            <div class="flex items-center gap-2 text-emerald">
              <Check :size="12" /> Extracted key from URL fragment
            </div>
            <div class="flex items-center gap-2 text-emerald">
              <Check :size="12" /> Fetched ciphertext from server
            </div>
            <div class="flex items-center gap-2 text-hi animate-pulse">
              <div class="border border-hi/30 border-t-hi rounded-full animate-spin" style="width: .75rem; height: .75rem;" />
              Decrypting with AES-256-GCM...
            </div>
          </div>
        </div>

        <!-- ═══ SUCCESS STATE ═══ -->
        <div v-else-if="state === 'success'">
          <!-- Header with branding -->
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center" style="gap: 0.75rem;">
              <div class="rounded-xl bg-emerald/10 border border-emerald/20 flex items-center justify-center shrink-0" style="width: 2.5rem; height: 2.5rem;">
                <Check :size="20" class="text-emerald" />
              </div>
              <div class="min-w-0">
                <h2 class="font-display font-bold text-lg sm:text-xl">Secret decrypted</h2>
                <p class="text-t3 text-xs font-mono truncate">End-to-end encrypted — the server never saw the key</p>
              </div>
            </div>
          </div>

          <!-- Status badges -->
          <div class="flex flex-wrap gap-2 mb-4">
            <div class="flex items-center gap-1.5 bg-s2 border border-b1 rounded-lg px-3 py-1.5 text-xs font-mono">
              <FileText :size="12" class="text-t3 shrink-0" />
              <span class="text-t2 truncate">{{ filename }}</span>
            </div>
            <div class="flex items-center gap-1.5 bg-s2 border border-b1 rounded-lg px-3 py-1.5 text-xs font-mono">
              <Clock :size="12" class="text-t3 shrink-0" />
              <span class="text-t2">{{ envLines.filter(l => l.type === 'var').length }} vars</span>
            </div>
            <div
              v-if="wasLastRead"
              class="flex items-center gap-1.5 bg-red/8 border border-red/20 rounded-lg px-3 py-1.5 text-xs font-mono text-red"
            >
              <Trash2 :size="12" class="shrink-0" />
              <span class="hidden sm:inline">Link destroyed — last read</span>
              <span class="sm:hidden">Destroyed</span>
            </div>
            <div
              v-else-if="readsLeft > 0"
              class="flex items-center gap-1.5 bg-amber/8 border border-amber/20 rounded-lg px-3 py-1.5 text-xs font-mono text-amber"
            >
              <Eye :size="12" class="shrink-0" />
              {{ readsLeft }} {{ readsLeft === 1 ? 'read' : 'reads' }} left
            </div>
          </div>

          <!-- Content card -->
          <div class="bg-s2 border border-b1 rounded-xl sm:rounded-2xl overflow-hidden">
            <!-- Toolbar -->
            <div class="flex flex-wrap items-center gap-2 px-3 sm:px-4 py-2.5 sm:py-3 border-b border-b1 bg-s3">
              <div class="flex items-center gap-2 mr-auto min-w-0">
                <span class="font-mono text-xs text-t3 truncate">{{ filename }}</span>
                <span class="text-t3 hidden sm:inline">·</span>
                <button
                  @click="revealContent = !revealContent"
                  class="hidden sm:flex items-center gap-1.5 text-xs text-t3 hover:text-t2 transition-colors cursor-pointer whitespace-nowrap"
                >
                  <Eye v-if="!revealContent" :size="12" />
                  <EyeOff v-else :size="12" />
                  {{ revealContent ? 'Mask' : 'Reveal' }}
                </button>
              </div>
              <div class="flex items-center gap-1.5 sm:gap-2">
                <!-- Mobile reveal toggle -->
                <button
                  @click="revealContent = !revealContent"
                  class="sm:hidden flex items-center gap-1 text-xs font-mono text-t3 bg-s4 border border-b1 px-2 py-1 rounded-md hover:text-t1 transition-all cursor-pointer"
                  :title="revealContent ? 'Mask values' : 'Reveal values'"
                >
                  <Eye v-if="!revealContent" :size="12" />
                  <EyeOff v-else :size="12" />
                </button>
                <button
                  @click="copyContent"
                  class="flex items-center gap-1.5 text-xs font-mono text-t3 bg-s4 border border-b1 px-2 sm:px-2.5 py-1 rounded-md hover:text-t1 hover:border-b2 transition-all cursor-pointer"
                >
                  <Check v-if="copied" :size="12" class="text-emerald" />
                  <Copy v-else :size="12" />
                  <span class="hidden sm:inline">{{ copied ? 'Copied!' : 'Copy' }}</span>
                </button>
                <button
                  @click="downloadFile"
                  class="flex items-center gap-1.5 text-xs font-mono text-t3 bg-s4 border border-b1 px-2 sm:px-2.5 py-1 rounded-md hover:text-t1 hover:border-b2 transition-all cursor-pointer"
                >
                  <Download :size="12" />
                  <span class="hidden sm:inline">Download</span>
                </button>
              </div>
            </div>

            <!-- Env content -->
            <div class="px-3 sm:px-4 py-3 sm:py-4 font-mono text-[12px] sm:text-[13px] leading-[1.9] sm:leading-[2] overflow-x-auto max-h-[420px] sm:max-h-[480px] overflow-y-auto">
              <div v-for="(line, i) in envLines" :key="i">
                <!-- Comment -->
                <div v-if="line.type === 'comment'" class="text-t3">{{ line.text }}</div>
                <!-- Key=Value -->
                <div v-else-if="line.type === 'var'" class="whitespace-nowrap">
                  <span class="text-hi">{{ line.key }}</span>
                  <span class="text-t3">=</span>
                  <span :class="revealContent ? 'text-t1' : 'text-amber'">
                    {{ revealContent ? line.value : maskValue(line.value) }}
                  </span>
                </div>
                <!-- Plain text -->
                <div v-else class="text-t2">{{ line.text }}</div>
              </div>

            </div>
          </div>

          <!-- CTA below content -->
          <div class="cta-card rounded-xl sm:rounded-2xl border border-hi/25 overflow-hidden" style="margin-top: 1.25rem; background: linear-gradient(135deg, color-mix(in srgb, var(--t-hi) 8%, var(--t-s2)), var(--t-s2));">
            <!-- Glow line top -->
            <div style="height: 2px; background: linear-gradient(90deg, transparent 5%, var(--t-hi), transparent 95%); opacity: 0.6;" />
            <div style="padding: 1.25rem 1.25rem 1rem;">
              <div class="flex items-center" style="gap: 0.5rem; margin-bottom: 0.75rem;">
                <Terminal :size="16" class="text-hi" />
                <span class="font-display font-bold text-t1" style="font-size: 0.9375rem;">Share secrets like this from your terminal</span>
              </div>
              <p class="text-t3" style="font-size: 0.75rem; margin-bottom: 1rem; line-height: 1.5;">
                Your teammate used <span class="text-hi font-semibold">envshare</span> to send this. One command to encrypt, share, and auto-destroy.
              </p>

              <!-- Install tabs -->
              <div class="flex items-center flex-wrap" style="gap: 0.375rem; margin-bottom: 0.625rem;">
                <button
                  v-for="(method, i) in installMethods"
                  :key="'cta'+method.label"
                  @click="activeInstall = i; copiedInstall = false"
                  class="rounded-lg font-mono font-medium transition-all cursor-pointer"
                  :class="activeInstall === i ? 'bg-hi/20 text-hi border border-hi/30' : 'text-t3 hover:text-t2 bg-s3 border border-b1'"
                  style="padding: 0.375rem 0.75rem; font-size: 0.75rem;"
                >
                  {{ method.label }}
                </button>
              </div>

              <!-- Command box -->
              <div
                @click="copyInstallCommand"
                class="flex items-center font-mono cursor-pointer transition-all rounded-lg border border-hi/20 hover:border-hi/40 bg-s1"
                style="padding: 0.625rem 0.875rem; gap: 0.5rem; font-size: 0.8125rem;"
              >
                <span class="text-hi font-bold">$</span>
                <span class="text-t1 truncate">{{ installMethods[activeInstall].command }}</span>
                <button
                  @click.stop="copyInstallCommand"
                  class="shrink-0 flex items-center font-semibold rounded-md transition-all cursor-pointer"
                  :class="copiedInstall ? 'text-emerald bg-emerald/10' : 'text-s1 bg-hi hover:opacity-90'"
                  style="gap: 0.25rem; font-size: 0.6875rem; padding: 0.375rem 0.75rem; margin-left: auto;"
                >
                  <Check v-if="copiedInstall" :size="12" />
                  <Copy v-else :size="12" />
                  {{ copiedInstall ? 'Copied!' : 'Copy' }}
                </button>
              </div>
            </div>
          </div>

        </div>

        <!-- ═══ ERROR STATE ═══ -->
        <div v-else-if="state === 'error'" class="text-center" style="max-width: 28rem; margin-left: auto; margin-right: auto;">
          <div class="rounded-2xl bg-red/8 border border-red/20 flex items-center justify-center mx-auto mb-6" style="width: 4rem; height: 4rem;">
            <AlertTriangle :size="28" class="text-red" />
          </div>

          <template v-if="errorType === 'NOT_FOUND'">
            <h2 class="font-display font-bold text-2xl mb-2">Link not found</h2>
            <p class="text-t3 text-sm mb-6">
              This link has already been used, expired, or never existed. Secrets are permanently deleted after being read.
            </p>
          </template>
          <template v-else-if="errorType === 'EXPIRED'">
            <h2 class="font-display font-bold text-2xl mb-2">Link expired</h2>
            <p class="text-t3 text-sm mb-6">
              This link's TTL has passed. The encrypted data was automatically deleted.
            </p>
          </template>
          <template v-else-if="errorType === 'INVALID_KEY'">
            <h2 class="font-display font-bold text-2xl mb-2">Invalid link</h2>
            <p class="text-t3 text-sm mb-6">
              The decryption key in the URL is missing or malformed. Make sure you copied the full link including the # fragment.
            </p>
          </template>
          <template v-else>
            <h2 class="font-display font-bold text-2xl mb-2">Something went wrong</h2>
            <p class="text-t3 text-sm mb-6">
              Could not connect to the server. Please check your connection and try again.
            </p>
          </template>

          <a
            href="/"
            class="inline-flex items-center gap-2 bg-s2 border border-b1 text-t2 font-medium text-sm px-6 py-3 rounded-xl hover:border-b2 hover:text-t1 transition-all"
          >
            ← Back to envshare
          </a>
        </div>

        <!-- ═══ DESTROYED STATE ═══ -->
        <div v-else-if="state === 'destroyed'" class="text-center" style="max-width: 28rem; margin-left: auto; margin-right: auto;">
          <div class="rounded-2xl bg-red/8 border border-red/20 flex items-center justify-center mx-auto mb-6" style="width: 4rem; height: 4rem;">
            <Trash2 :size="28" class="text-red" />
          </div>
          <h2 class="font-display font-bold text-2xl mb-2">Secret destroyed</h2>
          <p class="text-t3 text-sm mb-6">
            This link has already been opened and the encrypted data was permanently deleted from the server.
          </p>
          <a
            href="/"
            class="inline-flex items-center gap-2 bg-s2 border border-b1 text-t2 font-medium text-sm px-6 py-3 rounded-xl hover:border-b2 hover:text-t1 transition-all"
          >
            ← Back to envshare
          </a>
        </div>

      </div>
    </div>

    <!-- Install Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div
          v-if="showInstallModal"
          class="fixed inset-0 z-[100] flex items-center justify-center"
          style="padding: 1rem;"
          @click.self="showInstallModal = false"
        >
          <!-- Backdrop -->
          <div class="absolute inset-0" style="background: rgba(0, 0, 0, 0.6); backdrop-filter: blur(4px);" />

          <!-- Modal card -->
          <div
            class="relative rounded-2xl border border-hi/25 overflow-hidden modal-content"
            style="width: 100%; max-width: 28rem; background: var(--t-s1);"
          >
            <!-- Glow line -->
            <div style="height: 2px; background: linear-gradient(90deg, transparent 5%, var(--t-hi), transparent 95%);" />

            <!-- Close button -->
            <button
              @click="showInstallModal = false"
              class="absolute text-t3 hover:text-t1 transition-colors cursor-pointer"
              style="top: 0.875rem; right: 0.875rem; padding: 0.25rem;"
            >
              <X :size="18" />
            </button>

            <div style="padding: 1.75rem 1.5rem 1.5rem;">
              <!-- Header -->
              <div class="flex items-center" style="gap: 0.5rem; margin-bottom: 0.375rem;">
                <span class="rounded-full bg-hi" style="width: 0.5rem; height: 0.5rem;" />
                <span class="font-mono font-bold text-hi" style="font-size: 0.8125rem;">envshare</span>
              </div>
              <h3 class="font-display font-bold text-t1" style="font-size: 1.25rem; margin-bottom: 0.5rem;">
                Share secrets from your terminal
              </h3>
              <p class="text-t3" style="font-size: 0.8125rem; line-height: 1.6; margin-bottom: 1.25rem;">
                Your teammate used envshare to send this. Install it and start sharing encrypted secrets in seconds.
              </p>

              <!-- Selling points -->
              <div style="margin-bottom: 1.25rem; gap: 0.625rem;" class="flex flex-col">
                <div class="flex items-center" style="gap: 0.625rem;">
                  <div class="shrink-0 flex items-center justify-center rounded-lg bg-hi/10" style="width: 1.75rem; height: 1.75rem;">
                    <Zap :size="14" class="text-hi" />
                  </div>
                  <div>
                    <span class="text-t1 font-medium" style="font-size: 0.8125rem;">Install in 5 seconds</span>
                    <span class="text-t3" style="font-size: 0.75rem;"> — one command, done</span>
                  </div>
                </div>
                <div class="flex items-center" style="gap: 0.625rem;">
                  <div class="shrink-0 flex items-center justify-center rounded-lg bg-hi/10" style="width: 1.75rem; height: 1.75rem;">
                    <UserX :size="14" class="text-hi" />
                  </div>
                  <div>
                    <span class="text-t1 font-medium" style="font-size: 0.8125rem;">No login, no account</span>
                    <span class="text-t3" style="font-size: 0.75rem;"> — just works</span>
                  </div>
                </div>
                <div class="flex items-center" style="gap: 0.625rem;">
                  <div class="shrink-0 flex items-center justify-center rounded-lg bg-hi/10" style="width: 1.75rem; height: 1.75rem;">
                    <Shield :size="14" class="text-hi" />
                  </div>
                  <div>
                    <span class="text-t1 font-medium" style="font-size: 0.8125rem;">AES-256-GCM encryption</span>
                    <span class="text-t3" style="font-size: 0.75rem;"> — zero-knowledge</span>
                  </div>
                </div>
                <div class="flex items-center" style="gap: 0.625rem;">
                  <div class="shrink-0 flex items-center justify-center rounded-lg bg-hi/10" style="width: 1.75rem; height: 1.75rem;">
                    <Timer :size="14" class="text-hi" />
                  </div>
                  <div>
                    <span class="text-t1 font-medium" style="font-size: 0.8125rem;">Self-destructing links</span>
                    <span class="text-t3" style="font-size: 0.75rem;"> — TTL + read limits</span>
                  </div>
                </div>
              </div>

              <!-- Divider -->
              <div class="border-t border-b1" style="margin-bottom: 1.25rem;" />

              <!-- Install tabs -->
              <div class="flex items-center flex-wrap" style="gap: 0.375rem; margin-bottom: 0.75rem;">
                <button
                  v-for="(method, i) in installMethods"
                  :key="'modal'+method.label"
                  @click="activeInstall = i; copiedInstall = false"
                  class="rounded-lg font-mono font-medium transition-all cursor-pointer"
                  :class="activeInstall === i ? 'bg-hi/20 text-hi border border-hi/30' : 'text-t3 hover:text-t2 bg-s2 border border-b1'"
                  style="padding: 0.4375rem 0.875rem; font-size: 0.75rem;"
                >
                  {{ method.label }}
                </button>
              </div>

              <!-- Command box -->
              <div
                @click="copyInstallCommand"
                class="flex items-center font-mono cursor-pointer transition-all rounded-xl border border-hi/20 hover:border-hi/40 bg-s2"
                style="padding: 0.75rem 1rem; gap: 0.5rem; font-size: 0.8125rem;"
              >
                <span class="text-hi font-bold">$</span>
                <span class="text-t1 truncate">{{ installMethods[activeInstall].command }}</span>
                <button
                  @click.stop="copyInstallCommand"
                  class="shrink-0 flex items-center font-semibold rounded-lg transition-all cursor-pointer"
                  :class="copiedInstall ? 'text-emerald bg-emerald/10' : 'text-s1 bg-hi hover:opacity-90'"
                  style="gap: 0.25rem; font-size: 0.75rem; padding: 0.4375rem 0.875rem; margin-left: auto;"
                >
                  <Check v-if="copiedInstall" :size="12" />
                  <Copy v-else :size="12" />
                  {{ copiedInstall ? 'Copied!' : 'Copy' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Footer CTA -->
    <footer class="border-t border-b1" style="background: linear-gradient(180deg, var(--t-s1), color-mix(in srgb, var(--t-hi) 4%, var(--t-s1)));">
      <div style="max-width: 42rem; margin: 0 auto; padding: 1.5rem 1rem;" class="text-center">
        <div class="flex items-center justify-center" style="gap: 0.5rem; margin-bottom: 0.5rem;">
          <span class="rounded-full bg-hi pulse-glow" style="width: 0.4rem; height: 0.4rem;" />
          <span class="font-mono font-bold text-t1" style="font-size: 0.8125rem;">envshare</span>
        </div>
        <p class="text-t3" style="font-size: 0.75rem; margin-bottom: 1rem;">
          Share .env files with your team securely. Zero-knowledge, self-destructing, open source.
        </p>
        <button
          @click="showInstallModal = true"
          class="inline-flex items-center font-mono font-semibold text-s1 bg-hi rounded-lg transition-all hover:opacity-90 cursor-pointer"
          style="gap: 0.375rem; padding: 0.5rem 1.25rem; font-size: 0.75rem;"
        >
          <Terminal :size="14" />
          Get the CLI →
        </button>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.cta-card {
  animation: fadeIn 0.4s ease-out both;
  animation-delay: 0.3s;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(0.5rem); }
  to { opacity: 1; transform: translateY(0); }
}

/* Modal transitions */
.modal-enter-active { transition: opacity 0.2s ease; }
.modal-leave-active { transition: opacity 0.15s ease; }
.modal-enter-from,
.modal-leave-to { opacity: 0; }

.modal-enter-active .modal-content {
  animation: modalIn 0.25s ease-out;
}
.modal-leave-active .modal-content {
  animation: modalOut 0.15s ease-in;
}

@keyframes modalIn {
  from { opacity: 0; transform: scale(0.95) translateY(0.5rem); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}
@keyframes modalOut {
  from { opacity: 1; transform: scale(1) translateY(0); }
  to { opacity: 0; transform: scale(0.95) translateY(0.5rem); }
}
</style>
