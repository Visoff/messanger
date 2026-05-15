<script lang="ts">
    import { createChat, fetchChat } from "$lib/api/chats";
    import { createTopic, fetchTopic } from "$lib/api/topics";
    import { API_URL } from "$lib/api/env";
    import ChatList from "$lib/components/ChatList.svelte";
    import MessageList from "$lib/components/MessageList.svelte";
    import TopicList from "$lib/components/TopicList.svelte";
    import { selectedChatId, selectedTopicId } from "$lib/stores/chat";
    import { onMount } from "svelte";
    import { getMe } from "$lib/api/auth";
    import ChatCreationModel from "$lib/components/ChatCreationModel.svelte";

    async function getServiceWorkerRegistration() {
        if (navigator.serviceWorker.controller) {
            return navigator.serviceWorker.ready;
        }
        const registration = await navigator.serviceWorker.register("/scripts/sw.js");
        return registration;
    }

    async function subscribeToPush() {
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
    let title: string = $state("");

    $effect(() => {
        (async () => {
            if (!$selectedChatId) {
                return;
            }
            if (!$selectedTopicId) {
                const resp = await fetchChat($selectedChatId);
                if ("error" in resp) {
                    console.error(resp.error);
                    return;
                }
                title = resp.title;
            } else {
                const resp = await fetchTopic($selectedTopicId);
                if ("error" in resp) {
                    console.error(resp.error);
                    return;
                }
                title = resp.title;
            }
        })();
    });

    onMount(() => {
        const token = localStorage.getItem("token");
        if (!token) {
            window.location.href = "/login";
        } else {
            console.log(token);
            getMe().then(resp => console.log(resp));
        }

        const url = new URL(window.location.href);
        const chat_id = url.searchParams.get("chat_id");
        if (chat_id) {
            selectedChatId.set(chat_id);
            console.log("chat_id", chat_id);
        }

        const topic_id = url.searchParams.get("topic_id");
        if (topic_id) {
            selectedTopicId.set(topic_id);
        }

        const stream = new EventSource(`${API_URL}/pubsub/sse`);
        stream.addEventListener("message", (e) => {
            console.log(e.data);
        });

        const prompt_notifications = () => {
            window.removeEventListener("click", prompt_notifications);
            subscribeToPush();
        }

        window.addEventListener("click", prompt_notifications);
    });

    async function createChatEvent() {
        const title = prompt("Chat title");
        if (title) {
            const resp = await createChat(title);
            if ("error" in resp) {
                console.error(resp.error);
                return;
            }
            selectedChatId.set(resp.id);
        }
    }

    function clearChat() {
        selectedChatId.set("");
        selectedTopicId.set("");
        const url = new URL(location.href);
        url.searchParams.delete("chat_id");
        url.searchParams.delete("topic_id");
        history.pushState(null, "", url);
    }

    function backToChats() {
        selectedTopicId.set(undefined);
        const url = new URL(location.href);
        url.searchParams.delete("topic_id");
        history.pushState(null, "", url);
    }
</script>

<main class="app-container">
    <div class="sidebar">
        <div class="sidebar-header">
            <span class="app-title">Messanger</span>
            <ChatCreationModel />
        </div>
        <ChatList />
    </div>
    
    {#if $selectedChatId}
    <div class="chat-panel">
        <div class="chat-header">
            {#if $selectedTopicId}
                <button class="back-btn" onclick={backToChats} aria-label="Back to topics">
                    <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
                        <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/>
                    </svg>
                </button>
            {:else}
                <button class="back-btn" onclick={clearChat} aria-label="Back to chats">
                    <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
                        <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
                    </svg>
                </button>
            {/if}
            <div class="chat-title-area">
                <h1 class="chat-title">{title}</h1>
                {#if $selectedTopicId}
                    <span class="topic-indicator">in topic</span>
                {/if}
            </div>
        </div>
        
        <div class="chat-content">
            {#if !$selectedTopicId}
                <TopicList chat_id={$selectedChatId} />
            {/if}
            <MessageList chat_id={$selectedChatId} topic_id={$selectedTopicId} />
        </div>
    </div>
    {:else}
    <div class="empty-state">
        <div class="empty-icon">
            <svg viewBox="0 0 24 24" width="64" height="64" fill="currentColor">
                <path d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm0 14H6l-2 2V4h16v12z"/>
            </svg>
        </div>
        <p>Select a chat to start messaging</p>
    </div>
    {/if}
</main>

<style>
    .app-container {
        display: flex;
        height: 100vh;
        background: #fff;
    }

    .sidebar {
        width: 350px;
        min-width: 350px;
        border-right: 1px solid #e6e8eb;
        display: flex;
        flex-direction: column;
        background: #fff;
    }

    .sidebar-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 16px;
        border-bottom: 1px solid #e6e8eb;
    }

    .app-title {
        font-weight: 600;
        font-size: 16px;
    }

    .menu-btn, .back-btn {
        width: 40px;
        height: 40px;
        border: none;
        border-radius: 50%;
        background: transparent;
        color: #8e8e93;
        display: flex;
        align-items: center;
        justify-content: center;
        transition: background 0.15s ease;
    }

    .menu-btn:hover, .back-btn:hover {
        background: #e8eaed;
    }

    .chat-panel {
        flex: 1;
        display: flex;
        flex-direction: column;
        background: #fff;
    }

    .chat-header {
        display: flex;
        align-items: center;
        gap: 12px;
        padding: 12px 16px;
        border-bottom: 1px solid #e6e8eb;
        background: #fff;
    }

    .chat-title-area {
        flex: 1;
        display: flex;
        flex-direction: column;
    }

    .chat-title {
        font-size: 16px;
        font-weight: 600;
        margin: 0;
    }

    .topic-indicator {
        font-size: 12px;
        color: #8e8e93;
    }

    .chat-content {
        flex: 1;
        display: flex;
        flex-direction: row;
        overflow: hidden;
    }

    .empty-state {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        color: #8e8e93;
        background: #f5f7f9;
    }

    .empty-icon {
        margin-bottom: 16px;
        opacity: 0.5;
    }

    @media (max-width: 768px) {
        .sidebar {
            width: 100%;
            min-width: 100%;
            position: absolute;
            z-index: 10;
        }

        .chat-panel {
            display: none;
        }
    }
</style>
