import { JobStatus } from "@/types/job";
import * as React from "react";

export default function JobStatusWidget() {
  const [status, setStatus] = React.useState<JobStatus[] | null>(null);

  const fetchJobStatus = async () => {
    try {
      const response = await fetch(`${import.meta.env.VITE_API_URL}/jobs/status`);
      const data = await response.json();
      setStatus(data);
    } catch {
      console.error("Error fetching job status");
    }
  };

  React.useEffect(() => {
    fetchJobStatus();
  }, []);

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString();
  };

  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-1 w-full h-[90%]">
      {status?.map((job) => (
        <div key={job.job_id} className="bg-slate-800 p-4 rounded-lg w-full h-full">
          <h2 className="text-xl font-semibold">
            {job.job_type === "movie" ? "Movie" : "Show"}
          </h2>
          <p className="text-gray-300">Runs every {job.interval} hours</p>
          <p className="text-gray-300">
            Next run: {formatDate(job.next_run.toString())}
          </p>
        </div>
      ))}
    </div>
  );
}
