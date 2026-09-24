import { motion } from "framer-motion";
import { useMotionSafe } from "../../hooks/useMotionSafe";

// Fixed (not random-per-render) positions/delays so the field is stable
// across re-renders — only the loop animation itself creates the "drifting
// AI magic" feel, evoking a poster generator that's genuinely at work.
const SPARKLES = [
  { left: "8%", top: "20%", delay: 0, size: 6 },
  { left: "22%", top: "70%", delay: 1.2, size: 4 },
  { left: "40%", top: "15%", delay: 2.4, size: 5 },
  { left: "58%", top: "55%", delay: 0.6, size: 4 },
  { left: "72%", top: "30%", delay: 1.8, size: 6 },
  { left: "85%", top: "65%", delay: 3, size: 4 },
  { left: "50%", top: "80%", delay: 2, size: 5 },
];

interface Props {
  // "light" (white dots) for gradient/colored backgrounds, "brand" (tinted
  // dots) for white/light panels where white dots would be invisible.
  tone?: "light" | "brand";
}

export function SparkleField({ tone = "light" }: Props) {
  const motionSafe = useMotionSafe();
  if (!motionSafe) return null;

  return (
    <div className="pointer-events-none absolute inset-0 overflow-hidden" aria-hidden="true">
      {SPARKLES.map((s, i) => (
        <motion.span
          key={i}
          className={tone === "light" ? "absolute rounded-full bg-white" : "absolute rounded-full bg-brand-secondary/50"}
          style={{ left: s.left, top: s.top, width: s.size, height: s.size }}
          animate={{ opacity: [0, 1, 0], y: [0, -24, -40] }}
          transition={{ duration: 4, repeat: Infinity, delay: s.delay, ease: "easeInOut" }}
        />
      ))}
    </div>
  );
}
