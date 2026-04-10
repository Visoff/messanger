<script lang="ts">
    import { API_URL } from "$lib/api/env";
    import ChatList from "$lib/components/ChatList.svelte";
    import ChatView from "$lib/components/ChatView.svelte";
    import { selectedChatId } from "$lib/stores/chat";
    import { onMount } from "svelte";

    onMount(() => {
        const token = localStorage.getItem("token");
        if (!token) {
            window.location.href = "/login";
        } else {
            console.log(token);
        }

        const url = new URL(window.location.href);
        const chat_id = url.searchParams.get("chat_id");
        if (chat_id) {
            selectedChatId.set(chat_id);
        }

        const stream = new EventSource(`${API_URL}/pubsub/sse`);
        stream.addEventListener("message", (e) => {
            console.log(e.data);
        });
    })

    function logout() {
        localStorage.removeItem("token");
        window.location.href = "/login";
    }
</script>

<main class="h-screen flex flex-row gap-5 items-start justify-start">
    <nav class="max-w-3xl flex flex-col gap-4 items-start">
        <button onclick={logout}>logout</button>
        <ChatList />
    </nav>
    {#if $selectedChatId}
        <ChatView chat_id={$selectedChatId} />
    {/if}
</main>
