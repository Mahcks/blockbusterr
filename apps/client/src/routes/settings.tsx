import * as React from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";

import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Button } from "@/components/ui/button";
import { useSetupStatus } from "@/context/SetupContext";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

const settingsFormSchema = z.object({
  ombi_or_sonarr_radarr: z.enum(["ombi", "radarr-sonarr"]),
});

export default function Settings() {
  const context = useSetupStatus();

  const settingsForm = useForm<z.infer<typeof settingsFormSchema>>({
    resolver: zodResolver(settingsFormSchema),
    defaultValues: {
      ombi_or_sonarr_radarr: context.ombiEnabled ? "ombi" : "radarr-sonarr",
    },
  });

  const { reset } = settingsForm;

  const onSubmitSettings = (values: z.infer<typeof settingsFormSchema>) => {
    console.log("Settings:", values);
  };

  React.useEffect(() => {
    // Update form values after data is fetched
    reset({
      ombi_or_sonarr_radarr: context.ombiEnabled ? "ombi" : "radarr-sonarr",
    });
  }, [context.ombiEnabled, reset]);

  return (
    <Form {...settingsForm}>
      <form
        onSubmit={settingsForm.handleSubmit(onSubmitSettings)}
        className="space-y-4"
      >
        <FormField
          control={settingsForm.control}
          name="ombi_or_sonarr_radarr"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Select Mode</FormLabel>
              <Select onValueChange={field.onChange} defaultValue={field.value}>
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a mode to use..." />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value="ombi">Ombi</SelectItem>
                  <SelectItem value="radarr-sonarr">Radarr/Sonarr</SelectItem>
                </SelectContent>
              </Select>
              <FormDescription>
                {field.value === "ombi"
                  ? "Use Ombi for requesting movies and TV shows. (Please note this will disable the other mode.)"
                  : "Use Radarr/Sonarr for requesting movies and TV shows. (Please note this will disable the other mode.)"}
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <Button type="submit">Save</Button>
      </form>
    </Form>
  );
}
