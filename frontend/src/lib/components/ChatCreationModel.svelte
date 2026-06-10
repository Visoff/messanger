<script lang="ts">
    import Dialog from "./Dialog.svelte";
    import { createChat, createPrivateChat } from "$lib/api/chats";
    import { resolveUsername } from "$lib/api/users";
    import type { User } from "$lib/types";

    let dialog: Dialog | undefined = $state();
    let userPopup: Dialog;
    let dialog_mode: "group" | "private" | "channel" = $state("group");
    let foundUser: User | null = $state(null);
    let searchError: string = $state("");
    let creating = $state(false);

    function submitform(e: SubmitEvent) {
        e.preventDefault();
        if (dialog_mode == "group") {
            creating = true;
            const title = (e.target as HTMLFormElement)["title"].value;
            createChat(title).then(() => {
                location.reload();
            });
        } else if (dialog_mode == "private") {
            const username = (e.target as HTMLFormElement)["username"].value;
            searchError = "";
            foundUser = null;
            resolveUsername(username).then((user) => {
                if ("error" in user) {
                    searchError = user.error ?? "User not found";
                    return;
                }
                foundUser = user;
                userPopup.open();
            });
        }
    }

    async function startChat() {
        if (!foundUser) return;
        creating = true;
        const resp = await createPrivateChat(foundUser.id);
        if ("error" in resp) {
            searchError = resp.error ?? "Failed to create chat";
            creating = false;
            return;
        }
        userPopup.close();
        dialog?.close();
        location.reload();
    }

    function startCall() {
        if (!foundUser) return;
        const url = new URL(`${window.location.origin}/conference`);
        window.open(url.toString(), "_blank", "width=960,height=640");
        userPopup.close();
    }

    function getInitials(name: string): string {
        return name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase();
    }
</script>

<div>
    <button class="add-chat-btn" onclick={dialog?.open} title="New Chat">
        <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
            <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z" />
        </svg>
    </button>
    <Dialog bind:this={dialog}>
        <form
            class="bg-gray-200 border border-gray-600 rounded-lg px-4 py-2 flex flex-col gap-2"
            onsubmit={submitform}
        >
            <select
                class="bold text-2xl"
                oninput={(e) => {
                    dialog_mode = (e.target as HTMLSelectElement).value as "group" | "private" | "channel";
                }}
                value={dialog_mode}
            >
                <option class="text-sm" value="group">Create Chat</option>
                <option class="text-sm" value="private">Find User</option>
                <option class="text-sm" value="channel">Create channel</option>
            </select>
            {#if dialog_mode === "group"}
                <input type="text" name="title" placeholder="Chat name" />
            {/if}
            {#if dialog_mode === "private"}
                <input type="text" name="username" placeholder="Username" />
            {/if}
            {#if dialog_mode === "channel"}
                <input type="text" placeholder="Channel name" />
            {/if}
            <button type="submit">{creating ? 'Creating...' : 'Search'}</button>
        </form>
    </Dialog>

    <Dialog bind:this={userPopup}>
        {#if foundUser}
            <div class="bg-gray-200 border border-gray-600 rounded-lg px-4 py-2 flex flex-col gap-2 min-w-64">
                <div class="flex flex-col items-center gap-2 py-2">
                    <div class="w-16 h-16 rounded-full bg-gradient-to-br from-blue-500 to-blue-700 flex items-center justify-center text-white font-bold text-xl">
                        {getInitials(foundUser.username)}
                    </div>
                    <h2 class="text-lg font-bold">@{foundUser.username}</h2>
                    {#if foundUser.last_seen_at}
                        <span class="text-xs text-gray-500">Last seen: {new Date(foundUser.last_seen_at).toLocaleDateString()}</span>
                    {/if}
                </div>

                {#if searchError}
                    <div class="text-red-500 text-xs text-center">{searchError}</div>
                {/if}

                <div class="flex justify-center gap-2 pt-2 border-t border-gray-400">
                    <button onclick={startChat} disabled={creating} class="px-4 py-1.5 bg-blue-300 hover:bg-blue-400 rounded-lg text-sm transition-colors disabled:opacity-50">
                        {creating ? '...' : 'Start Chat'}
                    </button>
                    <button onclick={startCall} class="px-4 py-1.5 bg-green-300 hover:bg-green-400 rounded-lg text-sm transition-colors">Call</button>
                    <button onclick={() => userPopup.close()} class="px-4 py-1.5 bg-gray-300 hover:bg-gray-400 rounded-lg text-sm transition-colors">Close</button>
                </div>
            </div>
        {/if}
    </Dialog>
</div>

<style>
    .add-chat-btn {
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

    .add-chat-btn:hover {
        background: #e8eaed;
    }
</style>
