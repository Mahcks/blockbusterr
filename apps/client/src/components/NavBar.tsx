import { Link, useLocation } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { RiHome2Fill, RiToolsFill } from "@remixicon/react";
import ThemeToggle from "@/components/ThemeToggle";

export default function sNavBar() {
  const location = useLocation();

  // If the current route is "/settings", change the link to "/"
  const isSettingsPage = location.pathname === "/settings";
  const linkPath = isSettingsPage ? "/" : "/settings";
  const icon = isSettingsPage ? (
    <RiHome2Fill className="h-6 w-6" />
  ) : (
    <RiToolsFill className="h-6 w-6" />
  );

  return (
    <div className="sticky top-0 flex justify-between items-center p-5">
      <div>
        <Link to={"/"}>
          <h3 className="text-xl font-bold">blockbusterr</h3>
        </Link>
      </div>

      <div className="flex space-x-2">
        <ThemeToggle />
        <Link to={linkPath}>
          <Button variant="ghost">{icon}</Button>
        </Link>
      </div>
    </div>
  );
}
