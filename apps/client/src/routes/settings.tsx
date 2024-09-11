import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import MovieSettings from "@/components/MovieSettings";
import TVShowSettings from "@/components/TVShowSettings";
import SidebarNav from "@/components/SideBarNav";
import { useState } from "react";

export default function Settings() {
  // State to persist tab content
  const [activeTab, setActiveTab] = useState("movies");

  return (
    <div>
      <SidebarNav />
      <div className="pl-60 pt-5">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-[400px]">
          <TabsList>
            <TabsTrigger value="movies">Movies</TabsTrigger>
            <TabsTrigger value="shows">Shows</TabsTrigger>
          </TabsList>
          <TabsContent value="movies" hidden={activeTab !== "movies"}>
            <MovieSettings />
          </TabsContent>
          <TabsContent value="shows" hidden={activeTab !== "shows"}>
            <TVShowSettings />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
