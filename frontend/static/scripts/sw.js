self.addEventListener('push', (event) => {
    console.log(event);
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
