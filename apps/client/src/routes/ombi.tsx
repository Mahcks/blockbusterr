import { MovieSettings, ShowSettings } from "@/lib/types";
import * as React from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { Separator } from "@radix-ui/react-select";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import Loading from "@/components/Loading";
import FormInputField from "@/components/FormInputField";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

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

export default function Ombi() {
  const [movieSettings, setMovieSettings] =
    React.useState<MovieSettings | null>(null);

  const [showSettings, setShowSettings] = React.useState<ShowSettings | null>(
    null
  );

  const movieForm = useForm<z.infer<typeof movieFormSchema>>({
    resolver: zodResolver(movieFormSchema),
    defaultValues: {
      interval: 0,
      anticipated: 0,
      boxOffice: 0,
      popular: 0,
      trending: 0,
      min_runtime: 0,
      max_runtime: 0,
      min_year: 0,
      max_year: 0,
      allowed_countries: "",
      allowed_languages: "",
      blacklisted_genres: "",
      blacklisted_title_keywords: "",
      blacklisted_tmdb_ids: "",
    },
  });

  const showForm = useForm<z.infer<typeof showFormSchema>>({
    resolver: zodResolver(showFormSchema),
    defaultValues: {
      interval: 0,
      anticipated: 0,
      popular: 0,
      trending: 0,
      min_runtime: 0,
      max_runtime: 0,
      min_year: 0,
      max_year: 0,
      allowed_countries: "",
      allowed_languages: "",
      blacklisted_genres: "",
      blacklisted_networks: "",
      blacklisted_title_keywords: "",
      blacklisted_tvdb_ids: "",
    },
  });

  const { reset: resetMovieForm } = movieForm;
  const { reset: resetShowForm } = showForm;

  // Submit handlers
  const onSubmitMovie = (values: z.infer<typeof movieFormSchema>) => {
    console.log("Movie Settings:", values);
  };

  const onSubmitShow = (values: z.infer<typeof showFormSchema>) => {
    console.log("Show Settings:", values);
  };

  // Fetch movie settings
  const getMovieSettings = async () => {
    try {
      const res = await fetch(`${import.meta.env.VITE_API_URL}/movie/settings`);
      const data = await res.json();
      setMovieSettings(data);
    } catch (error) {
      console.error("Error fetching movie settings", error);
    }
  };

  // Fetch show settings
  const getShowSettings = async () => {
    try {
      const res = await fetch(`${import.meta.env.VITE_API_URL}/show/settings`);
      const data = await res.json();
      setShowSettings(data);
    } catch (error) {
      console.error("Error fetching show settings", error);
    }
  };

  // Fetch the settings when the component mounts.
  React.useEffect(() => {
    getMovieSettings();
    getShowSettings();
  }, []);

  // Reset form values when movieSettings and showSettings are fetched
  React.useEffect(() => {
    if (movieSettings) {
      resetMovieForm({
        interval: movieSettings.interval || 0,
        anticipated: movieSettings.anticipated || 0,
        boxOffice: movieSettings.box_office || 0,
        popular: movieSettings.popular || 0,
        trending: movieSettings.trending || 0,
        max_runtime: movieSettings.max_runtime || 0,
        min_runtime: movieSettings.min_runtime || 0,
        min_year: movieSettings.min_year || 0,
        max_year: movieSettings.max_year || 0,
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
      });
    }

    if (showSettings) {
      resetShowForm({
        interval: showSettings.interval || 0,
        anticipated: showSettings.anticipated || 0,
        popular: showSettings.popular || 0,
        trending: showSettings.trending || 0,
        max_runtime: showSettings.max_runtime || 0,
        min_runtime: showSettings.min_runtime || 0,
        min_year: showSettings.min_year || 0,
        max_year: showSettings.max_year || 0,
        allowed_countries: showSettings.allowed_countries
          .map((c) => c.country_code)
          .join(", "),
        allowed_languages: showSettings.allowed_languages
          .map((l) => l.language_code)
          .join(", "),
        blacklisted_genres: showSettings.blacklisted_genres
          .map((g) => g.genre)
          .join(", "),
        blacklisted_networks: showSettings.blacklisted_networks
          .map((n) => n.network)
          .join(", "),
        blacklisted_title_keywords: showSettings.blacklisted_title_keywords
          .map((k) => k.keyword)
          .join(", "),
        blacklisted_tvdb_ids: showSettings.blacklisted_tvdb_ids
          .map((id) => id.tvdb_id.toString())
          .join(", "),
      });
    }
  }, [movieSettings, showSettings, resetMovieForm, resetShowForm]);

  if (!movieSettings || !showSettings) {
    return <Loading />;
  }

  return (
    <div>
      <Tabs defaultValue="movies">
        <TabsList>
          <TabsTrigger value="movies">Movies</TabsTrigger>
          <TabsTrigger value="tv">TV Shows</TabsTrigger>
        </TabsList>
        {/* Movie Form */}
        <TabsContent value="movies">
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
        </TabsContent>

        {/* Show Form */}
        <TabsContent value="tv">
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
                description="Set the interval for pulling movies from all lists."
              />

              {/* 2x2 grid for Anticipated, Box Office, Popular, Trending */}
              <div className="grid grid-cols-2 gap-4">
                <FormInputField
                  form={showForm}
                  name="anticipated"
                  label="Anticipated Shows"
                  placeholder="Enter # of anticipated shows"
                  description="The number of shows to pull from the Trakt anticipated list."
                />
                <FormInputField
                  form={showForm}
                  name="popular"
                  label="Popular Shows"
                  placeholder="Enter # of popular shows"
                  description="The number of shows to pull from the Trakt popular list."
                />
                <FormInputField
                  form={showForm}
                  name="trending"
                  label="Trending Shows"
                  placeholder="Enter # of trending shows"
                  description="The number of movies to pull from the Trakt trending list."
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
                />
                <FormInputField
                  form={showForm}
                  name="max_runtime"
                  label="Max Run Time"
                  placeholder="Enter maximum runtime"
                  description="The maximum length a episode can be in minutes. (0 for no limit)"
                />
                <FormInputField
                  form={showForm}
                  name="min_year"
                  label="Minimum Year"
                  placeholder="Enter minimum year"
                  description="The minimum year a show can be released. (0 for no limit)"
                />
                <FormInputField
                  form={showForm}
                  name="max_year"
                  label="Maximum Year"
                  placeholder="Enter maximum year"
                  description="The maximum year a show can be released. (0 for no limit)"
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
              <Button type="submit">Submit</Button>
            </form>
          </Form>
        </TabsContent>
      </Tabs>
    </div>
  );
}
