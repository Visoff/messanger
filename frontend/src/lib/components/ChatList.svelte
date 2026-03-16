<script lang="ts">
    import { fetchChats, createChat } from "$lib/api/chats";
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
</script>

<ul class="flex flex-col items-start justify-start">
{#each chats as chat (chat.id)}
    <li>{chat.title}</li>
{/each}
    <li onclick={createChatEvent}>new chat</li>
</ul>
