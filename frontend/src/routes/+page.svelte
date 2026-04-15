<script lang="ts">
    import { createChat, fetchChat } from "$lib/api/chats";
    import { fetchTopic } from "$lib/api/topics";
    import { API_URL } from "$lib/api/env";
    import ChatList from "$lib/components/ChatList.svelte";
    import MessageList from "$lib/components/MessageList.svelte";
    import TopicList from "$lib/components/TopicList.svelte";
    import { selectedChatId, selectedTopicId } from "$lib/stores/chat";
    import { onMount } from "svelte";

    let title: string = $state("");

    $effect(() => {
        (async () => {
            if (!$selectedChatId) {
                return
            }
            if (!$selectedTopicId) {
                const resp = await fetchChat($selectedChatId);
                if ("error" in resp) {
                    console.error(resp.error);
                    return
                }
                title = resp.title;
            } else {
                const resp = await fetchTopic($selectedTopicId);
                if ("error" in resp) {
                    console.error(resp.error);
                    return
                }
                title = resp.title;
            }
        })();
    })

    onMount(() => {
        const token = localStorage.getItem("token");
        if (!token) {
            window.location.href = "/login";
        } else {
            console.log(token);
        }

        const url = new URL(window.location.href);
        const chat_id = url.searchParams.get("chat_id");
        if (chat_id) {
            selectedChatId.set(chat_id);
        }

        const topic_id = url.searchParams.get("topic_id");
        if (topic_id) {
            selectedTopicId.set(topic_id);
        }

        const stream = new EventSource(`${API_URL}/pubsub/sse`);
        stream.addEventListener("message", (e) => {
            console.log(e.data);
        });
    })

    async function createChatEvent() {
        const title = prompt("Chat title");
        if (title) {
            const resp = await createChat(title);
            if ("error" in resp) {
                console.error(resp.error);
                return;
            }
            selectedChatId.set(resp.id);
        }
    }
</script>

<main class="h-screen flex flex-row gap-5 items-start justify-start">
    <nav class="max-w-3xl flex flex-row">
        <div class="overflow-hidden flex flex-col gap-4 items-start">
            <h1 class="text-2xl font-bold">Chats</h1>
            <ChatList />
            <button class="text-2xl font-bold" type="button" onclick={createChatEvent()}>+</button>
        </div>
        <TopicList chat_id={$selectedChatId} />
    </nav>
    <div class="h-full w-0.5 bg-gray-300"></div>
    <div class="h-full flex-1 flex flex-col">
        <h1 class="text-2xl font-bold">{title}</h1>
        {#if $selectedChatId}
            <MessageList chat_id={$selectedChatId} topic_id={$selectedTopicId} />
        {/if}
    </div>
</main>
