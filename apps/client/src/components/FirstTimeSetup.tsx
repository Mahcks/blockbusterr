import { useState } from "react";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { Button } from "./ui/button";
import { useNavigate } from "react-router-dom";
import { z } from "zod";
import {
  Controller,
  FormProvider,
  useForm,
  useFormContext,
} from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  FormControl,
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

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
  providers: z.object({
    provider: z.enum(["ombi", "radarr-sonarr"]),
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
      providers: { provider: "radarr-sonarr" },
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
      title: "Configure Ombi or Radarr/Sonarr",
      content: <Step2 />,
    },
    {
      title: "Step 3: Configure Movie Preferences",
      content: <Step3 />,
    },
    {
      title: "Step 4: Configure TV Preferences",
      content: <Step4 />,
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

  const handleFinish: React.MouseEventHandler<HTMLButtonElement> = async (
    event
  ) => {
    event.preventDefault(); // Prevent the default form submission behavior
    setLoading(true); // Set loading to true during the API request
    try {
      const data = formMethods.getValues(); // Get the form values using react-hook-form
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
  return (
    <div className="py-3">
      <InputField
        name="trakt.clientId"
        label="Client ID"
        placeholder="Enter Trakt Client ID"
      />
      <InputField
        name="trakt.clientSecret"
        label="Client Secret"
        placeholder="Enter Trakt Client Secret"
      />
    </div>
  );
};

const Step2 = () => {
  const providerOptions = [
    { value: "radarr-sonarr", label: "Radarr/Sonarr" },
    { value: "ombi", label: "Ombi" },
  ];

  return (
    <SelectField
      name="providers.provider"
      label="Select your provider"
      options={providerOptions}
    />
  );
};

const Step3 = () => {
  return (
    <div>
      <FormItem>
        <FormLabel>
          <span className="flex items-center gap-2">
            <p className="font-bold text-lg">Interval</p>
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
        <div className="flex items-center gap-2 w-full">
          <div className="flex-grow">
            <InputField
              name="movies.interval"
              label=""
              placeholder="Enter interval"
              type="number"
            />
          </div>
          <p>hours</p>
        </div>
      </FormItem>
    </div>
  );
};

// Step 3: Review and Confirm
const Step4 = () => {
  const qualityOptions = [
    { value: "low", label: "Low" },
    { value: "medium", label: "Medium" },
    { value: "high", label: "High" },
  ];

  return (
    <div>
      <SelectField
        name="tvShows.quality"
        label="TV Show Quality"
        options={qualityOptions}
      />
      <CheckboxField
        name="tvShows.notifications"
        label="Enable Notifications"
      />
    </div>
  );
};

export default SetupStepper;

// FormField components for different field types

// InputField component
export const InputField = ({
  name,
  label,
  placeholder,
  type = "text",
}: // eslint-disable-next-line @typescript-eslint/no-explicit-any
any) => {
  const { control } = useFormContext();

  return (
    <FormItem>
      <FormLabel>{label}</FormLabel>
      <FormControl>
        <Controller
          name={name}
          control={control}
          render={({ field }) => (
            <Input {...field} type={type} placeholder={placeholder} />
          )}
        />
      </FormControl>
      <FormMessage />
    </FormItem>
  );
};

// SelectField component
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const SelectField = ({ name, label, options }: any) => {
  const { control } = useFormContext();

  return (
    <FormItem>
      <FormLabel>{label}</FormLabel>
      <FormControl>
        <Controller
          name={name}
          control={control}
          render={({ field }) => (
            <Select onValueChange={field.onChange} value={field.value}>
              <SelectTrigger className="w-[180px]">
                <SelectValue>{field.value || "Select an option"}</SelectValue>
              </SelectTrigger>
              <SelectContent>
                {options.map((option: any) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        />
      </FormControl>
      <FormMessage />
    </FormItem>
  );
};

// CheckboxField component
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const CheckboxField = ({ name, label }: any) => {
  const { control } = useFormContext();

  return (
    <FormItem>
      <FormLabel>{label}</FormLabel>
      <FormControl>
        <Controller
          name={name}
          control={control}
          render={({ field }) => (
            <Checkbox checked={field.value} onCheckedChange={field.onChange} />
          )}
        />
      </FormControl>
      <FormMessage />
    </FormItem>
  );
};
