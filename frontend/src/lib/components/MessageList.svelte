<script lang="ts">
    import { fetchMessages, sendMessage } from "$lib/api/messages";
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
        (async () => {
            const resp = await sendMessage(chat_id, topic_id, (e.target as HTMLFormElement).message.value);
            if ("error" in resp) {
                console.error(resp.error);
                return
            };
            (e.target as HTMLFormElement).message.value = "";
            messages = [...messages, resp];
        })();
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
