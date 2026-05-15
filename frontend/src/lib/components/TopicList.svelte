<script lang="ts">
    import { createTopic, fetchTopics } from "$lib/api/topics";
    import { selectedTopicId } from "$lib/stores/chat";
    import type { Topic } from "$lib/types";

    const { chat_id }: { chat_id?: string } = $props();

    let topics: Topic[] = $state([]);

    $effect(() => {
        (async () => {
            if (!chat_id) {
                return
            }
            const resp = await fetchTopics(chat_id);
            if ("error" in resp) {
                console.error(resp.error);
                return
            }
            topics = resp;
        })();
    });

    function selectTopicEvent(topic_id: string) {
        return () => {
            selectedTopicId.set(topic_id);
            const url = new URL(location.href);
            url.searchParams.set("topic_id", topic_id);
            history.pushState(null, "", url);
        }
    }
</script>

{#if topics.length > 0}
<div class="topic-list">
    <h3 class="topic-header">Topics</h3>
    {#each topics as topic (topic.id)}
        <button
            class="topic-item"
            class:selected={$selectedTopicId === topic.id}
            onclick={selectTopicEvent(topic.id)}
        >
            <div class="topic-avatar">
                <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
                    <path d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2z"/>
                </svg>
            </div>
            <div class="topic-info">
                <span class="topic-title">{topic.title}</span>
            </div>
        </button>
    {/each}
</div>
{/if}

<style>
    .topic-list {
        display: flex;
        flex-direction: column;
        width: fit-content;
        border-right: 1px solid #e6e8eb;
    }

    .topic-header {
        font-size: 12px;
        font-weight: 600;
        color: #8e8e93;
        text-transform: uppercase;
        padding: 16px 16px 8px;
        margin: 0;
    }

    .topic-item {
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

    .topic-item:hover {
        background: #e8eaed;
    }

    .topic-item.selected {
        background: #e5f3fd;
    }

    .topic-avatar {
        width: 36px;
        height: 36px;
        border-radius: 50%;
        background: #2481d2;
        display: flex;
        align-items: center;
        justify-content: center;
        color: white;
        flex-shrink: 0;
    }

    .topic-info {
        flex: 1;
        min-width: 0;
    }

    .topic-title {
        font-weight: 500;
        font-size: 14px;
        color: #000000;
    }
</style>
