export type APIErrorBody = {
    status_code: number;
    timestamp: number;
    error: APIError;
    trace_id?: string;
}

export type APIError = {
    status_code: number;
    error: string;
    error_code: number;
    details?: string[];
}