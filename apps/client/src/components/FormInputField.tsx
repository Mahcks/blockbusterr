// components/FormInputField.tsx
import {
  FormItem,
  FormLabel,
  FormControl,
  FormDescription,
  FormMessage,
  FormField,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { UseFormReturn } from "react-hook-form";

interface FormInputFieldProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  form: UseFormReturn<any>;
  name: string;
  label: string;
  placeholder: string;
  description?: string;
  disabled?: boolean;
  isNumber?: boolean; // Added prop to indicate if the input should be treated as a number
}

const FormInputField: React.FC<FormInputFieldProps> = ({
  form,
  name,
  label,
  placeholder,
  description,
  disabled = false,
  isNumber = false, // Default to false if not provided
}) => {
  return (
    <FormField
      control={form.control}
      name={name}
      render={({ field }) => (
        <FormItem>
          <FormLabel>{label}</FormLabel>
          <FormControl>
            <Input
              disabled={disabled}
              type={isNumber ? "number" : "text"} // Set input type based on isNumber prop
              value={field.value ?? ""} // Ensure value is never undefined or null
              onChange={(e) =>
                field.onChange(
                  isNumber ? e.target.valueAsNumber : e.target.value
                )
              } // Use valueAsNumber for number inputs to handle numeric values correctly
              placeholder={placeholder}
            />
          </FormControl>
          {description && <FormDescription>{description}</FormDescription>}
          <FormMessage />
        </FormItem>
      )}
    />
  );
};

export default FormInputField;
