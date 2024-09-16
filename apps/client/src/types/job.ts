export interface JobStatus {
    job_id: string;
    job_type: string;
    last_run: Date;
    next_run: Date;
    interval: number;
}