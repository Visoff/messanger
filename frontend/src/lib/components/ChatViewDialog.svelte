<script lang="ts">
    import Dialog from "./Dialog.svelte";
    import PrivateChatInfo from "./PrivateChatInfo.svelte";
    import { fetchChat, fetchChatMembers, InviteUserToChat, updateChat, leaveChat, muteChat, createInvitation, uploadChatAvatar } from "$lib/api/chats";
    import { fetchTopic, createTopic } from "$lib/api/topics";
    import { resolveUsername } from "$lib/api/users";
    import type { Chat, Topic, TopicType, User } from "$lib/types";
    import { API_URL } from "$lib/api/env";
    import { user } from "$lib/stores/user";

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
    let members: User[] = $state([]);
    let loading = $state(false);
    let error = $state("");

    let showInviteInput = $state(false);
    let inviteUsername = $state("");
    let showNewTopic = $state(false);
    let newTopicName = $state("");
    let newTopicType: TopicType | null = $state(null);
    let editingTitle = $state(false);
    let editTitleValue = $state("");
    let invitationLink = $state("");
    let uploadingAvatar = $state(false);
    let avatarFileInput: HTMLInputElement;

    let otherUser = $derived(
        chatData?.type === "private"
            ? members.find(m => m.id !== $user?.id) ?? null
            : null
    );

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
        members = [];
        invitationLink = "";
        showInviteInput = false;
        showNewTopic = false;
        newTopicType = null;
        editingTitle = false;

        try {
            if (!topic_id) {
                const [chatResp, membersResp] = await Promise.all([
                    fetchChat(chat_id),
                    fetchChatMembers(chat_id)
                ]);
                if ("error" in chatResp) {
                    error = chatResp.error;
                } else {
                    chatData = chatResp;
                }
                if (!("error" in membersResp)) {
                    members = membersResp;
                } else {
                    console.error(membersResp.error);
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
        if (!newTopicName.trim() || !newTopicType) return;
        const resp = await createTopic(chat_id, newTopicName, newTopicType);
        if ("error" in resp) {
            error = resp.error ?? "Failed to create topic";
        } else {
            newTopicName = "";
            newTopicType = null;
            showNewTopic = false;
        }
    }

    function selectTopicType(type: TopicType) {
        newTopicType = type;
    }

    function cancelNewTopic() {
        newTopicName = "";
        newTopicType = null;
        showNewTopic = false;
    }

    async function handleInviteUser() {
        if (!inviteUsername.trim()) return;
        const respUser = await resolveUsername(inviteUsername);
        if ("error" in respUser) {
            error = respUser.error ?? "User not found";
            return;
        }
        const resp = await InviteUserToChat(chat_id, respUser.id);
        if ("error" in resp) {
            error = resp.error ?? "Failed to invite user";
        } else {
            inviteUsername = "";
            showInviteInput = false;
            loadData();
        }
    }

    async function handleRename() {
        if (!editTitleValue.trim() || !chatData) return;
        const resp = await updateChat(chat_id, editTitleValue);
        if ("error" in resp) {
            error = resp.error ?? "Failed to rename chat";
        } else {
            chatData = resp;
            editingTitle = false;
        }
    }

    async function handleLeave() {
        const resp = await leaveChat(chat_id);
        if ("error" in resp) {
            error = resp.error ?? "Failed to leave chat";
        } else {
            close();
            window.location.reload();
        }
    }

    async function handleCreateInvitation() {
        const resp = await createInvitation(chat_id);
        if ("error" in resp) {
            error = resp.error ?? "Failed to create invitation";
        } else {
            invitationLink = `${API_URL}/invitation/${resp.id}`;
        }
    }

    async function handleMute() {
        const isMuted = !!(chatData?.metadata && JSON.parse(chatData.metadata as string)?.muted);
        const resp = await muteChat(chat_id, !isMuted);
        if ("error" in resp) {
            error = resp.error ?? "Failed to toggle mute";
        } else {
            chatData = resp;
        }
    }

    function isMuted(): boolean {
        if (!chatData?.metadata) return false;
        try {
            const meta = JSON.parse(chatData.metadata as string);
            return !!meta.muted;
        } catch { return false; }
    }

    function getInitials(name: string): string {
        return name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase();
    }

    async function handleChatAvatarUpload(e: Event) {
        const target = e.target as HTMLInputElement;
        const file = target.files?.[0];
        if (!file || !chatData) return;

        uploadingAvatar = true;
        error = "";
        const resp = await uploadChatAvatar(chat_id, file);
        if ("error" in resp) {
            error = resp.error ?? "Failed to upload avatar";
        } else {
            chatData = resp;
        }
        uploadingAvatar = false;
        target.value = "";
    }
</script>

<Dialog bind:this={dialog}>
    <form
        class="bg-gray-200 border border-gray-600 rounded-lg px-4 py-2 flex flex-col gap-2 min-w-72 max-w-md"
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
            {#if chatData.type === "private" && otherUser}
                <PrivateChatInfo {otherUser} />
            {:else}
                <div class="flex items-center gap-3">
                    <button onclick={() => avatarFileInput?.click()} class="relative w-12 h-12 rounded-full bg-gradient-to-br from-blue-400 to-blue-600 flex items-center justify-center text-white font-bold text-base overflow-hidden flex-shrink-0 hover:opacity-80 transition-opacity group" title="Change chat avatar">
                        {#if chatData.avatar_url}
                            <img src={chatData.avatar_url} alt="avatar" class="w-full h-full object-cover" />
                        {:else}
                            {getInitials(chatData.title)}
                        {/if}
                        <div class="absolute inset-0 bg-black/40 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                            <svg viewBox="0 0 24 24" width="16" height="16" fill="white"><path d="M17 5h-1.5l-1-1h-5l-1 1H7v2h10V5zM7 7v10a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2V7H7z"/></svg>
                        </div>
                    </button>
                    <input bind:this={avatarFileInput} type="file" accept="image/*" class="hidden" onchange={handleChatAvatarUpload} />
                    <div class="flex-1">
                        {#if editingTitle}
                            <div class="flex gap-1">
                                <input class="border rounded px-2 py-1 text-sm flex-1" bind:value={editTitleValue} />
                                <button type="button" onclick={handleRename} class="px-2 py-1 bg-blue-500 text-white rounded text-xs">Save</button>
                                <button type="button" onclick={() => editingTitle = false} class="px-2 py-1 bg-gray-400 text-white rounded text-xs">Cancel</button>
                            </div>
                        {:else}
                            <button type="button" onclick={() => { editTitleValue = chatData.title; editingTitle = true; }} class="text-xl font-bold hover:text-blue-600 transition-colors text-left">{chatData.title} ✎</button>
                        {/if}
                        <span class="text-xs px-2 py-0.5 rounded-full bg-blue-100 text-blue-700">{typeLabel(chatData.type)}</span>
                    </div>
                </div>
                {#if uploadingAvatar}
                    <div class="text-xs text-gray-500 text-center">Uploading avatar...</div>
                {/if}
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

                {#if members.length > 0}
                    <div class="border-t border-gray-400 pt-2">
                        <h3 class="text-sm font-semibold text-gray-700 mb-1">Members ({members.length})</h3>
                        <div class="max-h-32 overflow-y-auto space-y-1">
                            {#each members as member (member.id)}
                                <div class="flex items-center gap-2 text-xs">
                                    <div class="w-6 h-6 rounded-full bg-gradient-to-br from-blue-400 to-blue-600 flex items-center justify-center text-white font-bold text-[10px] flex-shrink-0">
                                        {getInitials(member.username)}
                                    </div>
                                    <span class="truncate">{member.username}</span>
                                    {#if member.id === $user?.id}
                                        <span class="text-gray-400">(you)</span>
                                    {/if}
                                </div>
                            {/each}
                        </div>
                    </div>
                {/if}
            {/if}

            {#if invitationLink}
                <div class="border-t border-gray-400 pt-2">
                    <h3 class="text-sm font-semibold text-gray-700 mb-1">Invitation Link</h3>
                    <div class="flex gap-1">
                        <input class="border rounded px-2 py-1 text-xs flex-1" readonly value={invitationLink} />
                        <button type="button" onclick={() => { navigator.clipboard.writeText(invitationLink); }} class="px-2 py-1 bg-gray-400 text-white rounded text-xs">Copy</button>
                    </div>
                </div>
            {/if}

            {#if showInviteInput && chatData?.type !== "private"}
                <div class="flex gap-1">
                    <input class="border rounded px-2 py-1 text-sm flex-1" placeholder="Username to invite" bind:value={inviteUsername} />
                    <button type="button" onclick={handleInviteUser} class="px-2 py-1 bg-blue-500 text-white rounded text-xs">Send</button>
                    <button type="button" onclick={() => showInviteInput = false} class="px-2 py-1 bg-gray-400 text-white rounded text-xs">Cancel</button>
                </div>
            {/if}

            {#if showNewTopic}
                {#if !newTopicType}
                    <div class="flex gap-2 justify-center pt-1">
                        <button type="button" onclick={() => selectTopicType('text_topic')} class="flex items-center gap-1.5 px-4 py-2 bg-blue-100 hover:bg-blue-200 rounded-lg text-sm transition-colors">
                            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                                <path d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2z"/>
                            </svg>
                            Text Topic
                        </button>
                        <button type="button" onclick={() => selectTopicType('voice_topic')} class="flex items-center gap-1.5 px-4 py-2 bg-green-100 hover:bg-green-200 rounded-lg text-sm transition-colors">
                            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                                <path d="M15 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm-9-2V7H4v3H1v2h3v3h2v-3h3v-2H6zm9 4c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
                            </svg>
                            Video Topic
                        </button>
                    </div>
                {:else}
                    <div class="flex gap-1">
                        <input class="border rounded px-2 py-1 text-sm flex-1" placeholder="Topic name" bind:value={newTopicName} />
                        <button type="button" onclick={handleCreateTopic} class="px-2 py-1 bg-green-500 text-white rounded text-xs">Create</button>
                        <button type="button" onclick={cancelNewTopic} class="px-2 py-1 bg-gray-400 text-white rounded text-xs">Cancel</button>
                    </div>
                {/if}
            {/if}

            <div class="flex flex-wrap justify-center gap-2 pt-2 border-t border-gray-400">
                {#if chatData?.type !== "private"}
                    <button type="button" onclick={() => showInviteInput = true} class="px-4 py-1.5 bg-blue-300 hover:bg-blue-400 rounded-lg text-sm transition-colors">Invite</button>
                    <button type="button" onclick={handleCreateInvitation} class="px-4 py-1.5 bg-purple-300 hover:bg-purple-400 rounded-lg text-sm transition-colors">Get Invite Link</button>
                {/if}
                <button type="button" onclick={() => showNewTopic = true} class="px-4 py-1.5 bg-green-300 hover:bg-green-400 rounded-lg text-sm transition-colors">New Topic</button>
                <button type="button" onclick={handleMute} class="px-4 py-1.5 bg-yellow-300 hover:bg-yellow-400 rounded-lg text-sm transition-colors">
                    {isMuted() ? 'Unmute' : 'Mute'}
                </button>
                <button type="button" onclick={handleLeave} class="px-4 py-1.5 bg-red-300 hover:bg-red-400 rounded-lg text-sm transition-colors">Leave</button>
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
