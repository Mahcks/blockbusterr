import { MovieSettings } from "@/lib/types";
import { zodResolver } from "@hookform/resolvers/zod";
import { Separator } from "@radix-ui/react-select";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import FormInputField from "@/components/FormInputField";

// Validation schema
const movieFormSchema = z.object({
  interval: z.number(),
  anticipated: z.number(),
  boxOffice: z.number(),
  popular: z.number(),
  trending: z.number(),
  min_runtime: z.number(),
  max_runtime: z.number(),
  min_year: z.number(),
  max_year: z.number(),
  allowed_countries: z.string(),
  allowed_languages: z.string(),
  blacklisted_genres: z.string(),
  blacklisted_title_keywords: z.string(),
  blacklisted_tmdb_ids: z.string(),
});

interface MovieSettingsFormProps {
  defaultValues: MovieSettings; // Prop to pass in the fetched movie settings
}

export default function MovieSettingsForm({
  defaultValues,
}: MovieSettingsFormProps) {
  const transformedDefaultValues = {
    ...defaultValues,
    allowed_countries: defaultValues.allowed_countries
      .map((c) => c.country_code)
      .join(", "), // Convert the array of country codes into a comma-separated string
    allowed_languages: defaultValues.allowed_languages
      .map((l) => l.language_code)
      .join(", "), // Convert the array of language codes into a comma-separated string
    blacklisted_genres: defaultValues.blacklisted_genres
      .map((g) => g.genre)
      .join(", "), // Convert the array of genres into a comma-separated string
    blacklisted_title_keywords: defaultValues.blacklisted_title_keywords
      .map((k) => k.keyword)
      .join(", "), // Convert the array of keywords into a comma-separated string
    blacklisted_tmdb_ids: defaultValues.blacklisted_tmdb_ids
      .map((id) => id.tmdb_id.toString())
      .join(", "), // Convert the array of TMDb IDs into a comma-separated string
  };

  const movieForm = useForm<z.infer<typeof movieFormSchema>>({
    resolver: zodResolver(movieFormSchema),
    defaultValues: transformedDefaultValues
  });

  const onSubmitMovie = (values: z.infer<typeof movieFormSchema>) => {
    console.log("Movie Settings:", values);
  };

  return (
    <Form {...movieForm}>
      <form
        onSubmit={movieForm.handleSubmit(onSubmitMovie)}
        className="space-y-8"
      >
        <FormInputField
          form={movieForm}
          name="interval"
          label="Interval (hours)"
          placeholder="Enter interval"
          description="Set the interval for pulling movies from all lists."
        />

        {/* 2x2 grid for Anticipated, Box Office, Popular, Trending */}
        <div className="grid grid-cols-2 gap-4">
          <FormInputField
            form={movieForm}
            name="anticipated"
            label="Anticipated Movies"
            placeholder="Enter # of anticipated movies"
            description="The number of movies to pull from the Trakt anticipated list."
          />
          <FormInputField
            form={movieForm}
            name="boxOffice"
            label="Box Office Movies"
            placeholder="Enter # of box office movies"
            description="The number of movies to pull from the Trakt box office list."
          />
          <FormInputField
            form={movieForm}
            name="popular"
            label="Popular Movies"
            placeholder="Enter # of popular movies"
            description="The number of movies to pull from the Trakt popular list."
          />
          <FormInputField
            form={movieForm}
            name="trending"
            label="Trending Movies"
            placeholder="Enter # of trending movies"
            description="The number of movies to pull from the Trakt trending list."
          />
        </div>

        <Separator />

        {/* Filters: Runtime, Year */}
        <div className="grid grid-cols-2 gap-4">
          <FormInputField
            form={movieForm}
            name="min_runtime"
            label="Minimum Run Time"
            placeholder="Enter minimum runtime"
            description="The minimum length a movie can be in minutes. (0 for no limit)"
          />
          <FormInputField
            form={movieForm}
            name="max_runtime"
            label="Max Run Time"
            placeholder="Enter maximum runtime"
            description="The maximum length a movie can be in minutes. (0 for no limit)"
          />
          <FormInputField
            form={movieForm}
            name="min_year"
            label="Minimum Year"
            placeholder="Enter minimum year"
            description="The minimum year a movie can be released. (0 for no limit)"
          />
          <FormInputField
            form={movieForm}
            name="max_year"
            label="Maximum Year"
            placeholder="Enter maximum year"
            description="The maximum year a movie can be released. (0 for no limit)"
          />
        </div>

        <Separator />

        {/* Blacklist and Filters */}
        <div className="grid grid-cols-2 gap-4">
          <FormInputField
            form={movieForm}
            name="allowed_countries"
            label="Allowed Countries"
            placeholder="e.g., US, GB"
            description="Comma-separated list of allowed countries."
          />
          <FormInputField
            form={movieForm}
            name="allowed_languages"
            label="Allowed Languages"
            placeholder="e.g., en, es"
            description="Comma-separated list of allowed languages."
          />
          <FormInputField
            form={movieForm}
            name="blacklisted_genres"
            label="Blacklisted Genres"
            placeholder="e.g., horror, anime"
            description="Comma-separated list of blacklisted genres."
          />
          <FormInputField
            form={movieForm}
            name="blacklisted_title_keywords"
            label="Blacklisted Title Keywords"
            placeholder="e.g., Untitled, Barbie"
            description="Comma-separated list of blacklisted title keywords."
          />
          <FormInputField
            form={movieForm}
            name="blacklisted_tmdb_ids"
            label="Blacklisted TMDb IDs"
            placeholder="e.g., 12345, 67890"
            description="Comma-separated list of blacklisted TMDb IDs."
          />
        </div>
        <Button type="submit">Submit</Button>
      </form>
    </Form>
  );
}
