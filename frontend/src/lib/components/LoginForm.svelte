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
    <h1 class="text-3xl font-bold">Войти</h1>
    <label class="text-xl cursor-text" for="username">Имя пользователя:</label>
    <input placeholder="Имя пользователя" id="username" name="username" />
    <label class="text-xl cursor-text" for="password">Пароль:</label>
    <input type="password" placeholder="Пароль" id="password" name="password" />
    <button class="ml-1 font-bold cursor-pointer" type="submit">Войти</button>
    <button class="text-sm italic cursor-pointer" type="button" onclick={() => {mode = "register"}}>Нет аккаунта? <span class="text-blue-400">Зарегистрировать</span></button>
{:else}
    <h1 class="text-3xl font-bold">Регистрация</h1>
    <label class="text-xl cursor-text" for="username">Имя пользователя:</label>
    <input placeholder="Имя пользователя" id="username" name="username" />
    <label class="text-xl cursor-text" for="password">Пароль:</label>
    <input type="password" placeholder="Пароль" id="password" name="password" />
    <button class="ml-1 font-bold cursor-pointer" type="submit">Зарегистрироваться</button>
    <button class="text-sm italic cursor-pointer" type="button" onclick={() => {mode = "login"}}>Уже есть аккаунт? <span class="text-blue-400">Войти</span></button>
{/if}
</form>
