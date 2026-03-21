<script lang="ts">
    import { fetchMessages } from "$lib/api/messages";
    import type { Message } from "$lib/types";

    const { chat_id, topic_id }: { chat_id: string, topic_id?: string } = $props();

    let messages: Message[] = $state([]);

    $effect(() => {
        (async () => {
            const resp = await fetchMessages(chat_id, topic_id);
            if ("error" in resp) {
                console.error(resp.error);
                return;
            }
            messages = resp;
        })();
    })

    function sendMessageEvent(e: SubmitEvent) {
        e.preventDefault();
        const data = {
            chat_id,
            topic_id,
            content: (e.target as HTMLFormElement).message.value,
        }
        console.log(data);
    }
</script>

<div class="h-full flex flex-col">
    <div class="flex-1">
        {#each messages as message(message.id)}
            <div>{message.content}</div>
        {/each}
    </div>
    <form onsubmit={sendMessageEvent}>
        <input placeholder="message" name="message" />
        <button type="submit">send</button>
    </form>
</div>
