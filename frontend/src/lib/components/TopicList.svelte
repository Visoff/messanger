<script lang="ts">
    import { fetchTopics } from "$lib/api/topics";
    import type { Topic } from "$lib/types";

    const { chat_id, onselect, current_topic_id }: { onselect: (id: string) => void, chat_id?: string, current_topic_id?: string } = $props();

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

    function selectTopicEvent(topic: Topic) {
        return () => {
            if (topic.type === 'voice_topic') {
                const url = new URL(`${window.location.origin}/conference`);
                url.searchParams.set("room_id", topic.id);
                window.open(url.toString(), "_blank", "width=960,height=640,menubar=no,toolbar=no");
                return;
            }
            onselect(topic.id);
            const url2 = new URL(location.href);
            url2.searchParams.set("topic_id", topic.id);
            history.pushState(null, "", url2);
        }
    }
</script>

{#if topics.length > 0}
<div class="topic-list">
    <h3 class="topic-header">Topics</h3>
    {#each topics as topic (topic.id)}
        <button
            class="topic-item"
            class:selected={current_topic_id === topic.id}
            onclick={selectTopicEvent(topic)}
        >
            <div class="topic-avatar" class:voice-topic={topic.type === 'voice_topic'}>
                {#if topic.type === 'voice_topic'}
                    <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
                        <path d="M15 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm-9-2V7H4v3H1v2h3v3h2v-3h3v-2H6zm9 4c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
                    </svg>
                {:else}
                    <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
                        <path d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2z"/>
                    </svg>
                {/if}
            </div>
            <div class="topic-info">
                <span class="topic-title">{topic.title}</span>
                {#if topic.type === 'voice_topic'}
                    <span class="topic-badge">Voice</span>
                {/if}
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

    @media (max-width: 767px) {
        .topic-list {
            flex-direction: row;
            width: 100%;
            border-right: none;
            border-bottom: 1px solid #e6e8eb;
            overflow-x: auto;
            -webkit-overflow-scrolling: touch;
        }

        .topic-list::-webkit-scrollbar {
            display: none;
        }

        .topic-header {
            flex-shrink: 0;
        }
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

    .topic-badge {
        font-size: 10px;
        color: #43a047;
        font-weight: 600;
        text-transform: uppercase;
    }

    .voice-topic {
        background: #43a047;
    }
</style>
