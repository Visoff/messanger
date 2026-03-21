<script lang="ts">
    import { fetchChats, createChat } from "$lib/api/chats";
    import { selectedChatId } from "$lib/stores/chat";
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

    function createChatEvent() {
        const title = prompt("chat title");
        if (!title) {
            return;
        }
        createChat(title).then((resp) => {
            if ("error" in resp) {
                console.error(resp.error);
                return;
            }
            chats = [...chats, resp];
        })
    }

    function selectChatEvent(chat_id: string) {
        return () => {
            $selectedChatId = chat_id
        }
    }
</script>

<ul class="flex flex-col items-start justify-start">
{#each chats as chat (chat.id)}
    <li><button onclick={selectChatEvent(chat.id)}>{chat.title}</button></li>
{/each}
    <li><button onclick={createChatEvent}>new chat</button></li>
</ul>
