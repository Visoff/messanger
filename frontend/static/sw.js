const CACHE_NAME = 'messenger-cache-v1';
const PRECACHE_URLS = [
  '/',
  '/favicon.svg',
];

self.addEventListener('install', (event) => {
  self.skipWaiting();
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(PRECACHE_URLS).catch((err) => {
          console.error(err);
      });
    })
  );
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    Promise.all([
      caches.keys().then((cacheNames) => {
        return Promise.all(
          cacheNames
            .filter((name) => name !== CACHE_NAME)
            .map((name) => caches.delete(name))
        );
      }),
      self.clients.claim(),
    ])
  );
});

self.addEventListener('fetch', (event) => {
  if (event.request.mode === 'navigate') {
    event.respondWith(
      caches.match(event.request).then((cached) => {
        const fetchPromise = fetch(event.request).then((response) => {
          if (response && response.status === 200 && response.type === 'basic') {
            const cloned = response.clone();
            event.waitUntil(
              caches.open(CACHE_NAME).then((cache) => cache.put(event.request, cloned))
            );
          }
          return response;
        });
        return cached || fetchPromise;
      })
    );
    return;
  }

  event.respondWith(
    fetch(event.request).then((response) => {
      if (response && response.status === 200 && response.type === 'basic') {
        const cloned = response.clone();
        event.waitUntil(
          caches.open(CACHE_NAME).then((cache) => cache.put(event.request, cloned))
        );
      }
      return response;
    }).catch(() => {
      return caches.match(event.request);
    })
  );
});

self.addEventListener('push', (event) => {
  const data = event.data.json();
  const options = {
    body: data.body,
    icon: '/favicon.svg',
    data: data.data,
  };
  event.waitUntil(
      self.registration.showNotification(data.title, options)
  );
});
