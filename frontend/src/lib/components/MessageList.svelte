<script lang="ts">
    import { API_URL } from "$lib/api/env";
    import { fetchMessages, sendMessage } from "$lib/api/messages";
    import { fetchTopics } from "$lib/api/topics";
    import { user } from "$lib/stores/user";
    import type { Message, Topic } from "$lib/types";

    const { chat_id, topic_id }: { chat_id: string, topic_id?: string } = $props();

    export function addOptimisticMessage(message: Message) {
        messages = [...messages, message];
        scrollToBottom();
    }

    let chat_topics: Topic[] = $state([]);
    let messages: Message[] = $state([]);
    let inputValue: string = $state("");
    let messagesContainer: HTMLDivElement;
    let userCache: Record<string, string> = $state({});
    let replyToMessage: Message | null = $state(null);

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

            const uniqueIds = [...new Set(resp.map(m => m.sender_id).filter(Boolean))];
            for (const uid of uniqueIds) {
                if (!userCache[uid]) {
                    try {
                        const resp2 = await fetch(`${API_URL}/users/id/${uid}`);
                        const u = await resp2.json();
                        if (!("error" in u)) {
                            userCache[uid] = u.username;
                        }
                    } catch {}
                }
            }

            scrollToBottom();
        })();
    });

    function getSenderName(sender_id: string): string {
        if (sender_id === $user?.id) return "You";
        return userCache[sender_id] || sender_id.slice(0, 8);
    }

    function getRepliedMessage(message: Message): Message | undefined {
        if (!message.reply_message_id) return undefined;
        return messages.find(m => m.id === message.reply_message_id);
    }

    function notify(title: string, body?: string) {
        Notification.requestPermission().then((permission) => {
            if (permission === "granted") {
                new Notification(title, { body });
            }
        });
    }

    onMount(() => {
        const token = localStorage.getItem("token");

        const stream = new EventSource(`${API_URL}/pubsub/sse?token=${token}`);
        stream.addEventListener("message", (e) => {
            const data = JSON.parse(e.data) as Message;
            if (
                data.chat_id == chat_id &&
                (!topic_id || data.topic_id == topic_id)
            ) {
                addOptimisticMessage(data);
            } else {
                let content = data.content || "";
                if (content.length > 50) {
                    content = content.slice(0, 50) + "...";
                }
                notify("new message", content);
            }
        });

    });

    function scrollToBottom() {
        setTimeout(() => {
            if (messagesContainer) {
                messagesContainer.scrollTop = messagesContainer.scrollHeight;
            }
        }, 10);
    }

    function scrollToMessage(messageId: string) {
        const el = document.getElementById(`msg-${messageId}`);
        if (el) {
            el.scrollIntoView({ behavior: "smooth", block: "center" });
        }
    }

    function startReply(message: Message) {
        replyToMessage = message;
    }

    function cancelReply() {
        replyToMessage = null;
    }

    async function sendMessageEvent(e: SubmitEvent) {
        e.preventDefault();
        if (!inputValue.trim()) return;
        
        const resp = await sendMessage(chat_id, topic_id, inputValue, replyToMessage?.id);
        if ("error" in resp) {
            console.error(resp.error);
            return
        };
        inputValue = "";
        replyToMessage = null;
        addOptimisticMessage(resp);
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
            <div class="message-row" class:sent={message.sender_id === $user?.id}>
                <div class="message-bubble" class:received={message.sender_id !== $user?.id} id="msg-{message.id}">
                    {#if message.reply_message_id}
                        {@const replied = getRepliedMessage(message)}
                        {#if replied}
                            <button type="button" class="reply-header" onclick={(e) => { e.stopPropagation(); scrollToMessage(replied.id); }}>
                                <div class="reply-header-sender">{getSenderName(replied.sender_id)}</div>
                                <div class="reply-header-text">{replied.content || "No content"}</div>
                            </button>
                        {/if}
                    {/if}
                    {#if message.sender_id !== $user?.id}
                        <div class="message-sender">{getSenderName(message.sender_id)}</div>
                    {/if}
                    <button type="button" class="message-content-btn" onclick={() => startReply(message)}>
                        <div class="message-content">{message.content}</div>
                    </button>
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
    {#if replyToMessage}
        <div class="reply-indicator">
            <div class="reply-indicator-bar"></div>
            <div class="reply-indicator-content">
                <span class="reply-indicator-sender">{getSenderName(replyToMessage.sender_id)}</span>
                <span class="reply-indicator-text">{replyToMessage.content || "No content"}</span>
            </div>
            <button type="button" class="reply-indicator-close" aria-label="Cancel reply" onclick={cancelReply}>
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                    <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
                </svg>
            </button>
        </div>
    {/if}
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
        flex-direction: column;
        align-items: flex-start;
        padding: 0 60px 0 0;
    }

    .message-row.sent {
        align-items: flex-end;
        padding: 0 0 0 60px;
    }

    .message-bubble {
        display: flex;
        flex-direction: column;
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

    .message-bubble:hover {
        filter: brightness(0.97);
    }

    .message-content {
        font-size: 15px;
        line-height: 1.4;
        word-wrap: break-word;
    }

    .message-content-btn {
        display: block;
        text-align: inherit;
        border: none;
        background: none;
        padding: 0;
        cursor: pointer;
    }

    .message-sender {
        font-size: 11px;
        font-weight: 600;
        color: #2481d2;
        margin-bottom: 2px;
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

    .reply-header {
        display: flex;
        flex-direction: column;
        margin: -10px -14px 6px -14px;
        padding: 8px 14px;
        border: none;
        background: #c8e6f8;
        border-radius: 18px 18px 0 0;
        border-bottom: 1px solid #b0d4e8;
        cursor: pointer;
        text-align: left;
        border-top-left-radius: 4px;
    }

    .message-bubble.received .reply-header {
        background: #e2e4e7;
        border-bottom: 1px solid #d0d2d5;
        margin: -11px -15px 6px -15px;
        border-top-left-radius: 4px;
    }

    .reply-header:hover {
        filter: brightness(0.95);
    }

    .reply-header-sender {
        font-size: 11px;
        font-weight: 600;
        color: #43a047;
    }

    .reply-header-text {
        font-size: 12px;
        color: #555;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
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

    .reply-indicator {
        display: flex;
        align-items: center;
        gap: 10px;
        padding: 8px 16px 0 16px;
        background: #fff;
    }

    .reply-indicator-bar {
        width: 3px;
        height: 36px;
        background: #43a047;
        border-radius: 2px;
        flex-shrink: 0;
    }

    .reply-indicator-content {
        flex: 1;
        display: flex;
        flex-direction: column;
        overflow: hidden;
    }

    .reply-indicator-sender {
        font-size: 12px;
        font-weight: 600;
        color: #43a047;
    }

    .reply-indicator-text {
        font-size: 13px;
        color: #8e8e93;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .reply-indicator-close {
        border: none;
        background: transparent;
        padding: 4px;
        cursor: pointer;
        color: #8e8e93;
        flex-shrink: 0;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
    }

    .reply-indicator-close:hover {
        background: #e8eaed;
        color: #000;
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
