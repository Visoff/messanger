<script lang="ts">
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
