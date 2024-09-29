import { Badge } from "@/components/ui/badge";
import { Log } from "@/types/log";
import { ColumnDef } from "@tanstack/react-table";

const levelColorMap: { [key in Log["level"]]: string } = {
  info: "bg-blue-500",
  warning: "bg-yellow-500",
  error: "bg-red-500",
  debug: "bg-gray-500",
};

export const columns: ColumnDef<Log>[] = [
  {
    header: "Timestamp",
    accessorKey: "timestamp",
    cell: ({ row }) => {
      const timestamp: string = row.getValue("timestamp");
      const date = new Date(timestamp);
      const formattedDate = date.toLocaleString();

      return <p>{formattedDate}</p>;
    },
  },
  {
    header: "Severity",
    accessorKey: "level",
    cell: ({ row }) => {
      const severity: Log["level"] = row.getValue("level");

      // Get the color class based on the severity level
      const colorClass = levelColorMap[severity];

      return (
        <Badge className={`${colorClass} text-white`}>
          {severity.toUpperCase()}
        </Badge>
      );
    },
  },
  {
    header: "Label",
    accessorKey: "label",
  },
  {
    header: "Message",
    accessorKey: "message",
  },
];
