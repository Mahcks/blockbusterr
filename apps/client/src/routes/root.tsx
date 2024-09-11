import "@/index.css";
import * as React from "react";
import { createSwapy } from "swapy";

const DEFAULT = {
  "1": "a",
  "3": "c",
  "4": "d",
  "2": null,
};

function A() {
  return (
    <div
      className="item a w-full h-full flex items-center justify-center text-white text-5xl cursor-pointer select-none relative bg-[#b95050]"
      data-swapy-item="a"
    >
      <div
        className="handle absolute left-0 top-0 w-[20px] h-full bg-[rgba(0,0,0,0.5)]"
        data-swapy-handle
      ></div>
      <div>A</div>
    </div>
  );
}

function C() {
  return (
    <div
      className="item c w-full h-full flex items-center justify-center text-white text-5xl cursor-pointer select-none relative bg-[#508db9]"
      data-swapy-item="c"
    >
      <div>C</div>
    </div>
  );
}

function D() {
  return (
    <div
      className="item d w-full h-full flex items-center justify-center text-white text-5xl cursor-pointer select-none relative bg-[#b95096]"
      data-swapy-item="d"
    >
      <div>D</div>
    </div>
  );
}

function getItemById(itemId: "a" | "c" | "d" | null) {
  switch (itemId) {
    case "a":
      return <A />;
    case "c":
      return <C />;
    case "d":
      return <D />;
  }
}

function Root() {
  const slotItems: Record<string, "a" | "c" | "d" | null> =
    localStorage.getItem("slotItem")
      ? JSON.parse(localStorage.getItem("slotItem")!)
      : DEFAULT;

  React.useEffect(() => {
    const container = document.querySelector(".container")!;
    if (!container) {
      console.error("Swapy container not found!");
      return;
    }

    const swapy = createSwapy(container);
    console.log("Swapy initialized", swapy);

    swapy.onSwap(({ data }) => {
      console.log("Swapped items", data.object);
      localStorage.setItem("slotItem", JSON.stringify(data.object));
    });

    return () => {
      swapy.destroy();
      console.log("Swapy destroyed");
    };
  }, []);

  return (
    <div className="container flex flex-col gap-1 w-full p-2">
      <div
        className="slot a bg-[#111] flex-shrink-0 basis-[150px] h-[150px]"
        data-swapy-slot="1"
      >
        {getItemById(slotItems["1"])}
      </div>
      <div className="second-row flex gap-1 h-[200px]">
        <div className="slot b bg-[#111] flex-[2]" data-swapy-slot="2">
          {getItemById(slotItems["2"])}
        </div>
        <div className="slot c bg-[#111] flex-[1]" data-swapy-slot="3">
          {getItemById(slotItems["3"])}
        </div>
      </div>
      <div
        className="slot d bg-[#111] flex-shrink-0 basis-[200px] h-[400px]"
        data-swapy-slot="4"
      >
        {getItemById(slotItems["4"])}
      </div>
    </div>
  );
}

export default Root;
