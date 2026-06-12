<script lang="ts">
    import { onMount } from "svelte";
    import { verifyInvitation, acceptInvitation } from "$lib/api/chats";
    import { goto } from "$app/navigation";

    let inviteId = $state("");
    let loading = $state(true);
    let error = $state("");
    let info: {
        invitation: { id: string; created_at: string };
        chat: { id: string; title: string; type: string; avatar_url?: string; created_at: string };
        creator: { id: string; username: string; avatar_url?: string };
    } | null = $state(null);
    let actionLoading = $state(false);

    function getInitials(name: string): string {
        return name.split(" ").map(w => w[0]).join("").slice(0, 2).toUpperCase();
    }

    function formatDate(date: string): string {
        return new Date(date).toLocaleDateString(undefined, {
            year: "numeric", month: "short", day: "numeric",
        });
    }

    function getChatTypeLabel(type: string): string {
        switch (type) {
            case "private": return "Private Chat";
            case "group": return "Group";
            case "channel": return "Channel";
            default: return type;
        }
    }

    onMount(async () => {
        const params = new URLSearchParams(window.location.search);
        const id = params.get("invite_id");
        if (!id) {
            goto("/");
            return;
        }
        inviteId = id;

        const token = localStorage.getItem("token");
        if (!token) {
            goto(`/login?redirect=/invitation?invite_id=${id}`);
            return;
        }

        const resp = await verifyInvitation(id);
        if ("error" in resp) {
            goto("/");
            return;
        }
        info = resp;
        loading = false;
    });

    async function handleJoin() {
        if (!info || actionLoading) return;
        actionLoading = true;
        const resp = await acceptInvitation(inviteId);
        if ("error" in resp) {
            error = resp.error;
            actionLoading = false;
            return;
        }
        goto(`/?chat_id=${resp.chat_id}`);
    }

    async function handleReject() {
        goto("/");
    }
</script>

<div class="invitation-page">
    {#if loading}
        <div class="loading-state">
            <div class="spinner"></div>
            <p>Loading invitation...</p>
        </div>
    {:else if info}
        <div class="invitation-card">
            <div class="card-body">
                <div class="avatar-section">
                    {#if info.chat.avatar_url}
                        <img src={info.chat.avatar_url} alt="chat avatar" class="chat-avatar" />
                    {:else}
                        <div class="chat-avatar initials">
                            {getInitials(info.chat.title)}
                        </div>
                    {/if}
                </div>

                <h1 class="chat-title">{info.chat.title}</h1>
                <span class="chat-type-badge">{getChatTypeLabel(info.chat.type)}</span>

                <div class="creator-info">
                    <span class="creator-label">Invited by</span>
                    <div class="creator-detail">
                        {#if info.creator.avatar_url}
                            <img src={info.creator.avatar_url} alt="" class="creator-avatar" />
                        {:else}
                            <div class="creator-avatar initials-sm">
                                {getInitials(info.creator.username)}
                            </div>
                        {/if}
                        <span class="creator-name">@{info.creator.username}</span>
                    </div>
                </div>

                <div class="meta-row">
                    <span class="meta-label">Created</span>
                    <span class="meta-value">{formatDate(info.chat.created_at)}</span>
                </div>

                {#if error}
                    <div class="error-msg">{error}</div>
                {/if}
            </div>

            <div class="card-actions">
                <button class="btn btn-primary" onclick={handleJoin} disabled={actionLoading}>
                    {actionLoading ? "Joining..." : "Join Chat"}
                </button>
                <button class="btn btn-secondary" onclick={handleReject} disabled={actionLoading}>
                    {actionLoading ? "..." : "Decline"}
                </button>
            </div>
        </div>
    {/if}
</div>

<style>
    .invitation-page {
        display: flex;
        align-items: center;
        justify-content: center;
        min-height: 100vh;
        background: var(--tg-bg-secondary, #f5f7f9);
        padding: 24px;
    }

    .loading-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 16px;
        color: var(--tg-gray, #8e8e93);
    }

    .spinner {
        width: 36px;
        height: 36px;
        border: 3px solid #e6e8eb;
        border-top-color: var(--tg-blue, #2481d2);
        border-radius: 50%;
        animation: spin 0.7s linear infinite;
    }

    @keyframes spin {
        to { transform: rotate(360deg); }
    }

    .invitation-card {
        background: white;
        border-radius: 16px;
        box-shadow: 0 4px 24px rgba(0,0,0,0.06), 0 1px 4px rgba(0,0,0,0.04);
        width: 100%;
        max-width: 380px;
        overflow: hidden;
    }

    .card-body {
        padding: 32px 24px 20px;
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 10px;
    }

    .avatar-section {
        margin-bottom: 4px;
    }

    .chat-avatar {
        width: 80px;
        height: 80px;
        border-radius: 50%;
        object-fit: cover;
    }

    .chat-avatar.initials {
        background: linear-gradient(135deg, var(--tg-blue, #2481d2), #0d7ad6);
        display: flex;
        align-items: center;
        justify-content: center;
        color: white;
        font-size: 28px;
        font-weight: 700;
    }

    .chat-title {
        font-size: 22px;
        font-weight: 700;
        margin: 0;
        text-align: center;
    }

    .chat-type-badge {
        font-size: 11px;
        padding: 3px 10px;
        border-radius: 999px;
        background: #e5f3fd;
        color: var(--tg-blue, #2481d2);
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.3px;
    }

    .creator-info {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 6px;
        margin-top: 6px;
        padding-top: 14px;
        width: 100%;
        border-top: 1px solid #e6e8eb;
    }

    .creator-label {
        font-size: 12px;
        color: var(--tg-gray, #8e8e93);
    }

    .creator-detail {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .creator-avatar {
        width: 28px;
        height: 28px;
        border-radius: 50%;
        object-fit: cover;
    }

    .creator-avatar.initials-sm {
        width: 28px;
        height: 28px;
        border-radius: 50%;
        background: linear-gradient(135deg, #2481d2, #0d7ad6);
        display: flex;
        align-items: center;
        justify-content: center;
        color: white;
        font-size: 11px;
        font-weight: 600;
    }

    .creator-name {
        font-size: 14px;
        font-weight: 500;
        color: #333;
    }

    .meta-row {
        display: flex;
        justify-content: space-between;
        width: 100%;
        font-size: 13px;
        padding: 4px 0;
    }

    .meta-label {
        color: var(--tg-gray, #8e8e93);
    }

    .meta-value {
        color: #333;
    }

    .error-msg {
        color: #e53935;
        font-size: 13px;
        text-align: center;
    }

    .card-actions {
        display: flex;
        gap: 10px;
        padding: 0 24px 24px;
    }

    .btn {
        flex: 1;
        padding: 12px 16px;
        border: none;
        border-radius: 10px;
        font-size: 14px;
        font-weight: 600;
        cursor: pointer;
        transition: background 0.15s ease, opacity 0.15s ease;
        text-align: center;
    }

    .btn:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .btn-primary {
        background: var(--tg-blue, #2481d2);
        color: white;
    }

    .btn-primary:hover:not(:disabled) {
        background: var(--tg-blue-dark, #1c6ea8);
    }

    .btn-secondary {
        background: #e8eaed;
        color: #333;
    }

    .btn-secondary:hover:not(:disabled) {
        background: #d1d3d6;
    }
</style>
