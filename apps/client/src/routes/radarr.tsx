import Loading from "@/components/Loading";
import MovieSettingsForm from "@/components/MovieSettingsForm";
import * as React from "react";

export default function Radarr() {
  const [movieSettings, setMovieSettings] = React.useState(null);
  const [loading, setLoading] = React.useState(true);

  const fetchSettings = async () => {
    try {
      const movieResponse = await fetch(
        `${import.meta.env.VITE_API_URL}/movie/settings`
      );
      const movieData = await movieResponse.json();
      setMovieSettings(movieData);

      setLoading(false); // Both requests have completed
    } catch (error) {
      console.error("Error fetching settings:", error);
      setLoading(false);
    }
  };

  React.useEffect(() => {
    fetchSettings();
  }, []);

  if (loading) {
    return <Loading />;
  }

  return (
    <>{movieSettings && <MovieSettingsForm defaultValues={movieSettings} />}</>
  );
}
