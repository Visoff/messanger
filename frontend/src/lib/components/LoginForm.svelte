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

<div class="login-form">
    <h1 class="form-title">{mode === "login" ? "Вход" : "Регистрация"}</h1>
    <form onsubmit={submit} class="form-fields">
        <div class="input-group">
            <input placeholder="Имя пользователя" id="username" name="username" class="form-input" />
        </div>
        <div class="input-group">
            <input type="password" placeholder="Пароль" id="password" name="password" class="form-input" />
        </div>
        <button type="submit" class="submit-btn">
            {mode === "login" ? "Войти" : "Зарегистрироваться"}
        </button>
    </form>
    <button class="switch-mode-btn" type="button" onclick={() => {mode = mode === "login" ? "register" : "login"}}>
        {#if mode === "login"}
            Нет аккаунта? <span>Зарегистрироваться</span>
        {:else}
            Уже есть аккаунт? <span>Войти</span>
        {/if}
    </button>
</div>

<style>
    .login-form {
        width: 100%;
        max-width: 320px;
        padding: 24px;
        display: flex;
        flex-direction: column;
        gap: 20px;
    }

    .form-title {
        font-size: 24px;
        font-weight: 600;
        margin: 0;
        text-align: center;
    }

    .form-fields {
        display: flex;
        flex-direction: column;
        gap: 16px;
    }

    .input-group {
        display: flex;
        flex-direction: column;
    }

    .form-input {
        padding: 14px 16px;
        border: 1px solid #e6e8eb;
        border-radius: 8px;
        font-size: 15px;
        outline: none;
        transition: border-color 0.15s ease;
    }

    .form-input:focus {
        border-color: #2481d2;
    }

    .submit-btn {
        padding: 14px 24px;
        border: none;
        border-radius: 8px;
        background: #2481d2;
        color: white;
        font-size: 15px;
        font-weight: 500;
        cursor: pointer;
        transition: background 0.15s ease;
    }

    .submit-btn:hover {
        background: #1c6ea8;
    }

    .switch-mode-btn {
        padding: 12px;
        border: none;
        background: transparent;
        color: #8e8e93;
        font-size: 14px;
        cursor: pointer;
        transition: color 0.15s ease;
    }

    .switch-mode-btn:hover {
        color: #2481d2;
    }

    .switch-mode-btn span {
        color: #2481d2;
    }
</style>