import * as React from "react";
import NavBar from "@/components/NavBar";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Check, ChevronsUpDown } from "lucide-react";
import MovieSettings from "@/components/MovieSettings";
import TVShowSettings from "@/components/TVShowSettings";

const categories = [
  {
    value: "movies",
    label: "Movies",
  },
  {
    value: "tv-shows",
    label: "TV Shows",
  },
];

export default function Settings() {
  // Deals with the state of the settings category
  const [openCategory, setOpenCategory] = React.useState(false);
  const [settingsCategory, setSettingsCategory] = React.useState("movies");

  // Conditionally render content based on settingsCategory
  const renderSettingsPage = () => {
    switch (settingsCategory) {
      case "movies":
        return <MovieSettings />;
      case "tv-shows":
        return <TVShowSettings />;
      default:
        return null;
    }
  };

  return (
    <div>
      <NavBar />

      <div className="flex min-h-screen flex-col pr-5 pl-5">
        <Popover open={openCategory} onOpenChange={setOpenCategory}>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              role="combobox"
              aria-expanded={openCategory}
              className="w-[200px] justify-between"
            >
              {settingsCategory
                ? categories.find(
                    (category) => category.value === settingsCategory
                  )?.label
                : "Select setting..."}
              <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-[200px] p-0">
            <Command>
              <CommandList>
                <CommandEmpty>No setting found.</CommandEmpty>
                <CommandGroup>
                  {categories.map((category) => (
                    <CommandItem
                      key={category.value}
                      value={category.value}
                      onSelect={(currentValue) => {
                        setSettingsCategory(
                          currentValue === settingsCategory ? "" : currentValue
                        );
                        setOpenCategory(false);
                      }}
                    >
                      <Check
                        className={cn(
                          "mr-2 h-4 w-4",
                          settingsCategory === category.value
                            ? "opacity-100"
                            : "opacity-0"
                        )}
                      />
                      {category.label}
                    </CommandItem>
                  ))}
                </CommandGroup>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>

        {/* Render the settings page based on the selected category */}
        <div className="mt-5">{renderSettingsPage()}</div>
      </div>
    </div>
  );
}
