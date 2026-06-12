<script lang="ts">
    let dialog: HTMLDialogElement;

    export function open() {
        dialog.showModal();
    }

    export function close() {
        dialog.close();
    }

    function dialog_click(e) {
        if (e.target === dialog) {
            close();
        }
    }
</script>

<div>
    <dialog onclick={dialog_click} bind:this={dialog} class="left-1/2 top-1/2 transform -translate-x-1/2 -translate-y-1/2 bg-transparent backdrop:opacity-40 backdrop:bg-gray-600 max-w-[calc(100vw-32px)]">
        <slot></slot>
    </dialog>
</div>
