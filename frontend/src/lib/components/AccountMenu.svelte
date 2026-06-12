<script lang="ts">
    import Dialog from "./Dialog.svelte";
    import { goto } from "$app/navigation";
    import { user } from "$lib/stores/user";
    import { updateUser, uploadAvatar } from "$lib/api/auth";

    let dialog: Dialog;
    let editingField: string | null = $state(null);
    let editValue: string = $state("");
    let errorMsg: string = $state("");
    let uploadingAvatar = $state(false);
    let fileInput: HTMLInputElement;

    export function open() { dialog.open(); }

    function startEdit(field: string, currentValue: string) {
        editingField = field;
        editValue = currentValue;
    }

    async function saveEdit() {
        if (!editingField || !$user) return;
        const update: Record<string, string | null> = {};
        if (editingField === 'username') update.username = editValue;

        const resp = await updateUser(update);
        if ("error" in resp) {
            errorMsg = resp.error;
            return;
        }
        user.set(resp);
        editingField = null;
        errorMsg = "";
    }

    function cancelEdit() { editingField = null; errorMsg = ""; }

    async function handleAvatarUpload(e: Event) {
        const target = e.target as HTMLInputElement;
        const file = target.files?.[0];
        if (!file || !$user) return;

        uploadingAvatar = true;
        errorMsg = "";
        const resp = await uploadAvatar(file);
        if ("error" in resp) {
            errorMsg = resp.error;
        } else {
            user.set(resp);
        }
        uploadingAvatar = false;
        target.value = "";
    }

    function logout() {
        localStorage.removeItem("token");
        user.set(undefined);
        goto("/login");
    }

    function formatDate(d: Date | string): string {
        if (!d) return "—";
        return new Date(d).toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
    }

    function getInitials(name: string): string {
        return name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase();
    }
</script>

<Dialog bind:this={dialog}>
    <div class="bg-gray-200 border border-gray-600 rounded-lg px-4 py-2 flex flex-col gap-2 min-w-72">
        <h2 class="text-xl font-bold text-center">Account</h2>

        {#if $user}
            <div class="flex flex-col items-center gap-2 py-2">
                <button onclick={() => fileInput.click()} class="relative w-16 h-16 rounded-full bg-gradient-to-br from-blue-500 to-blue-700 flex items-center justify-center text-white font-bold text-xl overflow-hidden hover:opacity-80 transition-opacity group" title="Change avatar">
                    {#if $user.avatar_url}
                        <img src={$user.avatar_url} alt="avatar" class="w-full h-full object-cover" />
                    {:else}
                        {getInitials($user.username)}
                    {/if}
                    <div class="absolute inset-0 bg-black/40 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                        <svg viewBox="0 0 24 24" width="20" height="20" fill="white"><path d="M12 15.5a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7z" opacity="0"/><path d="M12 15.5a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7zM17 5h-1.5l-1-1h-5l-1 1H7v2h10V5zM7 7v10a2 2 0 0 0 2 2h6a2 2 0 0 0 2-2V7H7z"/></svg>
                    </div>
                </button>
                {#if uploadingAvatar}
                    <span class="text-xs text-gray-500">Uploading...</span>
                {/if}
                <input bind:this={fileInput} type="file" accept="image/*" class="hidden" onchange={handleAvatarUpload} />
            </div>

            <div class="text-sm space-y-2">
                {#if editingField === 'username'}
                    <div class="flex flex-col gap-1">
                        <span class="text-gray-500 text-xs">Username</span>
                        <input class="border rounded px-2 py-1 text-sm" bind:value={editValue} />
                        <div class="flex gap-2">
                            <button onclick={saveEdit} class="px-3 py-1 bg-blue-500 text-white rounded text-xs">Save</button>
                            <button onclick={cancelEdit} class="px-3 py-1 bg-gray-400 text-white rounded text-xs">Cancel</button>
                        </div>
                    </div>
                {:else}
                    <button onclick={() => startEdit('username', $user.username)} class="w-full text-left flex justify-between items-center hover:bg-gray-300 px-2 py-1 rounded transition-colors">
                        <span class="text-gray-500">Username</span>
                        <span>{$user.username} ✎</span>
                    </button>
                {/if}

                <div class="flex justify-between px-2 py-1">
                    <span class="text-gray-500">Created</span>
                    <span>{formatDate($user.created_at)}</span>
                </div>
                <div class="flex justify-between px-2 py-1">
                    <span class="text-gray-500">Last seen</span>
                    <span>{formatDate($user.last_seen_at)}</span>
                </div>
            </div>

            {#if errorMsg}
                <div class="text-red-500 text-xs text-center">{errorMsg}</div>
            {/if}

            <div class="flex justify-center gap-2 pt-2 border-t border-gray-400">
                <button onclick={logout} class="px-4 py-1.5 bg-red-400 hover:bg-red-500 text-white rounded-lg text-sm transition-colors">Logout</button>
                <button onclick={dialog.close} class="px-4 py-1.5 bg-gray-300 hover:bg-gray-400 rounded-lg text-sm transition-colors">Close</button>
            </div>
        {/if}
    </div>
</Dialog>
