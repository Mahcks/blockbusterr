import { useState } from "react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { Menu, LayoutDashboard, Cog } from "lucide-react";
import { Link, useLocation } from "react-router-dom"; // Import useLocation
import { useSetupStatus } from "@/context/SetupContext";
import { Separator } from "@/components/ui/separator";

const Logo = () => {
  return (
    <p className="text-xl font-bold text-foreground mb-10 ml-3">blockbusterr</p>
  );
};

export default function SidebarNav() {
  const { mode } = useSetupStatus();
  const [open, setOpen] = useState(false);
  const location = useLocation(); // Get the current location

  const NavItems = () => (
    <>
      <Link
        className={`flex items-center gap-2 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-foreground ${
          location.pathname === "/" ? "font-bold text-foreground" : ""
        }`}
        to="/"
      >
        <LayoutDashboard className="h-4 w-4" />
        Dashboard
      </Link>

      <Link
        className={`flex items-center gap-2 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-foreground ${
          location.pathname === "/settings" ? "font-bold text-foreground" : ""
        }`}
        to="/settings"
      >
        <Cog className="h-4 w-4" />
        Settings
      </Link>

      <Separator />

      <Link
        className={`flex items-center gap-2 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-foreground ${
          location.pathname === "/trakt" ? "font-bold text-foreground" : ""
        }`}
        to="/trakt"
      >
        <img src="/trakt.png" alt="Trakt" className="h-4 w-4" />
        Trakt
      </Link>

      {mode !== "ombi" ? (
        <>
          <Link
            className={`flex items-center gap-2 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-foreground ${
              location.pathname === "/radarr" ? "font-bold text-foreground" : ""
            }`}
            to="/radarr"
          >
            <img src="/radarr.png" alt="Sonarr" className="h-4 w-4" />
            Radarr
          </Link>
          <Link
            className={`flex items-center gap-2 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-foreground ${
              location.pathname === "/sonarr" ? "font-bold text-foreground" : ""
            }`}
            to="/sonarr"
          >
            <img src="/sonarr.png" alt="Sonarr" className="h-4 w-4" />
            Sonarr
          </Link>
        </>
      ) : (
        <Link
          className={`flex items-center gap-2 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-foreground ${
            location.pathname === "/ombi" ? "font-bold text-foreground" : ""
          }`}
          to="/ombi"
        >
          <img src="/ombi.png" alt="Ombi" className="h-4 w-4" />
          Ombi
        </Link>
      )}
    </>
  );

  return (
    <>
      {/* Mobile Menu Button */}
      <Sheet open={open} onOpenChange={setOpen}>
        <SheetTrigger asChild>
          <Button
            variant="outline"
            size="icon"
            className="fixed left-4 top-4 z-40 lg:hidden"
          >
            <Menu className="h-6 w-6" />
            <span className="sr-only">Toggle Menu</span>
          </Button>
        </SheetTrigger>
        <SheetContent side="left" className="w-64 p-0 bg-background">
          <ScrollArea className="h-full px-3 py-16">
            <Logo />
            <nav>
              <NavItems />
            </nav>
          </ScrollArea>
        </SheetContent>
      </Sheet>

      {/* Desktop Sidebar */}
      <div className="hidden lg:fixed lg:inset-y-0 lg:left-0 lg:z-50 lg:block lg:w-56 lg:overflow-y-auto lg:border-r lg:border-border lg:bg-card lg:pb-4 lg:pt-6">
        <nav className="px-3">
          <Logo />
          <nav>
            <NavItems />
          </nav>
        </nav>
      </div>
    </>
  );
}
