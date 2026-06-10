<script lang="ts">
    import { user } from "$lib/stores/user";

    export interface UserCategory {
        id: string;
        name: string;
        chatIds: string[];
    }

    let {
        activeCategory,
        oncategorieselect,
        onaddchat,
    }: {
        activeCategory: { mode: string; category?: UserCategory } | null;
        oncategorieselect: (mode: string, category?: UserCategory) => void;
        onaddchat: (categoryId: string, chatTitle: string) => void;
    } = $props();

    let isOpen = $state(false);
    let categories: UserCategory[] = $state([]);
    let showAddCategory = $state(false);
    let newCategoryName = $state("");
    let showAddChatInput = $state(false);
    let newChatTitle = $state("");

    function storageKey(): string {
        return `categories_${$user?.id || "default"}`;
    }

    function loadCategories() {
        try {
            const key = storageKey();
            const stored = localStorage.getItem(key);
            categories = stored ? JSON.parse(stored) : [];
        } catch {
            categories = [];
        }
    }

    function saveCategories() {
        localStorage.setItem(storageKey(), JSON.stringify(categories));
    }

    let prevUserId: string | undefined;

    $effect(() => {
        const id = $user?.id;
        if (!id || id === prevUserId) return;
        prevUserId = id;
        loadCategories();
        oncategorieselect("all");
    });

    function toggleDropdown() {
        isOpen = !isOpen;
        if (isOpen) {
            loadCategories();
            showAddCategory = false;
            newCategoryName = "";
        }
    }

    function selectAllChats() {
        oncategorieselect("all");
        isOpen = false;
    }

    function selectPersonal() {
        oncategorieselect("personal");
        isOpen = false;
    }

    function selectCategory(cat: UserCategory) {
        oncategorieselect("category", cat);
        isOpen = false;
    }

    function createCategory() {
        if (!newCategoryName.trim()) return;
        const newCat: UserCategory = {
            id: crypto.randomUUID(),
            name: newCategoryName.trim(),
            chatIds: [],
        };
        categories = [...categories, newCat];
        saveCategories();
        showAddCategory = false;
        selectCategory(newCat);
    }

    function handleAddChatClick() {
        showAddChatInput = true;
        newChatTitle = "";
    }

    function confirmAddChat() {
        if (!newChatTitle.trim() || !activeCategory?.category) return;
        onaddchat(activeCategory.category.id, newChatTitle.trim());
        showAddChatInput = false;
        newChatTitle = "";
    }

    function handleClickOutside(e: MouseEvent) {
        const target = e.target as HTMLElement;
        if (!target.closest('.category-container')) {
            isOpen = false;
        }
    }

    $effect(() => {
        if (isOpen) {
            window.addEventListener('click', handleClickOutside);
            return () => window.removeEventListener('click', handleClickOutside);
        }
    });

    function getButtonLabel(): string {
        if (!activeCategory) return "Category";
        if (activeCategory.mode === "all") return "All Chats";
        if (activeCategory.mode === "personal") return "Personal";
        if (activeCategory.mode === "category" && activeCategory.category) return activeCategory.category.name;
        return "Category";
    }
</script>

<div class="category-container">
    <button class="category-btn" onclick={toggleDropdown}>
        <span>{getButtonLabel()}</span>
        <svg class:rotated={isOpen} viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
            <path d="M7 10l5 5 5-5z" />
        </svg>
    </button>

    {#if isOpen}
        <div class="dropdown" onclick={(e) => e.stopPropagation()}>
            <button class="dropdown-item" onclick={selectAllChats}>All Chats</button>
            <button class="dropdown-item" onclick={selectPersonal}>Personal</button>

            {#each categories as cat (cat.id)}
                <button class="dropdown-item" onclick={() => selectCategory(cat)}>
                    {cat.name}
                </button>
            {/each}

            {#if showAddCategory}
                <div class="add-category-form">
                    <input
                        type="text"
                        placeholder="Category name"
                        bind:value={newCategoryName}
                        onkeydown={(e) => { if (e.key === 'Enter') createCategory(); }}
                    />
                    <button onclick={createCategory}>Create</button>
                </div>
            {:else}
                <button class="dropdown-item add-category-btn" onclick={() => { showAddCategory = true; }}>
                    + Add Category
                </button>
            {/if}
        </div>
    {/if}

    {#if activeCategory?.mode === "category" && activeCategory.category}
        <div class="add-chat-section">
            {#if showAddChatInput}
                <div class="add-chat-form">
                    <input
                        type="text"
                        placeholder="Enter chat title to add"
                        bind:value={newChatTitle}
                        onkeydown={(e) => { if (e.key === 'Enter') confirmAddChat(); }}
                    />
                    <button onclick={confirmAddChat}>Add</button>
                    <button class="cancel-btn" onclick={() => { showAddChatInput = false; }}>Cancel</button>
                </div>
            {:else}
                <button class="add-chat-btn" onclick={handleAddChatClick}>
                    + Add chat to category
                </button>
            {/if}
        </div>
    {/if}
</div>

<style>
    .category-container {
        position: relative;
        border-bottom: 1px solid #e5e7eb;
    }

    .category-btn {
        display: flex;
        align-items: center;
        justify-content: space-between;
        width: 100%;
        padding: 10px 16px;
        border: none;
        background: transparent;
        font-size: 14px;
        font-weight: 500;
        color: #333;
        cursor: pointer;
        transition: background 0.15s ease;
    }

    .category-btn:hover {
        background: #f3f4f6;
    }

    .category-btn svg {
        transition: transform 0.2s ease;
    }

    .category-btn svg.rotated {
        transform: rotate(180deg);
    }

    .dropdown {
        position: absolute;
        top: 100%;
        left: 0;
        right: 0;
        background: white;
        border: 1px solid #e5e7eb;
        border-radius: 8px;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
        z-index: 100;
        padding: 4px;
    }

    .dropdown-item {
        display: block;
        width: 100%;
        padding: 8px 12px;
        border: none;
        background: transparent;
        text-align: left;
        font-size: 14px;
        color: #333;
        border-radius: 6px;
        cursor: pointer;
        transition: background 0.15s ease;
    }

    .dropdown-item:hover {
        background: #f3f4f6;
    }

    .add-category-form {
        display: flex;
        gap: 4px;
        padding: 4px;
    }

    .add-category-form input {
        flex: 1;
        padding: 6px 8px;
        border: 1px solid #d1d5db;
        border-radius: 6px;
        font-size: 13px;
        outline: none;
    }

    .add-category-form input:focus {
        border-color: #3b82f6;
    }

    .add-category-form button {
        padding: 6px 12px;
        border: none;
        background: #3b82f6;
        color: white;
        border-radius: 6px;
        font-size: 13px;
        cursor: pointer;
        transition: background 0.15s ease;
    }

    .add-category-form button:hover {
        background: #2563eb;
    }

    .add-category-btn {
        color: #3b82f6;
        font-weight: 500;
    }

    .add-chat-section {
        border-bottom: 1px solid #e5e7eb;
    }

    .add-chat-btn {
        display: block;
        width: 100%;
        padding: 8px 16px;
        border: none;
        background: transparent;
        text-align: left;
        font-size: 13px;
        color: #3b82f6;
        font-weight: 500;
        cursor: pointer;
        transition: background 0.15s ease;
    }

    .add-chat-btn:hover {
        background: #f3f4f6;
    }

    .add-chat-form {
        display: flex;
        gap: 4px;
        padding: 6px 16px;
    }

    .add-chat-form input {
        flex: 1;
        padding: 6px 8px;
        border: 1px solid #d1d5db;
        border-radius: 6px;
        font-size: 13px;
        outline: none;
    }

    .add-chat-form input:focus {
        border-color: #3b82f6;
    }

    .add-chat-form button {
        padding: 6px 12px;
        border: none;
        background: #3b82f6;
        color: white;
        border-radius: 6px;
        font-size: 13px;
        cursor: pointer;
        transition: background 0.15s ease;
    }

    .add-chat-form button:hover {
        background: #2563eb;
    }

    .add-chat-form .cancel-btn {
        background: #9ca3af;
    }

    .add-chat-form .cancel-btn:hover {
        background: #6b7280;
    }
</style>
