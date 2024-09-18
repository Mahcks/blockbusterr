"use client";

import { useState, useEffect } from "react";
import { Clock, Film, Tv } from "lucide-react";
import { Separator } from "@/components/ui/separator";
import { JobStatus } from "@/types/job";

type GroupedJobs = {
  [key: string]: {
    [key: string]: string;
  };
};

export default function VerticalJobStatusWidget() {
  const [jobStatus, setJobStatus] = useState<JobStatus[] | null>(null);

  useEffect(() => {
    const fetchJobStatus = async () => {
      try {
        const response = await fetch(
          `${import.meta.env.VITE_API_URL}/jobs/status`
        );
        if (!response.ok) throw new Error("Failed to fetch job status");
        const data = await response.json();
        setJobStatus(data);
      } catch (error) {
        console.error("Error fetching job status:", error);
        setJobStatus(null);
      }
    };

    fetchJobStatus();
  }, []);

  const groupJobs = (jobs: JobStatus[]): GroupedJobs => {
    return jobs.reduce((acc, job) => {
      const [mediaType, jobType] = job.job_type.split("-");
      if (!acc[mediaType]) acc[mediaType] = {};
      acc[mediaType][jobType] = new Date(job.next_run).toLocaleString();
      return acc;
    }, {} as GroupedJobs);
  };

  if (!jobStatus)
    return <div className="text-center p-4">Loading job statuses...</div>;

  const groupedJobs = groupJobs(jobStatus);

  return (
    <div className="w-[full] max-w-md mx-auto h-[250px] bg-inherit px-5">
        {Object.entries(groupedJobs).map(([mediaType, jobs], index, array) => (
          <div key={mediaType}>
            <div className="flex items-center space-x-2 mb-1">
              {mediaType === "movie" ? (
                <Film className="w-5 h-5" />
              ) : (
                <Tv className="w-5 h-5" />
              )}
              <h3 className="text-lg font-semibold capitalize">{mediaType}s</h3>
            </div>
            <div className="space-y-1 ml-7 mb-2">
              {Object.entries(jobs).map(([jobType, nextRun]) => (
                <div
                  key={jobType}
                  className="flex items-center justify-between text-sm"
                >
                  <span className="capitalize">{jobType}</span>
                  <span className="flex items-center text-muted-foreground">
                    <Clock className="w-4 h-4 mr-1" />
                    {nextRun}
                  </span>
                </div>
              ))}
            </div>
            {index < array.length - 1 && <Separator className="my-3" />}
          </div>
        ))}
    </div>
  );
}
