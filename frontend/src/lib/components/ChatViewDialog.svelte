<script lang="ts">
    import Dialog from "./Dialog.svelte";
    import { fetchChat, InviteUserToChat } from "$lib/api/chats";
    import { fetchTopic, createTopic } from "$lib/api/topics";
    import { resolveUsername } from "$lib/api/users";
    import type { Chat, Topic } from "$lib/types";

    let {
        chat_id,
        topic_id
    }: {
        chat_id: string,
        topic_id?: string
    } = $props();

    let dialog: Dialog;
    let chatData: Chat | null = $state(null);
    let topicData: Topic | null = $state(null);
    let loading = $state(false);
    let error = $state("");

    export function open() {
        dialog.open();
        loadData();
    }

    export function close() {
        dialog.close();
    }

    async function loadData() {
        loading = true;
        error = "";
        chatData = null;
        topicData = null;

        try {
            if (!topic_id) {
                const resp = await fetchChat(chat_id);
                if ("error" in resp) {
                    error = resp.error;
                } else {
                    chatData = resp;
                }
            } else {
                const resp = await fetchTopic(topic_id);
                if ("error" in resp) {
                    error = resp.error;
                } else {
                    topicData = resp;
                }
            }
        } catch {
            error = "Failed to load data";
        } finally {
            loading = false;
        }
    }

    function formatDate(date: Date | string): string {
        if (!date) return "—";
        return new Date(date).toLocaleDateString(undefined, {
            year: "numeric",
            month: "short",
            day: "numeric",
            hour: "2-digit",
            minute: "2-digit"
        });
    }

    function typeLabel(type: string): string {
        switch (type) {
            case "private": return "Private Chat";
            case "group": return "Group Chat";
            case "channel": return "Channel";
            case "text_topic": return "Text Topic";
            case "voice_topic": return "Voice Topic";
            default: return type;
        }
    }

    async function handleCreateTopic() {
        const title = prompt("Topic name:");
        if (!title) return;
        const resp = await createTopic(chat_id, title, "text_topic");
        if ("error" in resp) {
            alert(resp.error ?? "Failed to create topic");
        } else {
            alert(`Topic "${resp.title}" created`);
        }
    }

    async function handleInviteUser() {
        const username = prompt("Username to invite:");
        if (!username) return;
        const user = await resolveUsername(username);
        if ("error" in user) {
            alert(user.error ?? "User not found");
            return;
        }
        const resp = await InviteUserToChat(chat_id, user.id);
        if ("error" in resp) {
            alert(resp.error ?? "Failed to invite user");
        } else {
            alert(`User "${user.username}" invited`);
        }
    }
</script>

<Dialog bind:this={dialog}>
    <form
        class="bg-gray-200 border border-gray-600 rounded-lg px-4 py-2 flex flex-col gap-2"
        onsubmit={(e) => e.preventDefault()}
    >
        {#if loading}
            <div class="flex items-center justify-center py-6">
                <p class="text-gray-500">Loading…</p>
            </div>
        {:else if error}
            <div class="text-red-500 py-3 text-center">{error}</div>
            <button type="button" onclick={close} class="self-end px-4 py-1.5 bg-gray-300 hover:bg-gray-400 rounded-lg text-sm transition-colors">Close</button>
        {:else if chatData}
            <div class="flex items-center justify-between">
                <h2 class="text-xl font-bold">{chatData.title}</h2>
                <span class="text-xs px-2 py-0.5 rounded-full bg-blue-100 text-blue-700">{typeLabel(chatData.type)}</span>
            </div>
            <div class="text-sm text-gray-600 space-y-1">
                <div class="flex justify-between">
                    <span class="text-gray-500">ID</span>
                    <span class="font-mono text-xs">{chatData.id}</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-gray-500">Created</span>
                    <span>{formatDate(chatData.created_at)}</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-gray-500">Updated</span>
                    <span>{formatDate(chatData.updated_at)}</span>
                </div>
            </div>
            <div class="flex justify-end gap-2">
                <button type="button" onclick={handleInviteUser} class="px-4 py-1.5 bg-blue-300 hover:bg-blue-400 rounded-lg text-sm transition-colors">Invite</button>
                <button type="button" onclick={handleCreateTopic} class="px-4 py-1.5 bg-green-300 hover:bg-green-400 rounded-lg text-sm transition-colors">New Topic</button>
                <button type="button" onclick={close} class="px-4 py-1.5 bg-gray-300 hover:bg-gray-400 rounded-lg text-sm transition-colors">Close</button>
            </div>
        {:else if topicData}
            <div class="flex items-center justify-between">
                <h2 class="text-xl font-bold">{topicData.title}</h2>
                <span class="text-xs px-2 py-0.5 rounded-full bg-green-100 text-green-700">{typeLabel(topicData.type)}</span>
            </div>
            <div class="text-sm text-gray-600 space-y-1">
                <div class="flex justify-between">
                    <span class="text-gray-500">ID</span>
                    <span class="font-mono text-xs">{topicData.id}</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-gray-500">Chat ID</span>
                    <span class="font-mono text-xs">{topicData.chat_id}</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-gray-500">Created</span>
                    <span>{formatDate(topicData.created_at)}</span>
                </div>
                <div class="flex justify-between">
                    <span class="text-gray-500">Updated</span>
                    <span>{formatDate(topicData.updated_at)}</span>
                </div>
            </div>
            <div class="flex justify-end">
                <button type="button" onclick={close} class="px-4 py-1.5 bg-gray-300 hover:bg-gray-400 rounded-lg text-sm transition-colors">Close</button>
            </div>
        {/if}
    </form>
</Dialog>
