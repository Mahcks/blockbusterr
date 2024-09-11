import * as React from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import MovieSettingsForm from "@/components/MovieSettingsForm";
import ShowSettingsForm from "@/components/ShowSettingsForm";
import Loading from "@/components/Loading";

export default function Ombi() {
  const [movieSettings, setMovieSettings] = React.useState(null);
  const [showSettings, setShowSettings] = React.useState(null);
  const [loading, setLoading] = React.useState(true);

  const fetchSettings = async () => {
    try {
      const movieResponse = await fetch(
        `${import.meta.env.VITE_API_URL}/movie/settings`
      );
      const movieData = await movieResponse.json();
      setMovieSettings(movieData);

      const showResponse = await fetch(
        `${import.meta.env.VITE_API_URL}/show/settings`
      );
      const showData = await showResponse.json();
      setShowSettings(showData);

      setLoading(false); // Both requests have completed
    } catch (error) {
      console.error("Error fetching settings:", error);
      setLoading(false);
    }
  };

  React.useEffect(() => {
    fetchSettings();
  }, []); // Only fetch once on component mount

  if (loading) {
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
          {movieSettings && <MovieSettingsForm defaultValues={movieSettings} />}
        </TabsContent>

        {/* Show Form */}
        <TabsContent value="tv">
          {showSettings && <ShowSettingsForm defaultValues={showSettings} />}
        </TabsContent>
      </Tabs>
    </div>
  );
}
