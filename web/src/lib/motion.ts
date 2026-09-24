import type { Transition, Variants } from "framer-motion";

// Framer Motion is the animation boundary for route transitions, card
// hover/tap, and the template gallery reveal — Modal.tsx and Toast.tsx
// intentionally stay on Radix + tailwindcss-animate instead (see the
// comments there).

export const pageVariants: Variants = {
  initial: { opacity: 0, y: 8 },
  animate: { opacity: 1, y: 0, transition: { duration: 0.22, ease: [0.22, 1, 0.36, 1] } },
  exit: { opacity: 0, y: -8, transition: { duration: 0.15, ease: [0.4, 0, 1, 1] } },
};

export const springPremium: Transition = { type: "spring", stiffness: 400, damping: 32, mass: 0.9 };

export const springSnappy: Transition = { type: "spring", stiffness: 500, damping: 28, mass: 0.7 };

export const galleryStagger: Variants = {
  hidden: {},
  visible: { transition: { staggerChildren: 0.035 } },
};

export const galleryItem: Variants = {
  hidden: { opacity: 0, y: 6 },
  visible: { opacity: 1, y: 0 },
};
