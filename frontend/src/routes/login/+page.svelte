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

<div class="login-page">
    <div class="login-illustration">
        <svg viewBox="0 0 360 400" class="login-svg">
            <defs>
                <linearGradient id="blueGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                    <stop offset="0%" style="stop-color:#2481d2"/>
                    <stop offset="100%" style="stop-color:#0d7ad6"/>
                </linearGradient>
            </defs>
            <rect fill="url(#blueGrad)" width="360" height="400" rx="0"/>
            <g fill="#fff" opacity="0.9">
                <rect x="60" y="80" width="240" height="240" rx="16"/>
                <circle cx="180" cy="100" r="24"/>
                <rect x="80" y="140" width="200" height="8" rx="4"/>
                <rect x="80" y="160" width="160" height="8" rx="4"/>
                <rect x="80" y="180" width="180" height="8" rx="4"/>
                <circle cx="100" cy="280" r="32" opacity="0.5"/>
                <circle cx="260" cy="280" r="32" opacity="0.5"/>
            </g>
        </svg>
    </div>
    <div class="login-form-container">
        <LoginForm />
    </div>
</div>

<style>
    .login-page {
        display: flex;
        height: 100vh;
    }

    .login-illustration {
        flex: 1;
        background: linear-gradient(135deg, #2481d2, #0d7ad6);
        display: none;
    }

    @media (min-width: 768px) {
        .login-illustration {
            display: flex;
            align-items: center;
            justify-content: center;
        }

        .login-form-container {
            flex: 1;
            display: flex;
            align-items: center;
            justify-content: center;
        }
    }

    .login-svg {
        height: 400px;
        width: 100%;
        max-width: 360px;
    }
</style>