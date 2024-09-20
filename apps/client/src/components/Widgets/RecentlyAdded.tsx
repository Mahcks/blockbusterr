import * as React from "react";

import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from "@/components/ui/carousel";
import { RecentlyAddedMedia } from "@/types/recently_added";

export default function MoviePosterCarousel() {
  const [items, setItems] = React.useState<RecentlyAddedMedia[]>([]);
  const [page, setPage] = React.useState(1);
  const [loading, setLoading] = React.useState<boolean>(true);
  const [hasMore, setHasMore] = React.useState<boolean>(true);

  const observer = React.useRef<IntersectionObserver | null>(null);
  const pageSize = 10;

  // Function to fetch recently added media with pagination
  const fetchRecentlyAddedMedia = async (page: number) => {
    setLoading(true);
    try {
      const response = await fetch(
        `${
          import.meta.env.VITE_API_URL
        }/media/recentlyadded?page=${page}&pageSize=${pageSize}`
      );
      if (!response.ok) {
        throw new Error("Failed to fetch recently added media.");
      }
      const data: RecentlyAddedMedia[] = await response.json();

      // If the returned data is less than page size, there are no more items
      if (data.length < pageSize) {
        setHasMore(false); // No more data, stop further requests
      }

      setItems((prevItems) => [...prevItems, ...data]); // Append new items
    } catch (error) {
      console.error("Error fetching recently added media:", error);
    } finally {
      setLoading(false);
    }
  };

  // useEffect to fetch data on component mount and whenever the page changes
  React.useEffect(() => {
    if (hasMore) {
      fetchRecentlyAddedMedia(page);
    }
  }, [page]);

  // Callback function for the IntersectionObserver
  const lastMediaElementRef = React.useCallback(
    (node: HTMLDivElement | null) => {
      if (loading || !hasMore || !node) return;
      if (observer.current) observer.current.disconnect();
      observer.current = new IntersectionObserver((entries) => {
        if (entries[0].isIntersecting && hasMore) {
          setPage((prevPage) => prevPage + 1); // Load the next page
        }
      });
      observer.current.observe(node);
    },
    [loading, hasMore]
  );

  // If still loading and no items have been fetched yet
  if (loading && items.length === 0) {
    return <div>Loading...</div>;
  }

  // If loading is complete and no items are available
  if (!loading && items.length === 0) {
    return (
      <div className="text-center text-gray-500 mt-8">
        No recently added media... check back later!
      </div>
    ); // Display an alternative content
  }

  return (
    <div className="relative w-full max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
      <Carousel
        opts={{
          align: "start",
          containScroll: "trimSnaps",
          skipSnaps: true,
        }}
        className="w-full"
      >
        <CarouselContent className="-ml-2 md:-ml-4">
          {items.map((item, index) => {
            // Use the ref on the last item to trigger loading the next page
            if (items.length === index + 1) {
              return (
                <CarouselItem
                  key={item.id}
                  ref={hasMore ? lastMediaElementRef : null}
                  className="pl-1 md:pl-4 xs:basis-1/2 sm:basis-1/3 md:basis-1/4 lg:basis-1/5 xl:basis-1/6"
                >
                  <div className="relative w-[160px] h-[250px] aspect-w-2 aspect-h-3 overflow-hidden rounded-md group">
                    <img
                      src={item.poster}
                      alt={item.title}
                      width={300}
                      height={450}
                      sizes="(max-width: 640px) 50vw, (max-width: 768px) 33vw, (max-width: 1024px) 25vw, (max-width: 1280px) 20vw, 16vw"
                      className="object-cover w-full h-full transition-transform duration-300 group-hover:scale-110 select-none"
                    />
                    <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                      <div className="absolute bottom-0 left-0 right-0 p-4">
                        <h3 className="text-white text-lg font-semibold select-none">
                          {item.title}
                        </h3>
                      </div>
                    </div>
                  </div>
                </CarouselItem>
              );
            } else {
              return (
                <CarouselItem
                  key={item.id}
                  className="pl-1 md:pl-4 xs:basis-1/2 sm:basis-1/3 md:basis-1/4 lg:basis-1/5 xl:basis-1/6"
                >
                  <div className="relative w-[160px] h-[250px] aspect-w-2 aspect-h-3 overflow-hidden rounded-md group">
                    <img
                      src={item.poster}
                      alt={item.title}
                      width={300}
                      height={450}
                      sizes="(max-width: 640px) 50vw, (max-width: 768px) 33vw, (max-width: 1024px) 25vw, (max-width: 1280px) 20vw, 16vw"
                      className="object-cover w-full h-full transition-transform duration-300 group-hover:scale-110 select-none"
                    />
                    <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                      <div className="absolute bottom-0 left-0 right-0 p-4">
                        <h3 className="text-white text-lg font-semibold select-none">
                          {item.title}
                        </h3>
                      </div>
                    </div>
                  </div>
                </CarouselItem>
              );
            }
          })}
        </CarouselContent>
        <CarouselPrevious className="absolute left-0 top-1/2 -translate-y-1/2 -translate-x-1/2" />
        <CarouselNext className="absolute right-0 top-1/2 -translate-y-1/2 translate-x-1/2" />
      </Carousel>

      {loading && hasMore && <div>Loading more...</div>}
    </div>
  );
}
