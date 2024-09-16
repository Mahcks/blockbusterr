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

// Zod schema definition, adaptable to add more modes
const settingsFormSchema = z.object({
  mode: z.enum(["ombi", "radarr-sonarr"]), // Add new modes here
});

// Define mode-specific behavior and description
const modes = {
  ombi: {
    label: "Ombi",
    description:
      "Use Ombi for requesting movies and TV shows. (Please note this will disable the other mode, but will save the settings.)",
  },
  "radarr-sonarr": {
    label: "Radarr/Sonarr",
    description:
      "Use Radarr/Sonarr for requesting movies and TV shows. (Please note this will disable the other mode, but will save the settings.)",
  },
};

// Extendable to support more modes
type Mode = keyof typeof modes;

export default function Settings() {
  const context = useSetupStatus(); // Get context values

  const settingsForm = useForm<z.infer<typeof settingsFormSchema>>({
    resolver: zodResolver(settingsFormSchema),
    defaultValues: {
      mode: context.mode ?? "ombi", // Use context mode or default to "ombi"
    },
  });

  const { reset } = settingsForm;

  // Save settings via API
  const saveSettings = async (values: z.infer<typeof settingsFormSchema>) => {
    try {
      await fetch(`${import.meta.env.VITE_API_URL}/settings`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          key: "MODE",
          value: values.mode,
        }),
      });
      await context.checkMode(); // Refresh mode in context after saving
    } catch (error) {
      console.error("Error saving settings:", error);
    }
  };

  // Form submission handler
  const onSubmitSettings = async (
    values: z.infer<typeof settingsFormSchema>
  ) => {
    await saveSettings(values); // Call saveSettings function
    console.log("Settings saved:", values);
  };

  // Reset form values after fetching the context data
  React.useEffect(() => {
    if (context.mode !== null) {
      reset({
        mode: context.mode,
      });
    }
  }, [context.mode, reset]);

  const selectedMode = settingsForm.watch("mode") as Mode; // Get the selected mode

  return (
    <Form {...settingsForm}>
      <form
        onSubmit={settingsForm.handleSubmit(onSubmitSettings)}
        className="space-y-4"
      >
        {/* Select the Mode */}
        <FormField
          control={settingsForm.control}
          name="mode" // Use 'mode' consistently
          render={({ field }) => (
            <FormItem>
              <FormLabel>Select Mode</FormLabel>
              <Select onValueChange={field.onChange} value={field.value}>
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder="Select a mode to use..." />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  {Object.keys(modes).map((key) => (
                    <SelectItem key={key} value={key}>
                      {modes[key as Mode].label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FormDescription>
                {modes[selectedMode]?.description}
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
