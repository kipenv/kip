// Base58 alphabet (Bitcoin-style)
const BASE58_ALPHABET = '123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz'

export function decodeBase58(input: string): Uint8Array {
  const bytes: number[] = [0]
  for (const char of input) {
    const idx = BASE58_ALPHABET.indexOf(char)
    if (idx === -1) throw new Error(`Invalid base58 character: ${char}`)
    let carry = idx
    for (let j = 0; j < bytes.length; j++) {
      carry += bytes[j] * 58
      bytes[j] = carry & 0xff
      carry >>= 8
    }
    while (carry > 0) {
      bytes.push(carry & 0xff)
      carry >>= 8
    }
  }
  // Leading zeros
  for (const char of input) {
    if (char !== '1') break
    bytes.push(0)
  }
  return new Uint8Array(bytes.reverse())
}

export async function decryptAesGcm(
  key: Uint8Array,
  nonce: Uint8Array,
  ciphertext: Uint8Array
): Promise<Uint8Array> {
  const cryptoKey = await crypto.subtle.importKey(
    'raw',
    key,
    { name: 'AES-GCM' },
    false,
    ['decrypt']
  )
  const decrypted = await crypto.subtle.decrypt(
    { name: 'AES-GCM', iv: nonce },
    cryptoKey,
    ciphertext
  )
  return new Uint8Array(decrypted)
}

export async function decryptWithPassword(
  password: string,
  salt: Uint8Array,
  nonce: Uint8Array,
  ciphertext: Uint8Array
): Promise<Uint8Array> {
  // Derive key from password using PBKDF2 (Argon2id is not available in Web Crypto,
  // server would need to support PBKDF2 as fallback for web, or we use a JS argon2 lib)
  const encoder = new TextEncoder()
  const passwordKey = await crypto.subtle.importKey(
    'raw',
    encoder.encode(password),
    'PBKDF2',
    false,
    ['deriveKey']
  )
  const derivedKey = await crypto.subtle.deriveKey(
    { name: 'PBKDF2', salt, iterations: 100000, hash: 'SHA-256' },
    passwordKey,
    { name: 'AES-GCM', length: 256 },
    false,
    ['decrypt']
  )
  const decrypted = await crypto.subtle.decrypt(
    { name: 'AES-GCM', iv: nonce },
    derivedKey,
    ciphertext
  )
  return new Uint8Array(decrypted)
}

export function parseSecretUrl(): { id: string; key: string } | null {
  const path = window.location.pathname
  const hash = window.location.hash.slice(1) // remove #

  const match = path.match(/\/s\/(.+)/)
  if (!match || !hash) return null

  return { id: match[1], key: hash }
}

export interface SecretResponse {
  ciphertext: string // base64
  nonce: string // base64
  salt?: string // base64
  filename: string
  reads_left: number
  password_protected: boolean
}

export async function fetchSecret(baseUrl: string, id: string): Promise<SecretResponse> {
  const res = await fetch(`${baseUrl}/api/v1/secret/${id}`)
  if (res.status === 404) {
    throw new Error('NOT_FOUND')
  }
  if (!res.ok) {
    throw new Error('SERVER_ERROR')
  }
  return res.json()
}

function base64ToBytes(b64: string): Uint8Array {
  const binary = atob(b64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }
  return bytes
}

export async function decryptSecret(
  keyBase58: string,
  secret: SecretResponse,
  password?: string
): Promise<string> {
  const key = decodeBase58(keyBase58)
  if (key.length !== 32) {
    throw new Error('INVALID_KEY')
  }

  let ciphertext = base64ToBytes(secret.ciphertext)
  let nonce = base64ToBytes(secret.nonce)

  // If password-protected, decrypt outer layer first
  if (secret.password_protected && password) {
    const salt = secret.salt ? base64ToBytes(secret.salt) : new Uint8Array(16)
    const outerPlaintext = await decryptWithPassword(password, salt, nonce, ciphertext)

    // Unpack: first 12 bytes = inner nonce, rest = inner ciphertext
    nonce = outerPlaintext.slice(0, 12)
    ciphertext = outerPlaintext.slice(12)
  }

  // Decrypt inner layer with URL key
  const plaintext = await decryptAesGcm(key, nonce, ciphertext)
  return new TextDecoder().decode(plaintext)
}
