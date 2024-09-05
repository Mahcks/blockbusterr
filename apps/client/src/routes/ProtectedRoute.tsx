import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

function ProtectedRoute({ element }: { element: JSX.Element }) {
  const [setupComplete, setSetupComplete] = useState<boolean | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    const checkSetupStatus = async () => {
      try {
        const res = await fetch(
          `${import.meta.env.VITE_API_URL}/settings?key=SETUP_COMPLETE`
        );
        const data = await res.json();

        if (data.value === "true") {
          setSetupComplete(true);
        } else {
          navigate("/setup");
        }
      } catch (error) {
        console.error("Error checking setup status", error);
        navigate("/setup");
      }
    };

    checkSetupStatus();
  }, [navigate]);

  // Block rendering until setup status is confirmed
  if (setupComplete === null) {
    return null; // Return nothing while setup status is being checked
  }

  return element;
}

export default ProtectedRoute;
