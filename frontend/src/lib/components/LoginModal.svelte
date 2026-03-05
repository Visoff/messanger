<script lang="ts">
    let {
        mode = "login",
    }: { mode?: "login" | "register" } = $props();
</script>

<div class="modal">
{#if mode === "login"}
    <h1>login</h1>
    <button onclick={() => {mode = "register"}}>dont have an account</button>
{:else}
    <h1>register</h1>
    <button onclick={() => {mode = "login"}}>already have an account</button>
{/if}
</div>
