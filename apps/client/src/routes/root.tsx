import "@/index.css";
import "react-grid-layout/css/styles.css";
import "react-resizable/css/styles.css";
import { Responsive, WidthProvider } from "react-grid-layout";
import JobStatusWidget from "@/components/Widgets/JobStatus";
import RecentlyAddedWidget from "@/components/Widgets/RecentlyAdded";
import { GripVertical } from "lucide-react";
import LogWidget from "@/components/Widgets/LogWidget/LogWidget";

const ResponsiveGridLayout = WidthProvider(Responsive);
const initialLayouts = {
  lg: [
    { i: "a", x: 16, y: 0, w: 6, h: 3, minH: 2, minW: 2 },
    { i: "b", x: 0, y: 0, w: 16, h: 3 },
    { i: "c", x: 0, y: 3, w: 16, h: 5 },
  ],
};

const widgetComponents: {
  [key: string]: { component: React.ComponentType<unknown>; title: string };
} = {
  a: { component: JobStatusWidget, title: "Job Status" },
  b: { component: RecentlyAddedWidget, title: "Recently Added" },
  c: { component: LogWidget, title: "Logs" },
};

function Root() {
  return (
    <ResponsiveGridLayout
      className="layout"
      layouts={initialLayouts}
      breakpoints={{ lg: 1200, md: 996, sm: 768, xs: 480, xxs: 0 }}
      cols={{ lg: 22, md: 16, sm: 10, xs: 8, xxs: 6 }}
      rowHeight={100}
      width={1200}
      onLayoutChange={(layout) => console.log(layout)}
      draggableHandle=".drag-handle"
    >
      {initialLayouts.lg.map((layout) => {
        const { component: Component, title } = widgetComponents[layout.i]; // Get component and title
        return (
          <div
            key={layout.i}
            className="bg-[hsl(var(--primary-foreground))] rounded-md p-2 h-full flex flex-col overflow-hidden"
          >
            <div className="flex items-center mb-2">
              <GripVertical className="drag-handle cursor-move mr-2 align-middle" />
              <h2 className="text-white text-base font-bold mt-[0.2 rem] align-middle">
                {title}
              </h2>
            </div>
            <div className="flex-1 overflow-hidden">
              <Component />
            </div>
          </div>
        );
      })}
    </ResponsiveGridLayout>
  );
}

export default Root;
