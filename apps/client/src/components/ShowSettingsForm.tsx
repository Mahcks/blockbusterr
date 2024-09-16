import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import FormInputField from "@/components/FormInputField";
import { Separator } from "@radix-ui/react-separator";
import { ShowSettings } from "@/types/shows";

// Validation schema
const showFormSchema = z.object({
  interval: z.number(),
  anticipated: z.number(),
  popular: z.number(),
  trending: z.number(),
  min_runtime: z.number(),
  max_runtime: z.number(),
  min_year: z.number(),
  max_year: z.number(),
  allowed_countries: z.string(),
  allowed_languages: z.string(),
  blacklisted_genres: z.string(),
  blacklisted_networks: z.string(),
  blacklisted_title_keywords: z.string(),
  blacklisted_tvdb_ids: z.string(),
});

interface ShowSettingsFormProps {
  defaultValues: ShowSettings; // Prop to pass in the fetched show settings
}

export default function ShowSettingsForm({
  defaultValues,
}: ShowSettingsFormProps) {
  const transformedDefaultValues = {
    ...defaultValues,
    allowed_countries: defaultValues.allowed_countries
      .map((c) => c.country_code)
      .join(", "), // Convert array of country codes into a comma-separated string
    allowed_languages: defaultValues.allowed_languages
      .map((l) => l.language_code)
      .join(", "), // Convert array of language codes into a comma-separated string
    blacklisted_genres: defaultValues.blacklisted_genres
      .map((g) => g.genre)
      .join(", "), // Convert array of genres into a comma-separated string
    blacklisted_networks: defaultValues.blacklisted_networks
      .map((n) => n.network)
      .join(", "), // Convert array of networks into a comma-separated string
    blacklisted_title_keywords: defaultValues.blacklisted_title_keywords
      .map((k) => k.keyword)
      .join(", "), // Convert array of keywords into a comma-separated string
    blacklisted_tvdb_ids: defaultValues.blacklisted_tvdb_ids
      .map((id) => id.tvdb_id.toString())
      .join(", "), // Convert array of TVDb IDs into a comma-separated string
  };

  const showForm = useForm<z.infer<typeof showFormSchema>>({
    resolver: zodResolver(showFormSchema),
    defaultValues: transformedDefaultValues, // Use the passed-in settings as default values
  });

  const onSubmitShow = (values: z.infer<typeof showFormSchema>) => {
    console.log("Show Settings:", values);
  };

  return (
    <Form {...showForm}>
      <form
        onSubmit={showForm.handleSubmit(onSubmitShow)}
        className="space-y-8"
      >
        <FormInputField
          form={showForm}
          name="interval"
          label="Interval (hours)"
          placeholder="Enter interval"
          description="Set the interval for pulling shows from all lists."
          isNumber
        />

        {/* 2x2 grid for Anticipated, Popular, Trending */}
        <div className="grid grid-cols-2 gap-4">
          <FormInputField
            form={showForm}
            name="anticipated"
            label="Anticipated Shows"
            placeholder="Enter # of anticipated shows"
            description="The number of shows to pull from the Trakt anticipated list."
            isNumber
          />
          <FormInputField
            form={showForm}
            name="popular"
            label="Popular Shows"
            placeholder="Enter # of popular shows"
            description="The number of shows to pull from the Trakt popular list."
            isNumber
         />
          <FormInputField
            form={showForm}
            name="trending"
            label="Trending Shows"
            placeholder="Enter # of trending shows"
            description="The number of shows to pull from the Trakt trending list."
            isNumber
          />
        </div>

        <Separator />

        {/* Filters: Runtime, Year */}
        <div className="grid grid-cols-2 gap-4">
          <FormInputField
            form={showForm}
            name="min_runtime"
            label="Minimum Run Time"
            placeholder="Enter minimum runtime"
            description="The minimum length an episode can be in minutes. (0 for no limit)"
            isNumber
          />
          <FormInputField
            form={showForm}
            name="max_runtime"
            label="Max Run Time"
            placeholder="Enter maximum runtime"
            description="The maximum length an episode can be in minutes. (0 for no limit)"
            isNumber
          />
          <FormInputField
            form={showForm}
            name="min_year"
            label="Minimum Year"
            placeholder="Enter minimum year"
            description="The minimum year a show can be released. (0 for no limit)"
            isNumber
          />
          <FormInputField
            form={showForm}
            name="max_year"
            label="Maximum Year"
            placeholder="Enter maximum year"
            description="The maximum year a show can be released. (0 for no limit)"
            isNumber
          />
        </div>

        <Separator />

        {/* Blacklist and Filters */}
        <div className="grid grid-cols-2 gap-4">
          <FormInputField
            form={showForm}
            name="allowed_countries"
            label="Allowed Countries"
            placeholder="e.g., US, GB"
            description="Comma-separated list of allowed countries."
          />
          <FormInputField
            form={showForm}
            name="allowed_languages"
            label="Allowed Languages"
            placeholder="e.g., en, es"
            description="Comma-separated list of allowed languages."
          />
          <FormInputField
            form={showForm}
            name="blacklisted_genres"
            label="Blacklisted Genres"
            placeholder="e.g., horror, anime"
            description="Comma-separated list of blacklisted genres."
          />
          <FormInputField
            form={showForm}
            name="blacklisted_title_keywords"
            label="Blacklisted Title Keywords"
            placeholder="e.g., Untitled, Barbie"
            description="Comma-separated list of blacklisted title keywords."
          />
          <FormInputField
            form={showForm}
            name="blacklisted_tvdb_ids"
            label="Blacklisted TVDb IDs"
            placeholder="e.g., 12345, 67890"
            description="Comma-separated list of blacklisted TVDb IDs."
          />
        </div>
        <Button type="submit">Save</Button>
      </form>
    </Form>
  );
}
