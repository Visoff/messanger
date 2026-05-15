<script lang="ts">
    import Dialog from "./Dialog.svelte";
    import { createChat, createPrivateChat } from "$lib/api/chats";
    import { resolveUsername } from "$lib/api/users";

    let dialog: Dialog;

    let dialog_mode: "group" | "private" | "channel" = "group";

    function submitform(e: SubmitEvent) {
        e.preventDefault();
        if (dialog_mode == "group") {
            const title = e.target["title"].value;
            createChat(title).then(() => {
                location.reload();
            });
        } else if (dialog_mode == "private") {
            const username = e.target["username"].value;
            resolveUsername(username).then((user) => {
                if ("error" in user) return;
                createPrivateChat(user.id).then(() => {
                    location.reload();
                });
            });
        }
    }
</script>

<div>
    <button class="add-chat-btn" onclick={dialog.open} title="New Chat">
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
                    dialog_mode = e.target.value;
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
            <button type="submit">Create</button>
        </form>
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
