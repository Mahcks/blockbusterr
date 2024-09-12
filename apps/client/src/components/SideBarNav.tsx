import { useState } from "react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { Menu, HelpCircle, LayoutDashboard, Cog } from "lucide-react";
import { Link } from "react-router-dom";
import { useSetupStatus } from "@/context/SetupContext";
import { Separator } from "@/components/ui/separator";

const Logo = () => {
  return (
    <p className="text-xl font-bold text-foreground mb-10 ml-3">blockbusterr</p>
  );
};

export default function SidebarNav() {
  const { ombiEnabled } = useSetupStatus();
  const [open, setOpen] = useState(false);

  const NavItems = () => (
    <>
      <Link
        className="flex items-center gap-2 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-foreground"
        to="/"
      >
        <LayoutDashboard className="h-4 w-4" />
        Dashboard
      </Link>

      <Link
        className="flex items-center gap-2 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-foreground"
        to="/settings"
      >
        <Cog className="h-4 w-4" />
        Settings
      </Link>

      <Separator />

      {!ombiEnabled ? (
        <>
          <Link
            className="flex items-center gap-2 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-foreground"
            to="/radarr"
          >
            <HelpCircle className="h-4 w-4" />
            Radarr
          </Link>
          <Link
            className="flex items-center gap-2 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-foreground"
            to="/sonarr"
          >
            <HelpCircle className="h-4 w-4" />
            Sonarr
          </Link>
        </>
      ) : (
        <Link
          className="flex items-center gap-2 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-foreground"
          to="/ombi"
        >
          <HelpCircle className="h-4 w-4" />
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
