import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from "@/components/ui/carousel";

export default function MoviePosterCarousel() {
  const items = [
    {
      id: 1,
      title: "Deadpool & Wolverine",
      image: "https://image.tmdb.org/t/p/w200//8cdWjvZQUExUUTzyp4t6EDMubfO.jpg",
    },
    {
      id: 2,
      title: "Borderlands",
      image: "https://image.tmdb.org/t/p/w500//865DntZzOdX6rLMd405R0nFkLmL.jpg",
    },
    {
      id: 3,
      title: "Rebel Ridge",
      image: "https://image.tmdb.org/t/p/w500//xEt2GSz9z5rSVpIHMiGdtf0czyf.jpg",
    },
    {
      id: 4,
      title: "Inside Out 2",
      image: "https://image.tmdb.org/t/p/w500//vpnVM9B6NMmQpWeZvzLvDESb2QY.jpg",
    },
    {
      id: 5,
      title: "Beetlejuice Beetlejuice",
      image: "https://image.tmdb.org/t/p/w500//kKgQzkUCnQmeTPkyIwHly2t6ZFI.jpg",
    },
    {
      id: 6,
      title: "Despicable Me 4",
      image: "https://image.tmdb.org/t/p/w500//wWba3TaojhK7NdycRhoQpsG0FaH.jpg",
    },
    {
      id: 7,
      title: "Bad Boys: Ride or Die",
      image: "https://image.tmdb.org/t/p/w500//oGythE98MYleE6mZlGs5oBGkux1.jpg",
    },
    {
      id: 8,
      title: "The Killer",
      image: "https://image.tmdb.org/t/p/w500//6PCnxKZZIVRanWb710pNpYVkCSw.jpg",
    },
    {
      id: 9,
      title: "Twilight of the Warriors: Walled In",
      image: "https://image.tmdb.org/t/p/w500//PywbVPeIhBFc33QXktnhMaysmL.jpg",
    },
    {
      id: 10,
      title: "It Ends with Us",
      image: "https://image.tmdb.org/t/p/w500//4TzwDWpLmb9bWJjlN3iBUdvgarw.jpg",
    },
  ];

  return (
    <div className="relative w-full max-w-6xl mx-auto px-4 sm:px-6 lg:px-4">
      <Carousel
        opts={{
          align: "start",
          containScroll: "trimSnaps",
          skipSnaps: true,
        }}
        className="w-full"
      >
        <CarouselContent className="-ml-2 md:-ml-4">
          {items.map((item) => (
            <CarouselItem
              key={item.id}
              className="pl-2 md:pl-4 xs:basis-1/2 sm:basis-1/3 md:basis-1/4 lg:basis-1/5 xl:basis-1/6"
            >
              <div className="relative aspect-[2/3] overflow-hidden rounded-md group">
                <img
                  src={item.image}
                  alt={item.title}
                  sizes="(max-width: 640px) 50vw, (max-width: 768px) 33vw, (max-width: 1024px) 25vw, (max-width: 1280px) 20vw, 16vw"
                  className="object-cover transition-transform duration-300 group-hover:scale-110"
                />
                <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                  <div className="absolute bottom-0 left-0 right-0 p-4">
                    <h3 className="text-white text-lg font-semibold truncate">
                      {item.title}
                    </h3>
                  </div>
                </div>
              </div>
            </CarouselItem>
          ))}
        </CarouselContent>
        <CarouselPrevious className="absolute left-0 top-1/2 -translate-y-1/2 -translate-x-1/2" />
        <CarouselNext className="absolute right-0 top-1/2 -translate-y-1/2 translate-x-1/2" />
      </Carousel>
    </div>
  );
}
