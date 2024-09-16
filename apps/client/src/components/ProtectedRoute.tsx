import SidebarNav from "@/components/SideBarNav";
import { useSetupStatus } from "@/context/SetupContext";
import { useEffect } from "react";
import { useNavigate } from "react-router-dom";

function ProtectedRoute({ element }: { element: JSX.Element }) {
  const { setupComplete } = useSetupStatus();  // Use setupComplete from the context
  const navigate = useNavigate();

  useEffect(() => {
    if (setupComplete === false) {
      navigate("/setup");
    }
  }, [setupComplete, navigate]);

  // Block rendering until setup status is confirmed
  if (setupComplete === null) {
    return null; // Return nothing while setup status is being checked
  }

  return (
    <div className="min-h-screen">
      <SidebarNav />
      <div className="flex flex-col pr-5 pl-16 lg:pl-60 pt-5">{element}</div>
    </div>
  );
}

export default ProtectedRoute;
