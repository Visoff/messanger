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

    function getInitials(title: string): string {
        return title.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase();
    }
</script>

<div class="chat-list">
    {#each chats as chat (chat.id)}
        <button
            class="chat-item"
            class:selected={$selectedChatId === chat.id}
            onclick={selectChatEvent(chat.id)}
        >
            <div class="avatar">{getInitials(chat.title)}</div>
            <div class="chat-info">
                <div class="chat-header">
                    <span class="chat-title">{chat.title}</span>
                    <span class="chat-time">now</span>
                </div>
                <div class="chat-preview">Tap to view messages</div>
            </div>
        </button>
    {/each}
</div>

<style>
    .chat-list {
        display: flex;
        flex-direction: column;
        width: 100%;
    }

    .chat-item {
        display: flex;
        align-items: center;
        gap: 12px;
        padding: 12px 16px;
        border: none;
        background: transparent;
        text-align: left;
        transition: background 0.15s ease;
        width: 100%;
    }

    .chat-item:hover {
        background: #e8eaed;
    }

    .chat-item.selected {
        background: #e5f3fd;
    }

    .avatar {
        width: 48px;
        height: 48px;
        border-radius: 50%;
        background: linear-gradient(135deg, #2481d2, #0d7ad6);
        display: flex;
        align-items: center;
        justify-content: center;
        color: white;
        font-weight: 600;
        font-size: 14px;
        flex-shrink: 0;
    }

    .chat-info {
        flex: 1;
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .chat-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .chat-title {
        font-weight: 500;
        font-size: 14px;
        color: #000000;
    }

    .chat-time {
        font-size: 12px;
        color: #8e8e93;
    }

    .chat-preview {
        font-size: 13px;
        color: #8e8e93;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
</style>