<script lang="ts">
    let toasts: { id: number; message: string; type: 'error' | 'success' | 'info' }[] = $state([]);
    let nextId = 0;

    export function show(message: string, type: 'error' | 'success' | 'info' = 'error') {
        const id = nextId++;
        toasts = [...toasts, { id, message, type }];
        setTimeout(() => {
            toasts = toasts.filter(t => t.id !== id);
        }, 4000);
    }

    export function error(message: string) { show(message, 'error'); }
    export function success(message: string) { show(message, 'success'); }
    export function info(message: string) { show(message, 'info'); }
</script>

{#if toasts.length > 0}
    <div class="toast-container">
        {#each toasts as toast (toast.id)}
            <div class="toast toast-{toast.type}">
                <span>{toast.message}</span>
                <button onclick={() => toasts = toasts.filter(t => t.id !== toast.id)} class="toast-close">&times;</button>
            </div>
        {/each}
    </div>
{/if}

<style>
    .toast-container {
        position: fixed;
        bottom: 20px;
        right: 20px;
        z-index: 1000;
        display: flex;
        flex-direction: column;
        gap: 8px;
        max-width: 360px;
    }

    @media (max-width: 767px) {
        .toast-container {
            left: 16px;
            right: 16px;
            max-width: none;
        }
    }

    .toast {
        padding: 12px 16px;
        border-radius: 8px;
        color: white;
        font-size: 14px;
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
        box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        animation: slideIn 0.2s ease;
    }

    .toast-error { background: #e53935; }
    .toast-success { background: #43a047; }
    .toast-info { background: #1e88e5; }

    .toast-close {
        background: transparent;
        border: none;
        color: white;
        font-size: 18px;
        cursor: pointer;
        padding: 0;
        line-height: 1;
    }

    @keyframes slideIn {
        from { transform: translateX(100%); opacity: 0; }
        to { transform: translateX(0); opacity: 1; }
    }
</style>
