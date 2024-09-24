import * as React from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import FormInputField from "@/components/FormInputField";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import Loading from "@/components/Loading";
import {
  RadarrQualityProfile,
  RadarrRootFolder,
  RadarrSettings,
} from "@/types/radarr";
import cronValidator from "cron-validate";

import { MovieSettings } from "@/types/movies";
import FormCronJobField from "@/components/FormCronJobField";

const cronExpressionSchema = z.string().refine(
  (value) => {
    const result = cronValidator(value, { preset: "default" });
    return result.isValid();
  },
  {
    message: "Invalid cron expression",
  }
);

// Validation schema
const movieFormSchema = z.object({
  anticipated: z.number(),
  cron_job_anticipated: cronExpressionSchema,
  box_office: z.number(),
  cron_job_box_office: cronExpressionSchema,
  popular: z.number(),
  cron_job_popular: cronExpressionSchema,
  trending: z.number(),
  cron_job_trending: cronExpressionSchema,
  min_runtime: z.number().optional(),
  max_runtime: z.number().optional(),
  min_year: z.number().optional(),
  max_year: z.number().optional(),
  allowed_countries: z.string(),
  allowed_languages: z.string(),
  blacklisted_genres: z.string(),
  blacklisted_title_keywords: z.string(),
  blacklisted_tmdb_ids: z.string(),

  // Radarr-specific fields
  api_key: z.string(),
  url: z.string(),
  root_folder: z.string().optional(),
  quality_profile: z.string().optional(),
  minimum_availability: z
    .enum(["announced", "inCinemas", "released"])
    .optional(),
});

export default function Radarr() {
  const [movieSettings, setMovieSettings] =
    React.useState<MovieSettings | null>(null);
  const [radarrSettings, setRadarrSettings] =
    React.useState<RadarrSettings | null>(null);

  // Keeps track of radarr profiles from the API
  const [radarrRootFolders, setRadarrRootFolder] = React.useState<
    RadarrRootFolder[]
  >([]);
  const [radarrQualityProfiles, setRadarrQualityProfiles] = React.useState<
    RadarrQualityProfile[]
  >([]);

  const [loading, setLoading] = React.useState(true);

  const fetchSettings = async () => {
    try {
      const movieResponse = await fetch(
        `${import.meta.env.VITE_API_URL}/movie/settings`
      );
      const movieData = await movieResponse.json();
      setMovieSettings(movieData);

      const radarrResponse = await fetch(
        `${import.meta.env.VITE_API_URL}/radarr/settings`
      );
      const radarrData = await radarrResponse.json();
      setRadarrSettings(radarrData);

      const rootFoldersResponse = await fetch(
        `${import.meta.env.VITE_API_URL}/radarr/rootfolders`
      );
      const rootFolderData = await rootFoldersResponse.json();
      setRadarrRootFolder(rootFolderData);

      const qualityProfilesResponse = await fetch(
        `${import.meta.env.VITE_API_URL}/radarr/profiles`
      );
      const qualityProfilesData = await qualityProfilesResponse.json();
      setRadarrQualityProfiles(qualityProfilesData);

      setLoading(false);
    } catch (error) {
      console.error("Error fetching settings:", error);
      setLoading(false);
    }
  };

  React.useEffect(() => {
    fetchSettings();
  }, []);

  const movieForm = useForm<z.infer<typeof movieFormSchema>>({
    resolver: zodResolver(movieFormSchema),
    defaultValues: {}, // Avoid validation issues with null
  });

  const { reset } = movieForm;

  React.useEffect(() => {
    if (movieSettings && radarrSettings) {
      const transformedDefaultValues = {
        ...movieSettings,
        allowed_countries: movieSettings.allowed_countries
          .map((c) => c.country_code)
          .join(", "),
        allowed_languages: movieSettings.allowed_languages
          .map((l) => l.language_code)
          .join(", "),
        blacklisted_genres: movieSettings.blacklisted_genres
          .map((g) => g.genre)
          .join(", "),
        blacklisted_title_keywords: movieSettings.blacklisted_title_keywords
          .map((k) => k.keyword)
          .join(", "),
        blacklisted_tmdb_ids: movieSettings.blacklisted_tmdb_ids
          .map((id) => id.tmdb_id.toString())
          .join(", "),
        api_key: radarrSettings.api_key ?? "",
        url: radarrSettings.base_url ?? "",
        root_folder: radarrSettings.root_folder?.toString() ?? "",
        quality_profile: radarrSettings.quality?.toString() ?? "",
        minimum_availability: radarrSettings.minimum_availability as
          | "announced"
          | "inCinemas"
          | "released"
          | undefined,
      };
      reset(transformedDefaultValues); // Reset the form with fetched values
    }
  }, [movieSettings, radarrSettings, reset]);

  const onSubmitMovie = async (values: z.infer<typeof movieFormSchema>) => {
    try {
      await fetch(`${import.meta.env.VITE_API_URL}/movie/settings`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          anticipated: values.anticipated,
          cron_job_anticipated: values.cron_job_anticipated,
          box_office: values.box_office,
          cron_job_box_office: values.cron_job_box_office,
          popular: values.popular,
          cron_job_popular: values.cron_job_popular,
          trending: values.trending,
          cron_job_trending: values.cron_job_trending,
          min_runtime: values.min_runtime,
          max_runtime: values.max_runtime,
          min_year: values.min_year,
          max_year: values.max_year,
          allowed_countries: values.allowed_countries,
          allowed_languages: values.allowed_languages,
          blacklisted_genres: values.blacklisted_genres,
          blacklisted_title_keywords: values.blacklisted_title_keywords,
          blacklisted_tmdb_ids: values.blacklisted_tmdb_ids,
        }),
      });

      await fetch(`${import.meta.env.VITE_API_URL}/radarr/settings`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          api_key: values.api_key,
          base_url: values.url,
          minimum_availability: values.minimum_availability,
          quality_profile: Number(values.quality_profile),
          root_folder: Number(values.root_folder),
        }),
      });
    } catch (error) {
      console.error("Error saving movie settings:", error);
    }
  };

  if (loading) {
    return <Loading />;
  }

  return (
    <>
      {movieSettings && (
        <Form {...movieForm}>
          <form
            onSubmit={movieForm.handleSubmit(onSubmitMovie)}
            className="space-y-8"
          >
            <div className="grid grid-cols-2 gap-4">
              <FormInputField
                form={movieForm}
                name="api_key"
                label="API key"
                placeholder="1234567890987654321"
                description="Provide your Radarr API key to make requests on behalf of Radarr."
                isPassword
              />

              <FormInputField
                form={movieForm}
                name="url"
                label="Base URL"
                placeholder="http://localhost:7878"
                description="The base URL of your Radarr instance."
              />

              <FormField
                control={movieForm.control}
                name="root_folder"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Select root folder</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select a root folder to use..." />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {radarrRootFolders.map((folder) => (
                          <SelectItem
                            key={folder.id}
                            value={folder.id.toString()}
                          >
                            {folder.path} (Free space:{" "}
                            {(folder.free_space / 1_073_741_824).toFixed(1)} GB)
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={movieForm.control}
                name="quality_profile"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Select quality profile</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select a quality profile to use..." />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {radarrQualityProfiles.map((qualityProfile) => (
                          <SelectItem
                            key={qualityProfile.id}
                            value={qualityProfile.id.toString()}
                          >
                            {qualityProfile.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={movieForm.control}
                name="minimum_availability"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Minimum Availability</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value ?? "released"} // Default to "released" if no value
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select availability" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="announced">Announced</SelectItem>
                        <SelectItem value="inCinemas">In Cinemas</SelectItem>
                        <SelectItem value="released">Released</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <Separator />

            <div className="grid grid-cols-2 gap-4">
              <FormInputField
                form={movieForm}
                name="anticipated"
                label="Anticipated Movies"
                placeholder="Enter # of anticipated movies"
                description="The number of movies to pull from the Trakt anticipated list."
                isNumber
              />
              <FormCronJobField
                form={movieForm}
                name="cron_job_anticipated"
                label="Cron Job Anticipated"
                placeholder="0 0 * * *"
                description="Schedule for fetching anticipated movies."
              />

              <FormInputField
                form={movieForm}
                name="box_office"
                label="Box Office Movies"
                placeholder="Enter # of box office movies"
                description="The number of movies to pull from the Trakt box office list."
                isNumber
              />
              <FormCronJobField
                form={movieForm}
                name="cron_job_box_office"
                label="Cron Job Box Office"
                placeholder="0 0 * * *"
                description="The cron job schedule for fetching box office movies."
              />

              <FormInputField
                form={movieForm}
                name="popular"
                label="Popular Movies"
                placeholder="Enter # of popular movies"
                description="The number of movies to pull from the Trakt popular list."
                isNumber
              />
              <FormCronJobField
                form={movieForm}
                name="cron_job_popular"
                label="Cron Job Popular"
                placeholder="0 0 * * *"
                description="The cron job schedule for fetching popular movies."
              />

              <FormInputField
                form={movieForm}
                name="trending"
                label="Trending Movies"
                placeholder="Enter # of trending movies"
                description="The number of movies to pull from the Trakt trending list."
                isNumber
              />
              <FormCronJobField
                form={movieForm}
                name="cron_job_trending"
                label="Cron Job Trending"
                placeholder="0 0 * * *"
                description="The cron job schedule for fetching trending movies."
              />
            </div>

            <Separator />

            <div className="grid grid-cols-2 gap-4">
              <FormInputField
                form={movieForm}
                name="min_runtime"
                label="Minimum Run Time"
                placeholder="Enter minimum runtime"
                description="The minimum length a movie can be in minutes. (0 for no limit)"
                isNumber
              />
              <FormInputField
                form={movieForm}
                name="max_runtime"
                label="Max Run Time"
                placeholder="Enter maximum runtime"
                description="The maximum length a movie can be in minutes. (0 for no limit)"
                isNumber
              />
              <FormInputField
                form={movieForm}
                name="min_year"
                label="Minimum Year"
                placeholder="Enter minimum year"
                description="The minimum year a movie can be released. (0 for no limit)"
                isNumber
              />
              <FormInputField
                form={movieForm}
                name="max_year"
                label="Maximum Year"
                placeholder="Enter maximum year"
                description="The maximum year a movie can be released. (0 for no limit)"
                isNumber
              />
            </div>

            <Separator />

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

            <Button type="submit">Save</Button>
          </form>
        </Form>
      )}
    </>
  );
}
