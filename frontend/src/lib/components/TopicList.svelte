<script lang="ts">
    import { fetchTopics } from "$lib/api/topics";
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

<ul class="flex flex-col items-start justify-start">
{#each topics as topic (topic.id)}
    <li><button onclick={selectTopicEvent(topic.id)}>{topic.title}</button></li>
{/each}
</ul>
