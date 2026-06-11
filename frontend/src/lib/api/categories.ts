import { API_URL } from "./env";
import type { ErrorResponse } from '../types';

export interface UserCategoryResponse {
    id: string;
    user_id: string;
    name: string;
    chat_ids: string[];
    created_at: string;
    updated_at: string;
}

export async function fetchCategories(): Promise<UserCategoryResponse[] | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/categories/`, {
        headers: {
            Authorization: `Bearer ${token}`,
        },
    });
    const data = await response.json() || [];
    return data;
}

export async function createCategory(name: string, chatIds: string[] = []): Promise<UserCategoryResponse | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/categories/`, {
        method: 'POST',
        headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name, chat_ids: chatIds }),
    });
    const data = await response.json();
    return data;
}

export async function updateCategory(id: string, name: string, chatIds: string[]): Promise<UserCategoryResponse | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/categories/${id}`, {
        method: 'PUT',
        headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name, chat_ids: chatIds }),
    });
    const data = await response.json();
    return data;
}

export async function deleteCategory(id: string): Promise<{ status: string } | ErrorResponse> {
    const token = localStorage.getItem('token');
    const response = await fetch(`${API_URL}/categories/${id}`, {
        method: 'DELETE',
        headers: {
            Authorization: `Bearer ${token}`,
        },
    });
    const data = await response.json();
    return data;
}
