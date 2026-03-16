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

export async function createChat(title: string): Promise<Chat | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/chats/`, {
        method: 'POST',
        headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ title, type: "private" }),
    });
    const data = await response.json();
    return data;
}
