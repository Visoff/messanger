<script lang="ts">
    import { fetchChats } from "$lib/api/chats";
    import { selectedChatId, selectedTopicId } from "$lib/stores/chat";
    import type { Chat } from "$lib/types";
    import { onMount } from "svelte";

    let chats: Chat[] = $state([]);

    onMount(async () => {
        const resp = await fetchChats();
        if ("error" in resp) {
            console.error(resp.error);
            return
        }
        chats = resp;
    });

    function selectChatEvent(chat_id: string) {
        return () => {
            selectedChatId.set(chat_id);
            selectedTopicId.set(undefined);
            const url = new URL(location.href);
            url.searchParams.set("chat_id", chat_id);
            url.searchParams.delete("topic_id");
            history.pushState(null, "", url);
        }
    }
</script>

<ul class="flex flex-col items-start justify-start">
{#each chats as chat (chat.id)}
    <li><button onclick={selectChatEvent(chat.id)}>{chat.title}</button></li>
{/each}
</ul>
