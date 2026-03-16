<script lang="ts">
    import { login, register } from "$lib/api/auth";

    let {
        mode = "login",
    }: { mode?: "login" | "register" } = $props();

    function submit(e: SubmitEvent) {
        e.preventDefault();
        const data = {
            username: (e.target as HTMLFormElement).username.value,
            password: (e.target as HTMLFormElement).password.value,
        }
        
        let promise;
        if (mode === "login") {
            promise = login(data)
        } else {
            promise = register(data)
        }
        promise.then((resp) => {
            if ("error" in resp) {
                console.error(resp);
                return;
            }
            localStorage.setItem("token", resp.token);
            window.location.href = "/";
        })
    }
</script>

<form class="modal flex flex-col gap-2 items-start" onsubmit={submit}>
{#if mode === "login"}
    <h1>login</h1>
    <input placeholder="username" name="username" />
    <input type="password" placeholder="password" name="password" />
    <button type="submit">login</button>
    <button type="button" onclick={() => {mode = "register"}}>dont have an account</button>
{:else}
    <h1>register</h1>
    <input placeholder="username" name="username" />
    <input type="password" placeholder="password" name="password" />
    <button type="submit">register</button>
    <button type="button" onclick={() => {mode = "login"}}>already have an account</button>
{/if}
</form>
