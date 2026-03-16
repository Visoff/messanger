<script lang="ts">
    import ChatList from "$lib/components/ChatList.svelte";
    import { onMount } from "svelte";

    onMount(() => {
        const token = localStorage.getItem("token");
        if (!token) {
            window.location.href = "/login";
        } else {
            console.log(token);
        }
    })

    function logout() {
        localStorage.removeItem("token");
        window.location.href = "/login";
    }
</script>

<main class="flex flex-row gap-5 items-start justify-start">
    <nav class="max-w-3xl">
        <ChatList />
    </nav>
    <button onclick={logout}>logout</button>
</main>
