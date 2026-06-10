<script lang="ts">
    import type { User } from "$lib/types";

    let { otherUser }: { otherUser: User } = $props();

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

    function getInitials(name: string): string {
        return name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase();
    }
</script>

<div class="flex flex-col items-center gap-2 py-4">
    <div class="w-20 h-20 rounded-full bg-gradient-to-br from-blue-400 to-blue-600 flex items-center justify-center text-white font-bold text-2xl overflow-hidden flex-shrink-0">
        {#if otherUser.avatar_url}
            <img src={otherUser.avatar_url} alt="avatar" class="w-full h-full object-cover" />
        {:else}
            {getInitials(otherUser.username)}
        {/if}
    </div>
    <span class="text-xl font-bold">{otherUser.username}</span>
    <span class="text-xs px-2 py-0.5 rounded-full bg-blue-100 text-blue-700">Private Chat</span>
</div>

<div class="text-sm text-gray-600 space-y-1">
    <div class="flex justify-between">
        <span class="text-gray-500">User ID</span>
        <span class="font-mono text-xs">{otherUser.id}</span>
    </div>
    <div class="flex justify-between">
        <span class="text-gray-500">Last seen</span>
        <span>{formatDate(otherUser.last_seen_at)}</span>
    </div>
    <div class="flex justify-between">
        <span class="text-gray-500">Joined</span>
        <span>{formatDate(otherUser.created_at)}</span>
    </div>
</div>
