import "@/index.css";

import { Card, CardContent } from "@/components/ui/card";
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from "@/components/ui/carousel";
import SidebarNav from "@/components/SideBarNav";

// Sample data for the carousel items
const items = Array.from({ length: 50 }, (_, i) => ({
  id: i + 1,
  title: `Item ${i + 1}`,
}));

function Root() {
  return (
    <div className="min-h-screen">
      <SidebarNav />
      <div className="flex min-h-screen flex-col pr-5 pl-60 pt-5">
        <div>
          <Carousel
            opts={{
              align: "start",
              loop: true,
            }}
            className="w-full max-w-7xl mx-auto"
          >
            <CarouselContent className="-ml-2 md:-ml-4">
              {items.map((item) => (
                <CarouselItem
                  key={item.id}
                  className="pl-2 md:pl-4 basis-1/2 sm:basis-1/3 md:basis-1/5 lg:basis-1/10"
                >
                  <Card>
                    <CardContent className="flex aspect-square items-center justify-center p-6">
                      <span className="text-3xl font-semibold">
                        {item.title}
                      </span>
                    </CardContent>
                  </Card>
                </CarouselItem>
              ))}
            </CarouselContent>
            <CarouselPrevious />
            <CarouselNext />
          </Carousel>
        </div>
      </div>
    </div>
  );
}

export default Root;
