import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

// Conditional class composition with Tailwind conflict resolution — used by
// every shared primitive that accepts a `className` override or variant map.
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
