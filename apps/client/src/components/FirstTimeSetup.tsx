import { useState } from "react";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { Button } from "./ui/button";
import { useNavigate } from "react-router-dom";

const SetupStepper = () => {
  const [step, setStep] = useState(0);
  const [loading, setLoading] = useState(false); // Loading state for API request
  const [error, setError] = useState<string | null>(null); // Error state
  const navigate = useNavigate(); // Initialize navigate from react-router-dom

  // Steps for the setup
  const steps = [
    {
      title: "Step 1: User Information",
      content: <Step1 />,
    },
    {
      title: "Step 2: Configure Preferences",
      content: <Step2 />,
    },
    {
      title: "Step 3: Finalize Setup",
      content: <Step3 />,
    },
  ];

  const handleNext = () => {
    if (step < steps.length - 1) {
      setStep(step + 1);
    }
  };

  const handlePrevious = () => {
    if (step > 0) {
      setStep(step - 1);
    }
  };

  const handleFinish = async () => {
    setLoading(true); // Set loading to true during the API request
    try {
      // Make API request to set the setup as complete
      const res = await fetch(`${import.meta.env.VITE_API_URL}/settings`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          key: "SETUP_COMPLETE",
          value: "true",
          type: "boolean",
        }),
      });

      if (!res.ok) {
        throw new Error("Failed to update setup status");
      }

      // After the API request succeeds, navigate to the home page
      navigate("/");
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("An unknown error occurred");
      }
    } finally {
      setLoading(false); // Set loading to false after the request
    }
  };

  return (
    <Dialog open>
      <DialogContent>
        <DialogTitle>{steps[step].title}</DialogTitle>
        <div className="mt-4">{steps[step].content}</div>
        <div className="flex justify-between mt-4">
          {step > 0 && (
            <Button onClick={handlePrevious} className="px-4 py-2 bg-gray-300">
              Previous
            </Button>
          )}
          {step < steps.length - 1 ? (
            <Button
              onClick={handleNext}
              className="px-4 py-2 bg-blue-600 text-white"
            >
              Next
            </Button>
          ) : (
            <Button
              className="px-4 py-2 bg-green-600 text-white"
              onClick={handleFinish}
              disabled={loading} // Disable the button while loading
            >
              {loading ? "Finishing..." : "Finish"}
            </Button>
          )}
        </div>
        {error && <p className="text-red-500 mt-4">{error}</p>}{" "}
        {/* Display errors */}
      </DialogContent>
    </Dialog>
  );
};

// Components for individual steps (can be customized)
const Step1 = () => {
  return (
    <div>
      <p>Please enter your user information.</p>
      {/* Add your form fields here */}
    </div>
  );
};

const Step2 = () => {
  return (
    <div>
      <p>Configure your preferences.</p>
      {/* Add your form fields here */}
    </div>
  );
};

const Step3 = () => {
  return (
    <div>
      <p>Review and finalize your setup.</p>
      {/* Add your form fields here */}
    </div>
  );
};

export default SetupStepper;
