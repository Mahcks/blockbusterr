import React, { createContext, useContext, useEffect, useState } from 'react';

const SetupContext = createContext({
  setupComplete: null as boolean | null,
  checkSetupStatus: () => {},
});

// Provider to manage setup status
export function SetupProvider({ children }: { children: React.ReactNode }) {
  const [setupComplete, setSetupComplete] = useState<boolean | null>(null);

  const checkSetupStatus = async () => {
    try {
      const res = await fetch(`${import.meta.env.VITE_API_URL}/api/settings?key=SETUP_COMPLETE`);
      const data = await res.json();
      setSetupComplete(data.value === 'true');
    } catch (error) {
      console.error('Error checking setup status', error);
    }
  };

  useEffect(() => {
    checkSetupStatus(); // Fetch setup status when the app initializes
  }, []);

  return (
    <SetupContext.Provider value={{ setupComplete, checkSetupStatus }}>
      {children}
    </SetupContext.Provider>
  );
}

// eslint-disable-next-line react-refresh/only-export-components
export function useSetupStatus() {
  return useContext(SetupContext);
}
