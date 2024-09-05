import { useState, useEffect } from "react";
import { RiMoonFill, RiSunFill } from "@remixicon/react";
import { Button } from "@/components/ui/button";

const ThemeToggle = () => {
  const [isDarkMode, setIsDarkMode] = useState(
    typeof window !== "undefined" && localStorage.getItem("theme") === "dark"
  );

  useEffect(() => {
    if (isDarkMode) {
      document.documentElement.classList.add("dark");
      localStorage.setItem("theme", "dark");
    } else {
      document.documentElement.classList.remove("dark");
      localStorage.setItem("theme", "light");
    }
  }, [isDarkMode]);

  const toggleTheme = () => {
    setIsDarkMode(!isDarkMode);
  };

  return (
    <Button variant="ghost" onClick={toggleTheme}>
      {isDarkMode ? (
        <RiSunFill className="w-6 h-6 text-yellow-500" />
      ) : (
        <RiMoonFill className="w-6 h-6" />
      )}
    </Button>
  );
};

export default ThemeToggle;
