
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import * as React from "react";

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
import { RadarrQualityProfile, RadarrRootFolder } from "@/types/radarr";
import { MovieSettings } from "@/types/movies";

// Validation schema
const movieFormSchema = z.object({
  interval: z.number(),
  anticipated: z.number(),
  box_office: z.number(),
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

  // Radarr-specific fields
  root_folder: z.string().optional(), // These are optional to avoid validation issues for Ombi
  quality_profile: z.string().optional(),
});

interface MovieSettingsFormProps {
  defaultValues: MovieSettings; // Prop to pass in the fetched movie settings
  isRadarr: boolean; // Flag to determine if Radarr-specific fields should be shown
}

export default function MovieSettingsForm({
  defaultValues,
  isRadarr = false,
}: MovieSettingsFormProps) {
  // Keeps track of radarr profiles from the API
  const [radarrRootFolders, setRadarrRootFolder] = React.useState<
    RadarrRootFolder[]
  >([]);
  // Fetch the quality profiles for radarr
  const [radarrQualityPorfiles, setRadarrQualityProfiles] = React.useState<
    RadarrQualityProfile[]
  >([]);
  const [loading, setLoading] = React.useState(false);

  const fetchRadarrData = async () => {
    try {
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
      console.error("Failed to fetch Radarr root folders:", error);
      setLoading(false);
    }
  };

  React.useEffect(() => {
    fetchRadarrData();
  }, []);

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
    defaultValues: transformedDefaultValues,
  });

  const onSubmitMovie = (values: z.infer<typeof movieFormSchema>) => {
    console.log("Movie Settings:", values);
  };

  if (loading) {
    <Loading />;
  }

  return (
    <Form {...movieForm}>
      <form
        onSubmit={movieForm.handleSubmit(onSubmitMovie)}
        className="space-y-8"
      >
        {/* Radarr-Specific Fields */}
        {isRadarr && (
          <div className="grid grid-cols-2 gap-4">
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
                  {/*  <FormDescription></FormDescription> */}
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
                      {radarrQualityPorfiles.map((qualityProfile) => (
                        <SelectItem
                          key={qualityProfile.id}
                          value={qualityProfile.id.toString()}
                        >
                          {qualityProfile.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {/*  <FormDescription></FormDescription> */}
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        )}

        <Separator />

        <FormInputField
          form={movieForm}
          name="interval"
          label="Interval (hours)"
          placeholder="Enter interval"
          description="Set the interval for pulling movies from all lists."
          isNumber
        />

        {/* 2x2 grid for Anticipated, Box Office, Popular, Trending */}
        <div className="grid grid-cols-2 gap-4">
          <FormInputField
            form={movieForm}
            name="anticipated"
            label="Anticipated Movies"
            placeholder="Enter # of anticipated movies"
            description="The number of movies to pull from the Trakt anticipated list."
            isNumber
          />
          <FormInputField
            form={movieForm}
            name="box_office"
            label="Box Office Movies"
            placeholder="Enter # of box office movies"
            description="The number of movies to pull from the Trakt box office list."
            isNumber
          />
          <FormInputField
            form={movieForm}
            name="popular"
            label="Popular Movies"
            placeholder="Enter # of popular movies"
            description="The number of movies to pull from the Trakt popular list."
            isNumber
          />
          <FormInputField
            form={movieForm}
            name="trending"
            label="Trending Movies"
            placeholder="Enter # of trending movies"
            description="The number of movies to pull from the Trakt trending list."
            isNumber
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
        <Button type="submit">Save</Button>
      </form>
    </Form>
  );
}
