import Loading from "@/components/Loading";
import * as React from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import FormInputField from "@/components/FormInputField";
import { Separator } from "@/components/ui/separator";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ShowSettings } from "@/types/shows";
import { SonarrSettings, SonarrRootFolder, SonarrQualityProfile } from "@/types/sonarr";

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

  // Sonarr settings
  api_key: z.string(),
  url: z.string(),
  root_folder: z.string().optional(),
  quality_profile: z.string(),
});

export default function Sonarr() {
  const [showSettings, setShowSettings] = React.useState<ShowSettings | null>(
    null
  );
  const [sonarrSettings, setSonarrSettings] =
    React.useState<SonarrSettings | null>(null);
  const [loading, setLoading] = React.useState(true);

  const [sonarrRootFolders, setSonarrRootFolder] = React.useState<
    SonarrRootFolder[]
  >([]);
  const [sonarrQualityProfiles, setSonarrQualityProfiles] = React.useState<
    SonarrQualityProfile[]
  >([]);

  const fetchSettings = async () => {
    try {
      const showResponse = await fetch(
        `${import.meta.env.VITE_API_URL}/show/settings`
      );
      const showData = await showResponse.json();
      setShowSettings(showData);

      const sonarrResponse = await fetch(
        `${import.meta.env.VITE_API_URL}/sonarr/settings`
      );
      const sonarrData = await sonarrResponse.json();
      setSonarrSettings(sonarrData);

      const rootFoldersResponse = await fetch(
        `${import.meta.env.VITE_API_URL}/sonarr/rootfolders`
      );
      const rootFolderData = await rootFoldersResponse.json();
      setSonarrRootFolder(rootFolderData);

      const qualityProfilesResponse = await fetch(
        `${import.meta.env.VITE_API_URL}/sonarr/profiles`
      );
      const qualityProfilesData = await qualityProfilesResponse.json();
      setSonarrQualityProfiles(qualityProfilesData);

      setLoading(false); // Both requests have completed
    } catch (error) {
      console.error("Error fetching settings:", error);
      setLoading(false);
    }
  };

  React.useEffect(() => {
    fetchSettings();
  }, []);

  const showForm = useForm<z.infer<typeof showFormSchema>>({
    resolver: zodResolver(showFormSchema),
    defaultValues: {}, // Avoid validation issues with null
  });

  const { reset } = showForm;

  React.useEffect(() => {
    if (showSettings && sonarrSettings) {
      const transformedDefaultValues = {
        ...showSettings,
        allowed_countries: showSettings.allowed_countries
          .map((c) => c.country_code)
          .join(", "), // Convert array of country codes into a comma-separated string
        allowed_languages: showSettings.allowed_languages
          .map((l) => l.language_code)
          .join(", "), // Convert array of language codes into a comma-separated string
        blacklisted_genres: showSettings.blacklisted_genres
          .map((g) => g.genre)
          .join(", "), // Convert array of genres into a comma-separated string
        blacklisted_networks: showSettings.blacklisted_networks
          .map((n) => n.network)
          .join(", "), // Convert array of networks into a comma-separated string
        blacklisted_title_keywords: showSettings.blacklisted_title_keywords
          .map((k) => k.keyword)
          .join(", "), // Convert array showSettings keywords into a comma-separated string
        blacklisted_tvdb_ids: showSettings.blacklisted_tvdb_ids
          .map((id) => id.tvdb_id.toString())
          .join(", "), // Convert array of TVDb IDs into a comma-separated string
        api_key: sonarrSettings.api_key ?? "",
        url: sonarrSettings.url ?? "",
        root_folder: sonarrSettings.root_folder?.toString() ?? "",
        quality_profile: sonarrSettings.quality?.toString() ?? "",
      };
      reset(transformedDefaultValues); // Reset the form with fetched values
    }
  }, [showSettings, sonarrSettings, reset]);

  const onSubmitShow = (values: z.infer<typeof showFormSchema>) => {
    console.log("Movie Settings:", values);
  };

  if (loading) {
    return <Loading />;
  }

  console.log(sonarrSettings);

  return (
    <Form {...showForm}>
      <form
        onSubmit={showForm.handleSubmit(onSubmitShow)}
        className="space-y-8"
      >
        <div className="grid grid-cols-2 gap-4">
          <FormInputField
            form={showForm}
            name="api_key"
            label="API key"
            placeholder="1234567890987654321"
            description="Provide your Sonarr API key to make requests on behalf of Sonarr."
            isPassword
          />

          <FormInputField
            form={showForm}
            name="url"
            label="Base URL"
            placeholder="http://localhost:8989"
            description="The base URL of your Radarr instance."
          />

          <FormField
            control={showForm.control}
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
                    {sonarrRootFolders.map((folder) => (
                      <SelectItem key={folder.id} value={folder.id.toString()}>
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
            control={showForm.control}
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
                    {sonarrQualityProfiles.map((qualityProfile) => (
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
        </div>

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
