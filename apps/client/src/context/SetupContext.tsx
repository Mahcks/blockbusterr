import React, { createContext, useContext, useEffect, useState } from "react";

// Define the expected types for mode and setupComplete
interface SetupContextType {
  setupComplete: boolean | null;
  mode: "ombi" | "radarr-sonarr" | null;
  checkSetupStatus: () => void;
  checkMode: () => void;
}

// Default values for the context
const SetupContext = createContext<SetupContextType>({
  setupComplete: null,
  mode: null,
  checkSetupStatus: () => {},
  checkMode: () => {},
});

// Provider to manage setup status and mode
export function SetupProvider({ children }: { children: React.ReactNode }) {
  const [setupComplete, setSetupComplete] = useState<boolean | null>(null);
  const [mode, setMode] = useState<"ombi" | "radarr-sonarr" | null>(null);

  const checkSetupStatus = async () => {
    try {
      const res = await fetch(
        `${import.meta.env.VITE_API_URL}/settings?key=SETUP_COMPLETE`
      );

      // Check if the response is valid and in JSON format
      const contentType = res.headers.get("content-type");
      if (contentType && contentType.includes("application/json")) {
        const data = await res.json();
        setSetupComplete(data.value === "true");
      } else {
        throw new Error(`Unexpected content-type: ${contentType}`);
      }
    } catch (error) {
      console.error("Error checking setup status", error);
    }
  };

  const checkMode = async () => {
    try {
      const res = await fetch(
        `${import.meta.env.VITE_API_URL}/settings?key=MODE`
      );

      // Check if the response is valid and in JSON format
      const contentType = res.headers.get("content-type");
      if (contentType && contentType.includes("application/json")) {
        const data = await res.json();
        // Cast the response value to the expected types
        setMode(data.value as "ombi" | "radarr-sonarr");
      } else {
        throw new Error(`Unexpected content-type: ${contentType}`);
      }
    } catch (error) {
      console.error("Error checking mode", error);
    }
  };

  useEffect(() => {
    checkSetupStatus(); // Fetch setup status when the app initializes
    checkMode(); // Fetch mode when the app initializes
  }, []);

  return (
    <SetupContext.Provider
      value={{
        setupComplete,
        mode,
        checkSetupStatus,
        checkMode,
      }}
    >
      {children}
    </SetupContext.Provider>
  );
}

// Custom hook to use the SetupContext
export function useSetupStatus() {
  return useContext(SetupContext);
}
