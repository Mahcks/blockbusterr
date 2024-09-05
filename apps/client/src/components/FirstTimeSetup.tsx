import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

const SetupStepper = () => {
  const [step, setStep] = useState(0);

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

  return (
    <Dialog>
      <DialogTrigger>
        <button className="px-4 py-2 bg-blue-600 text-white">
          Start Setup
        </button>
      </DialogTrigger>
      <DialogContent>
        <DialogTitle>{steps[step].title}</DialogTitle>

        <div className="mt-4">{steps[step].content}</div>

        <div className="flex justify-between mt-4">
          {step > 0 && (
            <button onClick={handlePrevious} className="px-4 py-2 bg-gray-300">
              Previous
            </button>
          )}
          {step < steps.length - 1 ? (
            <button
              onClick={handleNext}
              className="px-4 py-2 bg-blue-600 text-white"
            >
              Next
            </button>
          ) : (
            <button
              className="px-4 py-2 bg-green-600 text-white"
              onClick={() => alert("Setup Complete!")}
            >
              Finish
            </button>
          )}
        </div>
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
