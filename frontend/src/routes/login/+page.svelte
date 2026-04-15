<script lang="ts">
    import LoginForm from "$lib/components/LoginForm.svelte";
    import { onMount } from "svelte";

    onMount(() => {
        const token = localStorage.getItem("token");
        if (token) {
            window.location.href = "/";
        }
    })
</script>

<div class="w-screen h-screen flex flex-row justify-between">
    <div class="flex-1 flex flex-col items-center justify-center">QR</div>
    <div class="flex-1 flex items-center justify-center">
        <LoginForm />
    </div>
</div>
