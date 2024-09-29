import { Log } from "@/types/log";
import * as React from "react";
import { DataTable } from "./data-table";
import { columns } from "./columns";

import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

export default function LogWidget() {
  const [logs, setLogs] = React.useState<Log[]>([]);

  const [levelFilter, setLevelFilter] = React.useState<string>("");
  const [searchTerm, setSearchTerm] = React.useState<string>("");

  const fetchLogs = React.useCallback(async () => {
    try {
      const params = new URLSearchParams();

      params.append("take", "30");
      params.append("skip", "0");

      if (levelFilter) {
        params.append("filter", levelFilter);
      }

      if (searchTerm) {
        params.append("search", searchTerm);
      }

      const response = await fetch(
        `${import.meta.env.VITE_API_URL}/logs?${params.toString()}`
      );
      const responseData = await response.json();

      if (Array.isArray(responseData)) {
        setLogs(responseData);
      } else {
        setLogs([]);
      }
    } catch (error) {
      console.error(error);
    }
  }, [levelFilter, searchTerm]);

  React.useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  return (
    <div className="w-full h-full flex flex-col">
      <div className="flex-none flex items-center mb-4 space-x-4">
        {/* Filter Controls */}
        <Select
          value={levelFilter}
          onValueChange={(value) => setLevelFilter(value)}
        >
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="All Levels" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Levels</SelectItem>
            <SelectItem value="info">Info</SelectItem>
            <SelectItem value="warning">Warning</SelectItem>
            <SelectItem value="error">Error</SelectItem>
            <SelectItem value="debug">Debug</SelectItem>
          </SelectContent>
        </Select>

        <Input
          type="text"
          placeholder="Search logs..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="flex-grow"
        />
      </div>

      {/* Data Table */}
      <div className="flex-1 overflow-hidden">
        <DataTable columns={columns} data={logs} />
      </div>
    </div>
  );
}
