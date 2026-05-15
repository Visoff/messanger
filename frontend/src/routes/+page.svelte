<script lang="ts">
    import { API_URL } from "$lib/api/env";
    import ChatList from "$lib/components/ChatList.svelte";
    import { onMount } from "svelte";
    import { getMe } from "$lib/api/auth";
    import ChatCreationModel from "$lib/components/ChatCreationModel.svelte";
    import ChatPage from "$lib/components/ChatPage.svelte";
    import { extractFromSearchParams } from "$lib/index";
    import { user } from "$lib/stores/user";

    async function getServiceWorkerRegistration() {
        if (navigator.serviceWorker.controller) {
            return navigator.serviceWorker.ready;
        }
        const registration =
            await navigator.serviceWorker.register("/scripts/sw.js");
        return registration;
    }

    async function subscribeToPush() {
        if ("serviceWorker" in navigator === false) return;
        const registration = await getServiceWorkerRegistration();
        const sub = await registration.pushManager.getSubscription();
        if (sub) {
            return;
        }
        const permission = await Notification.requestPermission();
        if (permission !== "granted") return;

        const vapidPublicKey = await fetch(
            `${API_URL}/pubsub/push/pubkey`,
        ).then((r) => r.text());
        console.log(vapidPublicKey);
        const subscription = await registration.pushManager.subscribe({
            userVisibleOnly: true,
            applicationServerKey: vapidPublicKey,
        });
        const token = localStorage.getItem("token");
        await fetch(`${API_URL}/pubsub/push/subscribe`, {
            method: "POST",
            headers: {
                Authorization: `Bearer ${token}`,
                "Content-Type": "application/json",
            },
            body: JSON.stringify(subscription),
        });
    }

    let chat_id: string | undefined = $state(extractFromSearchParams("chat_id"));

    onMount(() => {
        const token = localStorage.getItem("token");
        if (!token) {
            window.location.href = "/login";
        } else {
            console.log(token);
            getMe().then((resp) => {
                if ("error" in resp) {
                    console.error(resp.error);
                    return
                }
                user.set(resp)
            });
        }

        const prompt_notifications = () => {
            window.removeEventListener("click", prompt_notifications);
            subscribeToPush();
        };

        window.addEventListener("click", prompt_notifications);
    });

    let chatPageRef: ChatPage | undefined = $state();

    function clearChat() {
        chat_id = undefined
        chatPageRef?.resetTopic();
        const url = new URL(location.href);
        url.searchParams.delete("chat_id");
        url.searchParams.delete("topic_id");
        history.pushState(null, "", url);
    }
</script>

<main class="flex h-screen bg-white">
    <div class="md:w-80 md:relative min-w-80 border-r-gray-100 border-r flex flex-col bg-white w-full absolute">
        <div class="flex justify-between items-center p-4 border-b-gray-100 border-b">
            <span class="font-bold text-lg">Messanger</span>
            <ChatCreationModel />
        </div>
        <ChatList onselect={(id) => {chat_id = id; chatPageRef?.resetTopic();}} current_chat_id={chat_id} />
    </div>

    {#if chat_id}
        <ChatPage chat_id={chat_id} onclose={clearChat} bind:this={chatPageRef} />
    {:else}
        <div class="w-full flex flex-col items-center justify-center bg-gray-50">
            <div class="mb-4 opacity-50">
                <svg
                    viewBox="0 0 24 24"
                    width="64"
                    height="64"
                    fill="currentColor"
                >
                    <path
                        d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm0 14H6l-2 2V4h16v12z"
                    />
                </svg>
            </div>
            <p>Select a chat to start messaging</p>
        </div>
    {/if}
</main>
