import { API_URL } from "$lib/api/env";

async function subscribeToPush(registration: ServiceWorkerRegistration) {
	const token = localStorage.getItem("token");
	if (!token) return;
	const permission = await Notification.requestPermission();
	if (permission !== "granted") return;

	const vapidPublicKey = await fetch(
		`${API_URL}/pubsub/push/pubkey`,
	).then((r) => r.text());

	const subscription =
		(await registration.pushManager.getSubscription()) ??
		(await registration.pushManager.subscribe({
			userVisibleOnly: true,
			applicationServerKey: vapidPublicKey,
		}));
	console.log(subscription);
	await fetch(`${API_URL}/pubsub/push/subscribe`, {
		method: "POST",
		headers: {
			Authorization: `Bearer ${token}`,
			"Content-Type": "application/json",
		},
		body: JSON.stringify(subscription),
	});
}

export function init() {
	if (!("serviceWorker" in navigator)) return;

	navigator.serviceWorker
		.register("/sw.js")
		.then((registration) => {
			console.log(
				"Service Worker registration successful with scope: ",
				registration.scope,
			);
			subscribeToPush(registration);
		})
		.catch((err) => {
			console.log("Service Worker registration failed: ", err);
		});
}
