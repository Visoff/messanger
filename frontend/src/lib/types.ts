export type ErrorResponse = {
    error: string,
    status: number,
    message: string
}

export type User = {
    id: string,
    username: string,
    avatar_url: string,
    metadata: string,
    created_at: Date,
    updated_at: Date,
    deleted_at: Date,
    last_seen_at: Date
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

export type TopicType = 'text_topic' | 'voice_topic';

export type Topic = {
    id: string,
    chat_id: string,
    title: string,
    avatar_url: string,
    type: TopicType,
    created_at: Date,
    updated_at: Date,
    deleted_at: Date
};

export type Message = {
    id: string,
    chat_id: string,
    topic_id?: string,
    sender_id: string,
    reply_message_id?: string,
    content?: string,
    created_at: Date,
    updated_at: Date,
    deleted_at: Date
};

export type ChatWithLastMessage = Chat & {
    last_message: Message | null;
};
