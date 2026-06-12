<script lang="ts">
    import TopicList from "$lib/components/TopicList.svelte";
    import MessageList from "./MessageList.svelte";
    import { fetchChat } from "$lib/api/chats";
    import { fetchTopic } from "$lib/api/topics";
    import { extractFromSearchParams } from "$lib/index";
    import ChatViewDialog from "./ChatViewDialog.svelte";

    let messageListRef: MessageList;

    let chatModal: ChatViewDialog | undefined = $state();

    let {
        chat_id,
        onclose,
        onleave
    }: {
        chat_id: string,
        onclose: () => void,
        onleave?: () => void
    } = $props();

    let topic_id: string | undefined = $state(extractFromSearchParams("topic_id"));

    export function resetTopic() {
        topic_id = undefined;
        const url = new URL(location.href);
        url.searchParams.delete("topic_id");
        history.pushState(null, "", url);
    }

    let title: string = $state("...");

    $effect(() => {
        (async () => {
            if (!chat_id) {
                return;
            }
            if (!topic_id) {
                const resp = await fetchChat(chat_id);
                if ("error" in resp) {
                    console.error(resp.error);
                    return;
                }
                title = resp.title;
            } else {
                const resp = await fetchTopic(topic_id);
                if ("error" in resp) {
                    console.error(resp.error);
                    return;
                }
                title = resp.title;
            }
        })();
    });

    function backToChats() {
        topic_id = undefined
        const url = new URL(location.href);
        url.searchParams.delete("topic_id");
        history.pushState(null, "", url);
    }
</script>

<div class="chat-panel">
    <ChatViewDialog bind:this={chatModal} {chat_id} topic_id={topic_id} {onleave} />

    <div class="chat-header">
        {#if topic_id}
            <button
                class="back-btn"
                onclick={backToChats}
                aria-label="Back to topics"
            >
                <svg
                    viewBox="0 0 24 24"
                    width="24"
                    height="24"
                    fill="currentColor"
                >
                    <path
                        d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"
                    />
                </svg>
            </button>
        {:else}
            <button
                class="back-btn"
                onclick={onclose}
                aria-label="Back to chats"
            >
                <svg
                    viewBox="0 0 24 24"
                    width="24"
                    height="24"
                    fill="currentColor"
                >
                    <path
                        d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"
                    />
                </svg>
            </button>
        {/if}
        <button onclick={chatModal?.open} class="text-left chat-title-area">
            <h1 class="chat-title">{title}</h1>
            {#if topic_id}
                <span class="topic-indicator">in topic</span>
            {/if}
        </button>
    </div>

    <div class="chat-content">
        {#if !topic_id}
            <TopicList chat_id={chat_id} current_topic_id={topic_id} onselect={(id) => {topic_id = id}} />
        {/if}
        <MessageList
            bind:this={messageListRef}
            chat_id={chat_id}
            topic_id={topic_id}
        />
    </div>
</div>

<style>
    .back-btn {
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

    .back-btn:hover {
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

    @media (max-width: 767px) {
        .chat-content {
            flex-direction: column;
        }
    }
</style>
