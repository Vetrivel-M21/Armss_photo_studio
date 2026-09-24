import { AnimatePresence, motion, type Variants } from "framer-motion";
import { Outlet, useLocation } from "react-router-dom";
import { useMotionSafe } from "../../hooks/useMotionSafe";
import { pageVariants } from "../../lib/motion";

// Reduced motion: no animated frames, final state immediately.
const noopVariants: Variants = {
  initial: {},
  animate: {},
  exit: {},
};

// Wraps the routed page content (never Sidebar) in a fade/slide transition
// keyed on pathname, so navigating between pages animates out/in instead of
// snapping.
export function RouteTransition() {
  const location = useLocation();
  const motionSafe = useMotionSafe();
  const variants = motionSafe ? pageVariants : noopVariants;

  return (
    <AnimatePresence mode="wait" initial={false}>
      <motion.div key={location.pathname} variants={variants} initial="initial" animate="animate" exit="exit">
        <Outlet />
      </motion.div>
    </AnimatePresence>
  );
}
