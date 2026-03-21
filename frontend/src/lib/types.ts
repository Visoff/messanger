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
