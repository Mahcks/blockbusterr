import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import Root from "@/routes/root.tsx";
import Settings from "@/routes/settings.tsx";
import Setup from "./routes/setup";
import ProtectedRoute from "@/routes/ProtectedRoute";

const router = createBrowserRouter([
  {
    path: "/",
    element: <ProtectedRoute element={<Root />} />,
  },
  {
    path: "/settings",
    element: <ProtectedRoute element={<Settings />} />,
  },
  {
    path: "/setup",
    element: <Setup />,
  },
]);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>
);
