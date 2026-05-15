import { API_URL } from "./env";
import type { Chat,ErrorResponse } from '../types';

export async function fetchChats(): Promise<Chat[] | ErrorResponse> {
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

export async function InviteUserToChat(chat_id: string, user_id: string): Promise<Chat | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/chats/${chat_id}/invite`, {
        method: 'POST',
        headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ user_id }),
    });
    const data = await response.json();
    return data;
}
