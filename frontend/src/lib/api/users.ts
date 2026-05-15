import type { User, ErrorResponse } from '../types';
import { API_URL } from "./env";

export async function resolveUsername(username: string): Promise<User | ErrorResponse> {
    const response = await fetch(`${API_URL}/users/username/${username}`);
    const data = await response.json();
    return data;
}
