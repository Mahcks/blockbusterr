import Loading from "@/components/Loading";
import ShowSettingsForm from "@/components/ShowSettingsForm";
import * as React from "react";

export default function Sonarr() {
  const [showSettings, setShowSettings] = React.useState(null);
  const [loading, setLoading] = React.useState(true);

  const fetchSettings = async () => {
    try {
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
  }, []);

  if (loading) {
    return <Loading />;
  }

  return (
    <>{showSettings && <ShowSettingsForm defaultValues={showSettings} />}</>
  );
}
