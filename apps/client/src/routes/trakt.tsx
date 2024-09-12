import * as React from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { TraktSettings } from "@/lib/types";
import Loading from "@/components/Loading";
import FormInputField from "@/components/FormInputField";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";

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

  const fetchSettings = async () => {
    try {
      const traktResponse = await fetch(
        `${import.meta.env.VITE_API_URL}/trakt/settings`
      );
      const traktData = await traktResponse.json();
      setTraktSettings(traktData);
      setLoading(false);
    } catch (error) {
      console.error("Error fetching settings:", error);
      setLoading(false);
    }
  };

  const onSubmitSettings = (values: z.infer<typeof formSchema>) => {
    console.log("Settings:", values);
  };

  React.useEffect(() => {
    fetchSettings();
  }, []); // Only fetch once on component mount

  React.useEffect(() => {
    // Update form values after data is fetched
    reset({
      trakt_client_id: traktSettings?.client_id ?? "",
      trakt_client_secret: traktSettings?.client_secret ?? "",
    });
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
            label="Trakt client ID"
            placeholder="Trakt client ID here..."
            isPassword
          />

          <FormInputField
            form={settingsForm}
            name="trakt_client_secret"
            label="Trakt secret"
            placeholder="Trakt secret here..."
            isPassword
          />
        </div>

        <Button type="submit">Save</Button>
      </form>
    </Form>
  );
}
