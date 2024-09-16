import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatBytes(bytes: number): string {
  if (typeof bytes !== 'number' || isNaN(bytes) || bytes < 0) {
    return 'Invalid input';
  }
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const dm = 1; // Number of decimal places
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  // Ensure 'i' does not exceed the length of the sizes array
  const index = Math.min(i, sizes.length - 1);

  const converted = parseFloat((bytes / Math.pow(k, index)).toFixed(dm));
  return `${converted} ${sizes[index]}`;
}
