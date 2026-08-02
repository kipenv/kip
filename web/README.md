# kip web

Static site for kip: the landing page and — the part that matters — the
in-browser decrypt page, so a recipient can read a shared secret without
installing the CLI.

Astro 6 (SSG) + Vue 3 islands + Tailwind 4. No SSR, no backend: the output is
plain files.

## Commands

Run from this directory:

| Command | Action |
| :--- | :--- |
| `npm install` | Install dependencies (Node >= 22.12) |
| `npm run dev` | Dev server on `localhost:4321` |
| `npm run build` | Build to `./dist/` |
| `npm run preview` | Serve the built output locally |

## Decrypt page

`src/pages/decrypt.astro` + `src/components/DecryptPage.vue` are the security-
sensitive part of this codebase. Two invariants:

1. **The key never leaves the browser.** It arrives in the URL fragment
   (`/s/<id>#<key>`), which browsers do not send to servers. The page reads it
   from `location.hash` and decrypts with `SubtleCrypto` (`src/lib/crypto.ts`).
   Nothing may put that value into a request, a log, or an analytics call.
2. **No third-party scripts.** A tag manager or font CDN on this page would be
   able to read the fragment. Keep it dependency-free at runtime.

`src/pages/demo/decrypt.astro` renders the same UI with canned data, for
previewing the layout without a live secret.

## Routing note

Share links look like `/s/<id>#<key>`. There is no file per id — every `/s/*`
path must serve the decrypt page, which reads the id from `location.pathname`.
`nginx.conf` does that with `try_files /decrypt/index.html`; any other host
needs the equivalent rewrite, or shared links 404.

## API base URL

The decrypt page calls the kip server to fetch the ciphertext. It defaults to
same-origin, which is what the single-domain deploy in
`deploy/docker-compose.prod.yml` expects. Set `PUBLIC_API_URL` at build time
only when the API lives on another origin.
