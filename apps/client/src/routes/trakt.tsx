import * as React from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { TraktSettings } from "@/types/trakt";
import Loading from "@/components/Loading";
import FormInputField from "@/components/FormInputField";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";

// Schema for form validation
const formSchema = z.object({
  trakt_client_id: z.string(),
  trakt_client_secret: z.string(),
});

export default function Trakt() {
  const [traktSettings, setTraktSettings] =
    React.useState<TraktSettings | null>(null);
  const [loading, setLoading] = React.useState(true);

  const settingsForm = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
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
    } catch (error) {
      console.error("Error fetching settings:", error);
    } finally {
      setLoading(false);
    }
  };

  // Function to save Trakt settings via API
  const saveSettings = async (values: z.infer<typeof formSchema>) => {
    try {
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
      await fetchSettings(); // Refresh settings after saving
    } catch (error) {
      console.error("Error saving settings:", error);
    }
  };

  // Handles form submission
  const onSubmitSettings = async (values: z.infer<typeof formSchema>) => {
    await saveSettings(values);
    console.log("Settings saved:", values);
  };

  React.useEffect(() => {
    fetchSettings();
  }, []); // Fetch settings once on component mount

  React.useEffect(() => {
    // Set form values after fetching settings
    if (traktSettings) {
      reset({
        trakt_client_id: traktSettings.client_id || "",
        trakt_client_secret: traktSettings.client_secret || "",
      });
    }
  }, [reset, traktSettings]);

  if (loading) {
    return <Loading />;
  }

  return (
    <Form {...settingsForm}>
      <form
        onSubmit={settingsForm.handleSubmit(onSubmitSettings)}
        className="space-y-4"
      >
        <div className="grid grid-cols-2 gap-4">
          <FormInputField
            form={settingsForm}
            name="trakt_client_id"
            label="Trakt Client ID"
            placeholder="Enter Trakt Client ID"
            isPassword
          />

          <FormInputField
            form={settingsForm}
            name="trakt_client_secret"
            label="Trakt Client Secret"
            placeholder="Enter Trakt Client Secret"
            isPassword
          />
        </div>

        <Button type="submit">Save</Button>
      </form>
    </Form>
  );
}
