import "@/index.css";
import NavBar from "@/components/NavBar";

function Root() {
  return (
    <div>
      <NavBar />

      <div className="flex min-h-screen flex-col pr-5 pl-5">
        <p className="text-red-500">Home</p>
      </div>
    </div>
  );
}

export default Root;
