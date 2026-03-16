<script lang="ts">
    import LoginModal from "$lib/components/LoginModal.svelte";
    import { onMount } from "svelte";

    onMount(() => {
        const token = localStorage.getItem("token");
        if (token) {
            window.location.href = "/";
        }
    })
</script>

<LoginModal />
