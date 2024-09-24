// components/FormCronJobField.tsx
import {
  FormItem,
  FormLabel,
  FormControl,
  FormDescription,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { UseFormReturn } from "react-hook-form";
import { CalendarIcon } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import * as React from "react";

interface FormCronJobFieldProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: UseFormReturn<any>;
  name: string;
  label: string;
  placeholder: string;
  description?: string;
}

const CronJobHelper = ({
  value,
  onSelect,
}: {
  value: string;
  onSelect: (cronExpression: string) => void;
}) => {
  const [dialogOpen, setDialogOpen] = React.useState(false);
  const [minute, setMinute] = React.useState("0");
  const [hour, setHour] = React.useState("0");
  const [dayOfMonth, setDayOfMonth] = React.useState("*");
  const [month, setMonth] = React.useState("*");
  const [dayOfWeek, setDayOfWeek] = React.useState("*");

  React.useEffect(() => {
    if (value) {
      const parts = value.trim().split(/\s+/);
      if (parts.length === 5) {
        setMinute(parts[0]);
        setHour(parts[1]);
        setDayOfMonth(parts[2]);
        setMonth(parts[3]);
        setDayOfWeek(parts[4]);
      }
    }
  }, [value]);

  const generateCronExpression = () => {
    const cronExpression = `${minute} ${hour} ${dayOfMonth} ${month} ${dayOfWeek}`;
    onSelect(cronExpression);
    setDialogOpen(false);
  };

  return (
    <Dialog open={dialogOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="icon" className="ml-2" onClick={() => setDialogOpen(true)}>
          <CalendarIcon className="h-4 w-4" />
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Create Cron Job Schedule</DialogTitle>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          {[
            {
              label: "Minute",
              value: minute,
              setter: setMinute,
              placeholder: "0-59",
            },
            {
              label: "Hour",
              value: hour,
              setter: setHour,
              placeholder: "0-23",
            },
            {
              label: "Day of Month",
              value: dayOfMonth,
              setter: setDayOfMonth,
              placeholder: "1-31",
            },
            {
              label: "Month",
              value: month,
              setter: setMonth,
              placeholder: "1-12",
            },
            {
              label: "Day of Week",
              value: dayOfWeek,
              setter: setDayOfWeek,
              placeholder: "0-6",
            },
          ].map((field, index) => (
            <div key={index} className="grid grid-cols-5 items-center gap-4">
              <Label htmlFor={field.label} className="text-right">
                {field.label}
              </Label>
              <Input
                id={field.label}
                value={field.value}
                onChange={(e) => field.setter(e.target.value)}
                placeholder={field.placeholder}
                className="col-span-4"
              />
            </div>
          ))}
        </div>
        <Button onClick={generateCronExpression}>Set Cron Job</Button>
      </DialogContent>
    </Dialog>
  );
};

const FormCronJobField: React.FC<FormCronJobFieldProps> = ({
  form,
  name,
  label,
  placeholder,
  description,
}) => {
  const value = form.watch(name);

  return (
    <FormItem>
      <FormLabel>{label}</FormLabel>
      <div className="flex items-center">
        <FormControl>
          <Input
            placeholder={placeholder}
            {...form.register(name)}
            value={value ?? ""}
          />
        </FormControl>
        <CronJobHelper
          value={value}
          onSelect={(value) => form.setValue(name, value)}
        />
      </div>
      {description && <FormDescription>{description}</FormDescription>}
      <FormMessage />
    </FormItem>
  );
};

export default FormCronJobField;
