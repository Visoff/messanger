<script lang="ts">
    import "./layout.css";

    const { children } = $props();
</script>

<svelte:head>
    <title>Messanger</title>
</svelte:head>
{@render children()}
