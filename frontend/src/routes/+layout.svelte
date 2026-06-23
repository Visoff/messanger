<script lang="ts">
    import { onMount } from "svelte";
    import "./layout.css";
    import { init } from "$lib/init";

    const { children } = $props();

    onMount(init);
</script>

<svelte:head>
    <title>Messanger</title>
</svelte:head>
{@render children()}
