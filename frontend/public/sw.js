/*
 * Fanoos PWA service worker (hand-written — no build-time tooling).
 *
 * Strategies:
 *  - navigations: network-first, fall back to cached index.html (offline shell).
 *  - static assets (hashed bundles): cache-first (immutable).
 *  - GET device-state APIs: network-first with 3s timeout, fall back to the
 *    last cached response so the last-known circuit state still renders
 *    when the network is flaky.
 *  - everything else (POST commands, dataviews): network only.
 */
const CACHE = 'fanoos-v1';
const STATE_TIMEOUT_MS = 3000;

const STATE_API = /\/api\/devices\/\d+\/state\/(light|controller|sensor)$/;

self.addEventListener('install', () => {
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) => Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k)))),
  );
});

async function networkFirst(request) {
  const cache = await caches.open(CACHE);
  try {
    const response = await fetch(request);
    if (response.ok) cache.put(request, response.clone());
    return response;
  } catch {
    const cached = await cache.match(request);
    if (cached) return cached;
    return new Response(JSON.stringify({ known: false }), {
      status: 503,
      headers: { 'Content-Type': 'application/json' },
    });
  }
}

async function stateNetworkFirst(request) {
  const cache = await caches.open(CACHE);
  try {
    const response = await Promise.race([
      fetch(request),
      new Promise((_, reject) => setTimeout(() => reject(new Error('timeout')), STATE_TIMEOUT_MS)),
    ]);
    if (response.ok) cache.put(request, response.clone());
    return response;
  } catch {
    const cached = await cache.match(request);
    return (
      cached ||
      new Response(JSON.stringify({ on: false, online: false, ts: 0, known: false }), {
        status: 503,
        headers: { 'Content-Type': 'application/json' },
      })
    );
  }
}

async function cacheFirst(request) {
  const cache = await caches.open(CACHE);
  const cached = await cache.match(request);
  if (cached) return cached;
  const response = await fetch(request);
  if (response.ok && request.url.startsWith(self.location.origin)) {
    cache.put(request, response.clone());
  }
  return response;
}

self.addEventListener('fetch', (event) => {
  const { request } = event;
  if (request.method !== 'GET') return;

  const url = new URL(request.url);
  if (url.origin !== self.location.origin) return;

  if (request.mode === 'navigate') {
    event.respondWith(networkFirst(request));
    return;
  }
  if (STATE_API.test(url.pathname)) {
    event.respondWith(stateNetworkFirst(request));
    return;
  }
  if (url.pathname.startsWith('/api/')) return;

  event.respondWith(cacheFirst(request));
});
