export type ErrorResponse = {
    error: string,
    status: number,
    message: string
}

export type ChatType = 'private' | 'group' | 'channel';

export type Chat = {
    id: string,
    title: string,
    type: ChatType,
    avatar_url: string,
    metadata: string,
    created_at: Date,
    updated_at: Date,
    deleted_at: Date
};
