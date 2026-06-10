<script lang="ts">
    import Dialog from "./Dialog.svelte";
    import { user } from "$lib/stores/user";
    import { updateUser } from "$lib/api/auth";

    let dialog: Dialog;
    let editingField: string | null = $state(null);
    let editValue: string = $state("");
    let errorMsg: string = $state("");

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

    function logout() {
        localStorage.removeItem("token");
        user.set(undefined);
        window.location.href = "/login";
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
                <div class="w-16 h-16 rounded-full bg-gradient-to-br from-blue-500 to-blue-700 flex items-center justify-center text-white font-bold text-xl">
                    {getInitials($user.username)}
                </div>
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
