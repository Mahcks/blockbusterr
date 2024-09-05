import { useState } from "react";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { Button } from "./ui/button";
import { useNavigate } from "react-router-dom";
import { z } from "zod";
import { FormProvider, useForm, useFormContext } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Info } from "lucide-react";

// Define Zod schema for validation
const setupFormSchema = z.object({
  trakt: z.object({
    client_id: z
      .string()
      .min(10, { message: "Client ID must be at least 10 characters." }),
    client_secret: z
      .string()
      .min(10, { message: "Client Secret must be at least 10 characters." }),
  }),
  movies: z.object({
    interval: z.number().int().min(0),
  }),
  tvShows: z.object({
    quality: z.enum(["low", "medium", "high"], {
      message: "Please select a quality.",
    }),
    notifications: z.boolean(),
  }),
});

type SetupFormValues = z.infer<typeof setupFormSchema>;

const SetupStepper = () => {
  const [step, setStep] = useState(0);
  const [loading, setLoading] = useState(false); // Loading state for API request
  const [error, setError] = useState<string | null>(null); // Error state
  const navigate = useNavigate(); // Initialize navigate from react-router-dom

  // Use react-hook-form with Zod validation and set up form default values
  const formMethods = useForm<SetupFormValues>({
    resolver: zodResolver(setupFormSchema),
    defaultValues: {
      trakt: { client_id: "", client_secret: "" },
      movies: { interval: 1 },
      tvShows: { quality: "medium", notifications: false },
    },
  });

  // Steps for the setup
  const steps = [
    {
      title: "Step 1: Provide Trakt app information",
      content: <Step1 />,
    },
    {
      title: "Step 2: Configure Movie Preferences",
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

  const handleFinish = async (data: SetupFormValues) => {
    setLoading(true); // Set loading to true during the API request
    try {
      console.log("Form Data:", data);

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
    <FormProvider {...formMethods}>
      <Dialog open>
        <DialogContent>
          <DialogTitle>{steps[step].title}</DialogTitle>

          <form onSubmit={formMethods.handleSubmit(handleFinish)}>
            <div className="mt-4">{steps[step].content}</div>

            <div className="flex justify-between mt-4">
              {step > 0 && (
                <Button
                  onClick={handlePrevious}
                  type="button"
                  className="px-4 py-2 bg-gray-300"
                >
                  Previous
                </Button>
              )}
              {step < steps.length - 1 ? (
                <Button
                  onClick={handleNext}
                  type="button"
                  className="px-4 py-2 bg-blue-600 text-white"
                >
                  Next
                </Button>
              ) : (
                <Button
                  type="submit"
                  className="px-4 py-2 bg-green-600 text-white"
                  onClick={handleFinish}
                >
                  Finish
                </Button>
              )}
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </FormProvider>
  );
};

// Step 1: User Information
const Step1 = () => {
  const { control } = useFormContext();

  return (
    <div className="py-3">
      <FormField
        control={control}
        name="trakt.clientId"
        render={({ field }) => (
          <FormItem className="pb-5">
            <FormLabel>Client ID</FormLabel>
            <FormControl>
              <Input placeholder="Enter Trakt Client ID" {...field} />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />
      <FormField
        control={control}
        name="trakt.clientSecret"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Client Secret</FormLabel>
            <FormControl>
              <Input placeholder="Enter Trakt Client Secret" {...field} />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />
    </div>
  );
};

// Step 2: Preferences
const Step2 = () => {
  const { control } = useFormContext();

  return (
    <div>
      <FormField
        control={control}
        name="movies.interval"
        render={({ field }) => (
          <FormItem>
            <FormLabel>
              <span className="flex items-center gap-2 text-lg">
                <p className="font-bold">Interval</p>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Info className="w-4 h-4 text-gray-500 cursor-pointer" />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Interval in hours for checking movies.</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </span>
            </FormLabel>
            <FormControl>
              <div className="flex items-center gap-2">
                <Input type="number" {...field} />
                <span>hours</span>
              </div>
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />
    </div>
  );
};

// Step 3: Review and Confirm
const Step3 = () => {
  const { control } = useFormContext();

  return (
    <div>
      <FormField
        control={control}
        name="tvShows.quality"
        render={({ field }) => (
          <FormItem>
            <FormLabel>TV Show Quality</FormLabel>
            <FormControl>
              <select {...field}>
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
              </select>
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />
      <FormField
        control={control}
        name="tvShows.notifications"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Enable Notifications</FormLabel>
            <FormControl>
              <Checkbox
                checked={field.value}
                onCheckedChange={field.onChange}
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />
    </div>
  );
};

export default SetupStepper;
