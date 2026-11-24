/* Cache the public atlas only. Reviewer requests and submissions always go to the network. */
const scope = new URL(self.registration.scope);
const prefix = 'nirikshan-atlas-' + scope.pathname + '-';
const version = prefix + '1';
const relative = path => new URL(path, scope).href;
const shell = ['./', 'index.html', 'style.css', 'app.js', 'analysis.js', 'share.js',
  'projects.json', 'manifest.webmanifest', 'mark.svg', 'icon-192.png', 'icon-512.png', 'vendor/qrcode.js'];
const assets = new Set(shell.map(relative));
const essentials = ['index.html', 'style.css', 'app.js', 'analysis.js', 'share.js', 'projects.json'];

self.addEventListener('install', event => event.waitUntil((async () => {
  const cache = await caches.open(version);
  await cache.addAll(essentials.map(relative));
  await Promise.allSettled(shell.filter(path => !essentials.includes(path)).map(path => cache.add(relative(path))));
  await self.skipWaiting();
})()));

self.addEventListener('activate', event => event.waitUntil((async () => {
  for (const key of await caches.keys()) {
    if (key.startsWith(prefix) && key !== version) await caches.delete(key);
  }
  await self.clients.claim();
})()));

function eligible(request) {
  if (request.method !== 'GET' || request.headers.has('Authorization')) return false;
  const url = new URL(request.url);
  if (url.origin !== scope.origin) return false;
  const isPublicPage = request.mode === 'navigate' &&
    [scope.pathname, new URL('index.html', scope).pathname].includes(url.pathname);
  if (isPublicPage || assets.has(url.href)) return true;
  if (url.pathname === '/api/projects') return !url.search;
  return url.pathname === '/api/evidence' && /^[\w-]{1,100}$/.test(url.searchParams.get('project') || '') &&
    [...url.searchParams.keys()].length === 1;
}

self.addEventListener('fetch', event => {
  if (!eligible(event.request)) return;
  event.respondWith((async () => {
    const request = event.request;
    const url = new URL(request.url);
    const cache = await caches.open(version);
    const isApi = url.pathname.startsWith('/api/');
    const isData = isApi || url.href === relative('projects.json');
    const key = request.mode === 'navigate' ? relative('index.html') : request.url;
    try {
      const response = await fetch(request);
      const isJson = (response.headers.get('content-type') || '').includes('application/json');
      if (response.ok && (!isData || isJson)) {
        const saved = response.clone();
        if (isData) {
          const headers = new Headers(saved.headers);
          headers.set('X-Atlas-Saved', new Date().toISOString());
          await cache.put(key, new Response(await saved.arrayBuffer(), {headers}));
        } else {
          await cache.put(key, saved);
        }
      }
      return response;
    } catch {
      const saved = await cache.match(key);
      if (!saved) return Response.error();
      if (isData && event.clientId) {
        const client = await self.clients.get(event.clientId);
        client?.postMessage({type: 'atlas-cached', saved: saved.headers.get('X-Atlas-Saved') || ''});
      }
      return saved;
    }
  })());
});
