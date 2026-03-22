import { API_URL } from "./env";
import type { Message, ErrorResponse } from '../types';

export async function fetchMessages(chat_id: string, topic_id?: string): Promise<Message[] | ErrorResponse> {
    const token = localStorage.getItem('token');
    if (topic_id === undefined) {
        const response = await fetch(`${API_URL}/chats/${chat_id}/messages`, {
            headers: {
                Authorization: `Bearer ${token}`,
            },
        });
        const data = await response.json() || [];
        return data;
    }
    const response = await fetch(`${API_URL}/topics/${topic_id}/messages`, {
        headers: {
            Authorization: `Bearer ${token}`,
        },
    });
    const data = await response.json() || [];
    return data;
}

export async function sendMessage(chat_id: string, topic_id: string | undefined, content: string): Promise<Message | ErrorResponse> {
    const token = localStorage.getItem('token');
    if (topic_id === undefined) {
        const response = await fetch(`${API_URL}/chats/${chat_id}/messages`, {
            method: 'POST',
            headers: {
                Authorization: `Bearer ${token}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ content }),
        })
        const data = await response.json();
        return data;
    }
    const response = await fetch(`${API_URL}/topics/${topic_id}/messages`, {
        method: 'POST',
        headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ content }),
    })
    const data = await response.json();
    return data;
}
