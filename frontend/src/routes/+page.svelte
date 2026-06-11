<script lang="ts">
    import { API_URL } from "$lib/api/env";
    import ChatList from "$lib/components/ChatList.svelte";
    import CategoryFilter from "$lib/components/CategoryFilter.svelte";
    import type { UserCategory } from "$lib/components/CategoryFilter.svelte";
    import { fetchCategories, updateCategory } from "$lib/api/categories";
    import { onMount } from "svelte";
    import { getMe } from "$lib/api/auth";
    import ChatCreationModel from "$lib/components/ChatCreationModel.svelte";
    import ChatPage from "$lib/components/ChatPage.svelte";
    import AccountMenu from "$lib/components/AccountMenu.svelte";
    import Toast from "$lib/components/Toast.svelte";
    import { extractFromSearchParams } from "$lib/index";
    import { user } from "$lib/stores/user";
    import { newMessageEvent } from "$lib/stores/messages";

    let toast: Toast;

    async function getServiceWorkerRegistration() {
        if (navigator.serviceWorker.controller) {
            return navigator.serviceWorker.ready;
        }
        const registration =
            await navigator.serviceWorker.register("/scripts/sw.js");
        return registration;
    }

    async function subscribeToPush() {
        if ("serviceWorker" in navigator === false) return;
        const registration = await getServiceWorkerRegistration();
        const sub = await registration.pushManager.getSubscription();
        if (sub) {
            return;
        }
        const permission = await Notification.requestPermission();
        if (permission !== "granted") return;

        const vapidPublicKey = await fetch(
            `${API_URL}/pubsub/push/pubkey`,
        ).then((r) => r.text());
        console.log(vapidPublicKey);
        const subscription = await registration.pushManager.subscribe({
            userVisibleOnly: true,
            applicationServerKey: vapidPublicKey,
        });
        const token = localStorage.getItem("token");
        await fetch(`${API_URL}/pubsub/push/subscribe`, {
            method: "POST",
            headers: {
                Authorization: `Bearer ${token}`,
                "Content-Type": "application/json",
            },
            body: JSON.stringify(subscription),
        });
    }

    let chat_id: string | undefined = $state(extractFromSearchParams("chat_id"));
    let chatListRefreshKey = $state(0);
    let activeCategory: { mode: string; category?: UserCategory } | null = $state(null);
    let loadedChats: import("$lib/types").ChatWithLastMessage[] = $state([]);
    const initialCategoryParam = extractFromSearchParams("category");

    onMount(() => {
        const token = localStorage.getItem("token");
        if (!token) {
            window.location.href = "/login";
        } else {
            console.log(token);
            getMe().then(async (resp) => {
                if ("error" in resp) {
                    console.error(resp.error);
                    return
                }
                user.set(resp);

                if (initialCategoryParam) {
                    if (initialCategoryParam === "personal" || initialCategoryParam === "groups") {
                        activeCategory = { mode: initialCategoryParam };
                    } else if (initialCategoryParam !== "all") {
                        const cats = await fetchCategories();
                        if (!("error" in cats)) {
                            const cat = cats.find(c => c.id === initialCategoryParam);
                            if (cat) {
                                activeCategory = { mode: "category", category: cat };
                            }
                        }
                    }
                }
            });
        }

        const prompt_notifications = () => {
            window.removeEventListener("click", prompt_notifications);
            subscribeToPush();
        };

        window.addEventListener("click", prompt_notifications);

        if (token) {
            const stream = new EventSource(`${API_URL}/pubsub/sse?token=${token}`);
            stream.addEventListener("message", (e) => {
                try {
                    const data = JSON.parse(e.data);
                    if (data.type === "chat_created" || data.type === "user_added_to_chat") {
                        chatListRefreshKey++;
                    } else if (data.chat_id) {
                        newMessageEvent.set(data);
                    }
                } catch {}
            });
        }
    });

    let accountMenu: AccountMenu;
    let chatPageRef: ChatPage | undefined = $state();

    function clearChat() {
        chat_id = undefined
        chatPageRef?.resetTopic();
        const url = new URL(location.href);
        url.searchParams.delete("chat_id");
        url.searchParams.delete("topic_id");
        history.pushState(null, "", url);
    }

    function handleCategorySelect(mode: string, category?: UserCategory) {
        if (!mode) {
            activeCategory = null;
        } else if (mode === "category" && category) {
            activeCategory = { mode, category };
        } else {
            activeCategory = { mode };
        }

        const url = new URL(location.href);
        if (!mode) {
            url.searchParams.delete("category");
        } else if (mode === "category" && category) {
            url.searchParams.set("category", category.id);
        } else {
            url.searchParams.set("category", mode);
        }
        history.replaceState(null, "", url);
    }

    async function handleAddChatToCategory(categoryId: string, chatTitle: string) {
        const chat = loadedChats.find(c => c.title.toLowerCase() === chatTitle.toLowerCase());
        if (!chat) return;
        const cat = activeCategory?.category;
        if (!cat) return;
        const chatIds = [...(cat.chat_ids || [])];
        if (chatIds.includes(chat.id)) return;
        chatIds.push(chat.id);
        const resp = await updateCategory(categoryId, cat.name, chatIds);
        if ("error" in resp) {
            console.error(resp.error);
            return;
        }
        activeCategory = { mode: "category", category: resp };
    }

    function getInitials(name: string | undefined): string {
        if (!name) return "?";
        return name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase();
    }
</script>

<Toast bind:this={toast} />

<main class="flex h-screen bg-white">
    <div class="md:w-80 md:relative min-w-80 border-r-gray-100 border-r flex flex-col bg-white w-full absolute">
        <div class="flex justify-between items-center p-4 border-b-gray-100 border-b">
            <div class="flex items-center gap-2">
                <button onclick={() => accountMenu?.open()} class="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-blue-700 flex items-center justify-center text-white font-bold text-xs overflow-hidden hover:opacity-80 transition-opacity" title="Account">
                    {#if $user?.avatar_url}
                        <img src={$user.avatar_url} alt="avatar" class="w-full h-full object-cover" />
                    {:else}
                        {getInitials($user?.username)}
                    {/if}
                </button>
                <span class="font-bold text-lg">Messanger</span>
            </div>
            <ChatCreationModel onchatstarted={(id) => {
                chat_id = id;
                chatListRefreshKey++;
                const url = new URL(location.href);
                url.searchParams.set("chat_id", id);
                url.searchParams.delete("topic_id");
                history.pushState(null, "", url);
            }} />
        </div>
        <CategoryFilter
            activeCategory={activeCategory}
            oncategorieselect={handleCategorySelect}
            onaddchat={handleAddChatToCategory}
        />
        <ChatList
            onselect={(id) => {chat_id = id; chatPageRef?.resetTopic();}}
            current_chat_id={chat_id}
            refreshKey={chatListRefreshKey}
            filterMode={activeCategory?.mode}
            categoryChatIds={activeCategory?.category?.chat_ids}
            onchatsload={(chats) => { loadedChats = chats; }}
        />
    </div>

    {#if chat_id}
        <ChatPage chat_id={chat_id} onclose={clearChat} bind:this={chatPageRef} />
    {:else}
        <div class="w-full flex flex-col items-center justify-center bg-gray-50">
            <div class="mb-4 opacity-50">
                <svg
                    viewBox="0 0 24 24"
                    width="64"
                    height="64"
                    fill="currentColor"
                >
                    <path
                        d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm0 14H6l-2 2V4h16v12z"
                    />
                </svg>
            </div>
            <p>Select a chat to start messaging</p>
        </div>
    {/if}
</main>

<AccountMenu bind:this={accountMenu} />
