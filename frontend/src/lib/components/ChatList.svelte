<script lang="ts">
    import { fetchChats } from "$lib/api/chats";
    import { fetchMessages } from "$lib/api/messages";
    import { newMessageEvent } from "$lib/stores/messages";
    import type { ChatWithLastMessage, Message } from "$lib/types";

    const { onselect, current_chat_id, refreshKey, filterMode, categoryChatIds, onchatsload }: { onselect: (id: string) => void, current_chat_id?: string, refreshKey?: number, filterMode?: string, categoryChatIds?: string[], onchatsload?: (chats: ChatWithLastMessage[]) => void } = $props();

    let chats: ChatWithLastMessage[] = $state([]);
    let loadedChats: ChatWithLastMessage[] = $state([]);
    let fallbackMessages: Record<string, Message> = $state({});

    newMessageEvent.subscribe(msg => {
        if (!msg) return;
        fallbackMessages[msg.chat_id] = msg;
        const idx = loadedChats.findIndex(c => c.id === msg.chat_id);
        if (idx !== -1) {
            loadedChats[idx].last_message = msg;
            const chat = loadedChats.splice(idx, 1)[0];
            loadedChats = [chat, ...loadedChats];
        }
    });

    async function loadChats() {
        const resp = await fetchChats();
        if ("error" in resp) {
            console.error(resp.error);
            return
        }
        loadedChats = resp;
        onchatsload?.(resp);

        for (const chat of resp) {
            if (chat.last_message) continue;
            fetchMessages(chat.id).then(msgs => {
                if (Array.isArray(msgs) && msgs.length > 0) {
                    const lastMsg = msgs[msgs.length - 1];
                    if (lastMsg.content) {
                        fallbackMessages[chat.id] = lastMsg;
                    }
                }
            });
        }
    }

    function lastMsgFor(chat: ChatWithLastMessage): Message | null {
        if (chat.last_message) return chat.last_message;
        if (fallbackMessages[chat.id]) return fallbackMessages[chat.id];
        return null;
    }

    $effect(() => {
        refreshKey;
        loadChats();
    });

    $effect(() => {
        if (!loadedChats.length) {
            chats = [];
            return;
        }
        if (!filterMode || filterMode === "all") {
            chats = loadedChats;
        } else if (filterMode === "personal") {
            chats = loadedChats.filter(c => c.type === "private");
        } else if (filterMode === "groups") {
            chats = loadedChats.filter(c => c.type === "group" || c.type === "channel");
        } else if (filterMode === "channels") {
            chats = loadedChats.filter(c => c.type === "channel");
        } else if (filterMode === "category" && categoryChatIds) {
            chats = loadedChats.filter(c => categoryChatIds.includes(c.id));
        } else {
            chats = loadedChats;
        }
    });

    function selectChatEvent(chat_id: string) {
        return () => {
            onselect(chat_id);
            const url = new URL(location.href);
            url.searchParams.set("chat_id", chat_id);
            url.searchParams.delete("topic_id");
            history.pushState(null, "", url);
        }
    }

    function getInitials(title: string): string {
        return title.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase();
    }

    function formatTime(date: Date): string {
        const now = Date.now();
        const diff = now - new Date(date).getTime();
        const min = 60_000;
        const hour = 3_600_000;
        const day = 86_400_000;
        if (diff < min) return "now";
        if (diff < hour) return `${Math.floor(diff / min)}m`;
        if (diff < day) return `${Math.floor(diff / hour)}h`;
        return new Date(date).toLocaleDateString("en-US", { month: "short", day: "numeric" });
    }

    function previewText(msg: { content?: string } | null): string {
        if (!msg) return "No messages yet";
        if (!msg.content) return "Message";
        return msg.content.length > 100 ? msg.content.slice(0, 100) + "…" : msg.content;
    }
</script>

<div class="chat-list">
    {#each chats as chat (chat.id)}
        {@const msg = lastMsgFor(chat)}
        <button
            class="chat-item"
            class:selected={current_chat_id === chat.id}
            onclick={selectChatEvent(chat.id)}
        >
            <div class="avatar">
                {#if chat.avatar_url}
                    <img src={chat.avatar_url} alt="avatar" class="w-full h-full object-cover rounded-full" />
                {:else}
                    {getInitials(chat.title)}
                {/if}
            </div>
            <div class="chat-info">
                <div class="chat-header">
                    <span class="chat-title">{chat.title}</span>
                    {#if msg}
                        <span class="chat-time">{formatTime(msg.created_at)}</span>
                    {/if}
                </div>
                <div class="chat-preview">{previewText(msg)}</div>
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
