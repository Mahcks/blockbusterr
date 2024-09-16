/* eslint-disable @typescript-eslint/no-explicit-any */
export enum Code {
    // Default codes (0/10)
    CodeDispatch = 0, // Dispatches an event to client
    CodeHello = 1, // Sent immediately after connecting, contains heartbeat and session info
    CodeHeartbeat = 2, // Sent by client to keep connection alive
    CodeAck = 3, // Acknowledges a message

    // Command codes (11/20)
    CodeSubscribe = 11, // Subscribe to a topic
    CodeUnsubscribe = 12, // Unsubscribe from a topic
}

export interface Message {
    c: Code
    t: number
    d: any
}

export interface HelloPayload {
    session_id: string
}

export interface HeartbeatPayload {
    count: number
}

export interface DispatchPayload {
    topic: string
    data: any
}

export interface StreamOfflinePayload {
    broadcaster_user_id: string
    broadcaster_user_login: string
    broadcaster_user_name: string
}

export interface UserUpdatedPayload {
    user_id: string
    old_user_login: string
    user_login: string
    user_name: string
}

export interface ChannelTitlePayload {
    broadcaster_user_id: string
    title: string
}

export interface ChannelCategoryPayload {
    broadcaster_user_id: string
    category_id: string
    category_name: string
}

export interface ViewersPayload {
    viewers: Map<string, number>
}

export function NewMessage(code: Code, data: any) {
    return {
        c: code,
        t: Date.now(),
        d: data,
    }
}