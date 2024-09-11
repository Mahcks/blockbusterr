import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import Root from "@/routes/root.tsx";
import Setup from "@/routes/setup";
import ProtectedRoute from "@/components/ProtectedRoute";
import Radarr from "@/routes/radarr";
import Sonarr from "@/routes/sonarr";
import Ombi from "@/routes/ombi";
import { SetupProvider } from "@/context/SetupContext";
import Settings from "@/routes/settings";

const router = createBrowserRouter([
  {
    path: "/",
    element: <ProtectedRoute element={<Root />} />,
  },
  {
    path: "/setup",
    element: <Setup />,
  },
  {
    path: "/settings",
    element: <ProtectedRoute element={<Settings />} />,
  },
  {
    path: "/radarr",
    element: <ProtectedRoute element={<Radarr />} />,
  },
  {
    path: "/sonarr",
    element: <ProtectedRoute element={<Sonarr />} />,
  },
  {
    path: "/ombi",
    element: <ProtectedRoute element={<Ombi />} />,
  },
]);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <SetupProvider>
      <RouterProvider router={router} />
    </SetupProvider>
  </StrictMode>
);
