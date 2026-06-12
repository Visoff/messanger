<script lang="ts">
    import { login, register } from "$lib/api/auth";
    import { goto } from "$app/navigation";

    let {
        mode = "login",
    }: { mode?: "login" | "register" } = $props();

    function getRedirectUrl(): string {
        if (typeof window === "undefined") return "/";
        const params = new URLSearchParams(window.location.search);
        return params.get("redirect") || "/";
    }

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
            goto(getRedirectUrl());
        })
    }
</script>

<div class="login-card">
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
</div>

<style>
    .login-card {
        width: 100%;
        max-width: 400px;
        padding: 0 24px;
    }

    .login-form {
        width: 100%;
        padding: 40px 32px;
        display: flex;
        flex-direction: column;
        gap: 24px;
    }

    @media (min-width: 768px) {
        .login-form {
            background: white;
            border-radius: 16px;
            box-shadow: 0 4px 24px rgba(0,0,0,0.06), 0 1px 4px rgba(0,0,0,0.04);
        }
    }

    .form-title {
        font-size: 26px;
        font-weight: 700;
        margin: 0;
        text-align: center;
        letter-spacing: -0.3px;
    }

    .form-fields {
        display: flex;
        flex-direction: column;
        gap: 14px;
    }

    .input-group {
        display: flex;
        flex-direction: column;
    }

    .form-input {
        padding: 14px 16px;
        border: 1.5px solid #e6e8eb;
        border-radius: 10px;
        font-size: 15px;
        outline: none;
        transition: border-color 0.2s ease, box-shadow 0.2s ease;
        background: #fafafa;
    }

    .form-input:focus {
        border-color: #2481d2;
        box-shadow: 0 0 0 3px rgba(36,129,210,0.12);
        background: white;
    }

    .form-input::placeholder {
        color: #b0b3b8;
    }

    .submit-btn {
        padding: 14px 24px;
        border: none;
        border-radius: 10px;
        background: #2481d2;
        color: white;
        font-size: 15px;
        font-weight: 600;
        cursor: pointer;
        transition: background 0.2s ease, transform 0.15s ease, box-shadow 0.2s ease;
        letter-spacing: 0.2px;
        margin-top: 4px;
    }

    .submit-btn:hover {
        background: #1c6ea8;
        box-shadow: 0 4px 12px rgba(36,129,210,0.25);
    }

    .submit-btn:active {
        transform: scale(0.98);
    }

    .switch-mode-btn {
        padding: 12px;
        border: none;
        background: transparent;
        color: #8e8e93;
        font-size: 14px;
        cursor: pointer;
        transition: color 0.2s ease;
    }

    .switch-mode-btn:hover {
        color: #2481d2;
    }

    .switch-mode-btn span {
        color: #2481d2;
        font-weight: 500;
    }
</style>
