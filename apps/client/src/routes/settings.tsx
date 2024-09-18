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
import Loading from "@/components/Loading";
import FormInputField from "@/components/FormInputField";
import { Separator } from "@/components/ui/separator";

// Zod schema definition, adaptable to add more modes
const settingsFormSchema = z.object({
  mode: z.enum(["ombi", "radarr-sonarr"]), // Add new modes here
  trakt_client_id: z.string().optional(),
  trakt_client_secret: z.string().optional(),
  omdb_api_key: z.string().optional(),
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
  const [loading, setLoading] = React.useState(true);
  const [traktSettings, setTraktSettings] = React.useState<{
    client_id: string;
    client_secret: string;
  } | null>(null);
  const [omdbApiKey, setOmdbApiKey] = React.useState<string | null>(null);

  const settingsForm = useForm<z.infer<typeof settingsFormSchema>>({
    resolver: zodResolver(settingsFormSchema),
    defaultValues: {
      mode: context.mode ?? "ombi", // Use context mode or default to "ombi"
      trakt_client_id: "",
      trakt_client_secret: "",
    },
  });

  const { reset } = settingsForm;

  // Function to fetch Trakt settings from API
  const fetchSettings = async () => {
    try {
      const response = await fetch(
        `${import.meta.env.VITE_API_URL}/trakt/settings`
      );
      const data = await response.json();
      setTraktSettings(data);

      const omdbResponse = await fetch(
        `${import.meta.env.VITE_API_URL}/omdb/settings`
      );
      const omdbData = await omdbResponse.json();
      setOmdbApiKey(omdbData.api_key);
    } catch (error) {
      console.error("Error fetching settings:", error);
    } finally {
      setLoading(false);
    }
  };

  // Function to save Trakt settings via API
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

      await fetch(`${import.meta.env.VITE_API_URL}/trakt/settings`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          client_id: values.trakt_client_id,
          client_secret: values.trakt_client_secret,
        }),
      });

      await fetch(`${import.meta.env.VITE_API_URL}/omdb/settings`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          api_key: values.omdb_api_key,
        }),
      });

      await fetchSettings(); // Refresh settings after saving
      await context.checkMode();
    } catch (error) {
      console.error("Error saving settings:", error);
    }
  };

  // Handles form submission
  const onSubmitSettings = async (
    values: z.infer<typeof settingsFormSchema>
  ) => {
    await saveSettings(values);
    console.log("Settings saved:", values);
  };

  // Fetch settings once on component mount
  React.useEffect(() => {
    fetchSettings();
  }, []);

  // Set form values after fetching Trakt settings
  React.useEffect(() => {
    if (traktSettings) {
      reset({
        mode: context.mode ?? "ombi",
        trakt_client_id: traktSettings.client_id || "",
        trakt_client_secret: traktSettings.client_secret || "",
        omdb_api_key: omdbApiKey || "",
      });
    }
  }, [reset, traktSettings, context.mode, omdbApiKey]);

  const selectedMode = settingsForm.watch("mode") as Mode; // Get the selected mode

  if (loading) {
    return <Loading />;
  }

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

        <Separator />

        {/* Trakt Settings */}
        <div className="grid grid-cols-2 gap-4">
          {/* Trakt Client ID */}
          <FormInputField
            form={settingsForm}
            name="trakt_client_id"
            label="Trakt Client ID"
            placeholder="Enter Trakt Client ID"
            isPassword
          />

          {/* Trakt Client Secret */}
          <FormInputField
            form={settingsForm}
            name="trakt_client_secret"
            label="Trakt Client Secret"
            placeholder="Enter Trakt Client Secret"
            isPassword
          />

          <FormInputField
            form={settingsForm}
            name="omdb_api_key"
            label="OMDb API Key"
            placeholder="Enter your OMDb API key..."
            isPassword
          />
        </div>

        <Button type="submit">Save</Button>
      </form>
    </Form>
  );
}
