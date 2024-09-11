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
import FormInputField from "@/components/FormInputField";
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
  trakt_client_id: z.string(),
  trakt_client_secret: z.string(),
  ombi_url: z.string(),
  ombi_api_key: z.string(),
  radarr_url: z.string(),
  radarr_api_key: z.string(),
  sonarr_url: z.string(),
  sonarr_api_key: z.string(),
});

export default function Settings() {
  const context = useSetupStatus();

  const settingsForm = useForm<z.infer<typeof settingsFormSchema>>({
    resolver: zodResolver(settingsFormSchema),
    defaultValues: {
        ombi_or_sonarr_radarr: context.ombiEnabled ? "ombi" : "radarr-sonarr",
    }
  });

  const onSubmitSettings = (values: z.infer<typeof settingsFormSchema>) => {
    console.log("Settings:", values);
  };

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

        <div className="grid grid-cols-2 gap-4">
          <FormInputField
            form={settingsForm}
            name="trakt_client_id"
            label="Trakt client ID"
            placeholder="Trakt client ID here..."
          />

          <FormInputField
            form={settingsForm}
            name="trakt_client_secret"
            label="Trakt secret"
            placeholder="Trakt secret here..."
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <FormInputField
            form={settingsForm}
            name="ombi_url"
            label="Ombi base URL"
            placeholder="http://localhost:5000"
            disabled={settingsForm.getValues("ombi_or_sonarr_radarr") === "radarr-sonarr"}
          />

          <FormInputField
            form={settingsForm}
            name="ombi_api_key"
            label="Ombi API key"
            placeholder="1234567890987654321"
            disabled={settingsForm.getValues("ombi_or_sonarr_radarr") === "radarr-sonarr"}
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <FormInputField
            form={settingsForm}
            name="radarr_url"
            label="Radarr base URL"
            placeholder="http://localhost:7878"
            disabled={settingsForm.getValues("ombi_or_sonarr_radarr") === "ombi"}
          />

          <FormInputField
            form={settingsForm}
            name="radarr_api_key"
            label="Radarr API key"
            placeholder="1234567890987654321"
            disabled={settingsForm.getValues("ombi_or_sonarr_radarr") === "ombi"}
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <FormInputField
            form={settingsForm}
            name="sonarr_url"
            label="Sonarr base URL"
            placeholder="http://localhost:8989"
            disabled={settingsForm.getValues("ombi_or_sonarr_radarr") === "ombi"}
          />

          <FormInputField
            form={settingsForm}
            name="sonarr_api_key"
            label="Sonarr API key"
            placeholder="1234567890987654321"
            disabled={settingsForm.getValues("ombi_or_sonarr_radarr") === "ombi"}
          />
        </div>

        <Button type="submit">Save</Button>
      </form>
    </Form>
  );
}
