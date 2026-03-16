import { API_URL } from "./env";
import type { ErrorResponse } from '../types';

export async function login(creds: {username: string, password: string}): Promise<{token: string} | ErrorResponse> {
    const response = await fetch(`${API_URL}/users/login`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(creds),
    });
    const data = await response.json();
    return data;
}

export async function register(creds: {username: string, password: string}): Promise<{token: string} | ErrorResponse> {
    const response = await fetch(`${API_URL}/users/register`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(creds),
    });
    const data = await response.json();
    return data;
}
