import { API_URL } from "./env";
import type { Chat, ChatWithLastMessage, User, ErrorResponse } from '../types';

export async function fetchChats(): Promise<ChatWithLastMessage[] | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/chats/`, {
        headers: {
            Authorization: `Bearer ${token}`,
        },
    });
    const data = await response.json() || [];
    return data;
}

export async function fetchChat(id: string): Promise<Chat | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/chats/${id}`, {
        headers: {
            Authorization: `Bearer ${token}`,
        },
    });
    const data = await response.json() || [];
    return data;
}

export async function createChat(title: string): Promise<Chat | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/chats/group`, {
        method: 'POST',
        headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ title }),
    });
    const data = await response.json();
    return data;
}

export async function createPrivateChat(user_id: string): Promise<Chat | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/chats/private?user_id=${user_id}`, {
        method: 'POST',
        headers: {
            Authorization: `Bearer ${token}`,
        },
    });
    const data = await response.json();
    return data;
}

export async function InviteUserToChat(chat_id: string, user_id: string): Promise<Chat | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/chats/${chat_id}/invite/${user_id}`, {
        method: 'POST',
        headers: {
            Authorization: `Bearer ${token}`,
        },
    });
    const data = await response.json();
    return data;
}

export async function updateChat(chat_id: string, title: string): Promise<Chat | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/chats/${chat_id}`, {
        method: 'PUT',
        headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ title }),
    });
    const data = await response.json();
    return data;
}

export async function leaveChat(chat_id: string): Promise<{ status: string } | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/chats/${chat_id}/leave`, {
        method: 'POST',
        headers: {
            Authorization: `Bearer ${token}`,
        },
    });
    const data = await response.json();
    return data;
}

export async function muteChat(chat_id: string, muted: boolean): Promise<Chat | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/chats/${chat_id}/mute`, {
        method: 'PUT',
        headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ muted }),
    });
    const data = await response.json();
    return data;
}

export async function fetchChatMembers(chat_id: string): Promise<User[] | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/chats/${chat_id}/members`, {
        headers: {
            Authorization: `Bearer ${token}`,
        },
    });
    const data = await response.json() || [];
    return data;
}

export async function uploadChatAvatar(chat_id: string, file: File): Promise<Chat | ErrorResponse> {
    const token = localStorage.getItem('token');
    const formData = new FormData();
    formData.append('avatar', file);
    const response = await fetch(`${API_URL}/chats/${chat_id}/avatar`, {
        method: 'POST',
        headers: {
            Authorization: `Bearer ${token}`,
        },
        body: formData,
    });
    const data = await response.json();
    return data;
}

export async function createInvitation(chat_id: string): Promise<{ id: string } | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/chats/${chat_id}/invitation`, {
        method: 'POST',
        headers: {
            Authorization: `Bearer ${token}`,
        },
    });
    const data = await response.json();
    return data;
}
