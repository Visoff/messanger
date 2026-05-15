<script lang="ts">
    import { fetchMessages, sendMessage } from "$lib/api/messages";
    import { fetchTopics } from "$lib/api/topics";
    import type { Message, Topic } from "$lib/types";

    const { chat_id, topic_id }: { chat_id: string, topic_id?: string } = $props();

    let chat_topics: Topic[] = $state([]);
    let messages: Message[] = $state([]);
    let currentUserId: string = $state("");
    let inputValue: string = $state("");
    let messagesContainer: HTMLDivElement;

    $effect(() => {
        (async () => {
            const resp = await fetchMessages(chat_id, topic_id);
            if ("error" in resp) {
                console.error(resp.error);
                return;
            }
            messages = resp;
            const resp1 = await fetchTopics(chat_id);
            if ("error" in resp1) {
                console.error(resp1.error);
                return;
            }
            chat_topics = resp1
            scrollToBottom();
        })();
    });

    onMount(() => {
        const token = localStorage.getItem("token");
        if (token) {
            try {
                const payload = JSON.parse(atob(token.split('.')[1]));
                currentUserId = payload.sub || payload.user_id || "";
            } catch (e) {
                console.error("Failed to parse token");
            }
        }
    });

    function scrollToBottom() {
        setTimeout(() => {
            if (messagesContainer) {
                messagesContainer.scrollTop = messagesContainer.scrollHeight;
            }
        }, 10);
    }

    async function sendMessageEvent(e: SubmitEvent) {
        e.preventDefault();
        if (!inputValue.trim()) return;
        
        const resp = await sendMessage(chat_id, topic_id, inputValue);
        if ("error" in resp) {
            console.error(resp.error);
            return
        };
        inputValue = "";
        messages = [...messages, resp];
        scrollToBottom();
    }

    function formatTime(date: Date): string {
        const d = new Date(date);
        return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    }

    import { onMount } from "svelte";
</script>

<div class="message-container">
    <div class="messages-wrapper" bind:this={messagesContainer}>
        {#each messages as message (message.id)}
            <div class="message-row" class:sent={message.sender_id === currentUserId}>
                <div class="message-bubble" class:received={message.sender_id !== currentUserId}>
                    <div class="message-content">{message.content}</div>
                    <div class="message-meta">
                    {#if topic_id == undefined && message.topic_id != undefined}
                        <span class="text-xs text-blue-500">{chat_topics.find(t => t.id === message.topic_id)?.title}</span>
                        <div class="w-1.5"></div>
                    {/if}
                        <span class="message-time">{formatTime(message.created_at)}</span>
                    </div>
                </div>
            </div>
        {/each}
        {#if messages.length === 0}
            <div class="empty-messages">
                <p>No messages yet</p>
                <p class="empty-hint">Send a message to start the conversation</p>
            </div>
        {/if}
    </div>
    <form class="message-input-wrapper" onsubmit={sendMessageEvent}>
        <input 
            class="message-input" 
            placeholder="Message" 
            name="message"
            bind:value={inputValue}
        />
        <button class="send-button" type="submit" aria-label="Send message">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
                <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
            </svg>
        </button>
    </form>
</div>

<style>
    .message-container {
        display: flex;
        flex-direction: column;
        flex: 1;
        background: #fff;
    }

    .messages-wrapper {
        flex: 1;
        overflow-y: auto;
        padding: 16px;
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .message-row {
        display: flex;
        justify-content: flex-start;
        padding: 0 60px 0 0;
    }

    .message-row.sent {
        justify-content: flex-end;
        padding: 0 0 0 60px;
    }

    .message-bubble {
        max-width: 70%;
        padding: 10px 14px;
        border-radius: 18px;
        background: #e5f3fd;
        border-top-left-radius: 4px;
    }

    .message-bubble.received {
        background: #f5f7f9;
        border: 1px solid #e6e8eb;
    }

    .message-content {
        font-size: 15px;
        line-height: 1.4;
        word-wrap: break-word;
    }

    .message-meta {
        display: flex;
        justify-content: flex-end;
        margin-top: 4px;
    }

    .message-time {
        font-size: 11px;
        color: #8e8e93;
    }

    .empty-messages {
        flex: 1;
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        color: #8e8e93;
    }

    .empty-hint {
        font-size: 13px;
        margin-top: 4px;
    }

    .message-input-wrapper {
        display: flex;
        align-items: center;
        gap: 10px;
        padding: 12px 16px;
        background: #fff;
        border-top: 1px solid #e6e8eb;
    }

    .message-input {
        flex: 1;
        padding: 12px 16px;
        border: none;
        border-radius: 24px;
        background: #f5f7f9;
        font-size: 15px;
        outline: none;
    }

    .message-input:focus {
        background: #e8eaed;
    }

    .send-button {
        width: 44px;
        height: 44px;
        border: none;
        border-radius: 50%;
        background: #2481d2;
        color: white;
        display: flex;
        align-items: center;
        justify-content: center;
        transition: background 0.15s ease;
    }

    .send-button:hover {
        background: #1c6ea8;
    }
</style>
