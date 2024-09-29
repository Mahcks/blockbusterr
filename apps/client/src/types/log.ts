export interface Log {
    id: number;
    label: string;
    level: "info" | "warning" | "error" | "debug";
    message: string;
    timestamp: string;
}