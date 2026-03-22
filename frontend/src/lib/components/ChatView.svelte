<script lang="ts">
    import { fetchChat } from "$lib/api/chats";
    import { createTopic, fetchTopics } from "$lib/api/topics";
    import { selectedTopicId } from "$lib/stores/chat";
    import type { Chat, Topic } from "$lib/types";
    import MessageList from "./MessageList.svelte";

    const { chat_id }: { chat_id: string } = $props();

    let chat: Chat | undefined = $state(undefined);
    let topics: Topic[] = $state([]);

    $effect(() => {
        (async () => {
            const resp1 = await fetchChat(chat_id);
            if ("error" in resp1) {
                console.error(resp1.error);
                return;
            }
            chat = resp1;

            const resp2 = await fetchTopics(chat_id);
            if ("error" in resp2) {
                console.error(resp2.error);
                return;
            }
            topics = resp2;
        })();
    })

    function selectTopic(topic_id: string) {
        return () => {
            $selectedTopicId = topic_id;
        }
    }

    async function createTopicEvent() {
        const title = prompt("topic title");
        if (!title) {
            return;
        }
        const resp = await createTopic(chat_id, title, 'text_topic');
        if ("error" in resp) {
            console.error(resp.error);
            return;
        }
        topics = [...topics, resp];
    }
</script>

{#if chat}
    <div class="h-full flex-1 flex flex-col">
        <div class="flex flex-row justify-between">
            <h2 class="text-2xl font-bold">{chat.title}</h2>
            <button class="text-sm" onclick={createTopicEvent}>New topic</button>
        </div>
        <div class="flex flex-row flex-1 gap-4">
            {#if topics.length != 0}
            <nav class="flex flex-col">
                {#each topics as topic(topic.id)}
                    <button onclick={selectTopic(topic.id)}>{topic.title}</button>
                {/each}
            </nav>
            {/if}
            {#if topics.length == 0 || $selectedTopicId}
                <MessageList chat_id={chat_id} topic_id={$selectedTopicId} />
            {/if}
        </div>
    </div>
{/if}
