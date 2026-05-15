<script lang="ts">
    import { createChat } from "$lib/api/chats";

    let dialog: HTMLDialogElement;

    function open_dialog() {
        dialog.showModal();
    }

    function close_dialog() {
        dialog.close();
    }

    function dialog_click(e) {
        if (e.target === dialog) {
            close_dialog();
        }
    }

    let dialog_mode: "group" | "private" | "channel" = "group";

    function submitform(e) {
        e.preventDefault()
        if (dialog_mode == "group") {
            const title = e.target["title"].value
            createChat(title).then(() => {
                location.reload()
            });
        }
    }
</script>

<div>
    <button class="add-chat-btn" onclick={open_dialog} title="New Chat">
        <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
            <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/>
        </svg>
    </button>
    <dialog onclick={dialog_click} bind:this={dialog} class="left-1/2 top-1/2 transform -translate-x-1/2 -translate-y-1/2 bg-transparent backdrop:opacity-40 backdrop:bg-gray-600">
        <form class="bg-gray-200 border border-gray-600 rounded-lg px-4 py-2 flex flex-col gap-2" onsubmit={submitform}>
            <select class="bold text-2xl" oninput={e => {dialog_mode = e.target.value}} value={dialog_mode}>
                <option class="text-sm" value="group">Create Chat</option>
                <option class="text-sm" value="private">Find User</option>
                <option class="text-sm" value="channel">Create channel</option>
            </select>
            {#if dialog_mode === "group"}
                <input type="text" name="title" placeholder="Chat name" />
            {/if}
            {#if dialog_mode === "private"}
                <input type="text" placeholder="Username" />
            {/if}
            {#if dialog_mode === "channel"}
                <input type="text" placeholder="Channel name" />
            {/if}
            <button type="submit">Create</button>
        </form>
    </dialog>
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
