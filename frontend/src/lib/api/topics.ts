import { API_URL } from "./env";
import type { Topic, ErrorResponse,TopicType } from '../types';

export async function fetchTopics(chat_id: string): Promise<Topic[] | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/chats/${chat_id}/topics`, {
        headers: {
            Authorization: `Bearer ${token}`,
        },
    });
    const data = await response.json() || [];
    return data;
}

export async function createTopic(chat_id: string, title: string, type: TopicType): Promise<Topic | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/chats/${chat_id}/topics`, {
        method: 'POST',
        headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ title, type }),
    });
    const data = await response.json();
    return data;
}

export async function fetchTopic(id: string): Promise<Topic | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/topics/${id}`, {
        headers: {
            Authorization: `Bearer ${token}`,
        },
    });
    const data = await response.json() || [];
    return data;
}
