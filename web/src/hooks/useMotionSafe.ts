import { useReducedMotion } from "framer-motion";

// True unless the user has requested reduced motion — every new Framer Motion
// call site in this app gates its hover/tap/reveal animations on this.
export function useMotionSafe() {
  return !useReducedMotion();
}
