import { motion } from "framer-motion";
import { useMotionSafe } from "../../hooks/useMotionSafe";

interface Blob {
  className: string;
  animate: { x: number[]; y: number[]; scale: number[] };
  duration: number;
}

const BLOBS: Blob[] = [
  {
    className: "absolute -left-32 -top-20 h-[36rem] w-[36rem] rounded-full bg-gradient-to-tr from-brand-primary/20 via-brand-secondary/15 to-transparent blur-[120px]",
    animate: { x: [0, 50, -30, 0], y: [0, 40, -20, 0], scale: [1, 1.12, 0.95, 1] },
    duration: 18,
  },
  {
    className: "absolute -right-20 top-1/4 h-[32rem] w-[32rem] rounded-full bg-gradient-to-bl from-brand-violet/20 via-brand-accent/15 to-transparent blur-[110px]",
    animate: { x: [0, -40, 25, 0], y: [0, -30, 20, 0], scale: [1, 0.95, 1.15, 1] },
    duration: 22,
  },
  {
    className: "absolute bottom-10 left-1/4 h-[30rem] w-[30rem] rounded-full bg-gradient-to-t from-brand-highlight/15 via-brand-primary/10 to-transparent blur-[100px]",
    animate: { x: [0, 35, -45, 0], y: [0, -25, 30, 0], scale: [1, 1.1, 0.92, 1] },
    duration: 20,
  },
  {
    className: "absolute right-1/3 -bottom-20 h-[26rem] w-[26rem] rounded-full bg-brand-violet/15 blur-[100px]",
    animate: { x: [0, -20, 30, 0], y: [0, 25, -20, 0], scale: [0.95, 1.05, 1, 0.95] },
    duration: 25,
  },
];

// High-end ambient aurora glow for streaming aesthetic
export function AmbientBackground() {
  const motionSafe = useMotionSafe();
  return (
    <div className="pointer-events-none fixed inset-0 z-0 overflow-hidden" aria-hidden="true">
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_80%_80%_at_50%_-20%,rgba(120,119,198,0.12),rgba(255,255,255,0))]" />
      {BLOBS.map((blob, i) => (
        <motion.div
          key={i}
          className={blob.className}
          animate={motionSafe ? blob.animate : undefined}
          transition={{ duration: blob.duration, repeat: Infinity, ease: "easeInOut" }}
        />
      ))}
      {/* Subtle fine film grain / micro grid texture overlay */}
      <div className="absolute inset-0 bg-[linear-gradient(to_right,#ffffff04_1px,transparent_1px),linear-gradient(to_bottom,#ffffff04_1px,transparent_1px)] bg-[size:4rem_4rem] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_100%)]" />
    </div>
  );
}
