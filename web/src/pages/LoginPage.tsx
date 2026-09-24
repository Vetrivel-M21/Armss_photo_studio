import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import { Sparkles } from "lucide-react";
import { useLogin } from "../hooks/useAuth";
import { ErrorBanner } from "../components/common/ErrorBanner";
import { Input } from "../components/common/Input";
import { Button } from "../components/common/Button";
import { SparkleField } from "../components/common/SparkleField";
import { Footer } from "../components/common/Footer";
import { useMotionSafe } from "../hooks/useMotionSafe";
import { ApiError } from "../api/client";

export function LoginPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const login = useLogin();
  const navigate = useNavigate();
  const motionSafe = useMotionSafe();

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    login.mutate(
      { email: username, password },
      { onSuccess: () => navigate("/") },
    );
  }

  return (
    <div className="relative flex min-h-screen items-center justify-center bg-[#090B10] px-4 overflow-hidden">
      {/* Background ambient orbs */}
      <div className="pointer-events-none absolute -top-40 -left-40 h-[32rem] w-[32rem] rounded-full bg-brand-primary/20 blur-[120px]" />
      <div className="pointer-events-none absolute -bottom-40 -right-40 h-[32rem] w-[32rem] rounded-full bg-purple-600/20 blur-[120px]" />
      <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(to_right,#ffffff04_1px,transparent_1px),linear-gradient(to_bottom,#ffffff04_1px,transparent_1px)] bg-[size:4rem_4rem]" />

      <div className="relative z-10 w-full max-w-sm flex flex-col items-center gap-5">
        <div className="w-full overflow-hidden rounded-3xl border border-white/15 bg-gradient-to-b from-[#141926]/95 to-[#0C0F18]/95 p-1 shadow-2xl backdrop-blur-2xl">
          <div className="relative overflow-hidden rounded-[1.4rem] bg-[#0E121E]/90 p-8 shadow-glass">
            <motion.div
              className="pointer-events-none absolute -right-10 -top-10 h-40 w-40 rounded-full bg-brand-primary/15 blur-3xl"
              aria-hidden="true"
              animate={motionSafe ? { x: [0, 12, -8, 0], y: [0, -10, 8, 0], scale: [1, 1.1, 0.95, 1] } : undefined}
              transition={{ duration: 15, repeat: Infinity, ease: "easeInOut" }}
            />
            <SparkleField tone="brand" />

            <div className="relative text-center">
              <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-primary via-brand-secondary to-purple-600 shadow-glow">
                <Sparkles className="h-6 w-6 text-white animate-pulse" aria-hidden="true" />
              </div>
              <h1 className="mt-4 text-2xl font-extrabold tracking-tight text-white">
                ARMSS <span className="bg-gradient-to-r from-brand-primary to-brand-highlight bg-clip-text text-transparent">STUDIO</span>
              </h1>
              <p className="mt-1 text-xs font-medium uppercase tracking-wider text-slate-400">Next-Gen AI Poster Engine</p>
            </div>

            <form onSubmit={onSubmit} className="relative mt-7">
              {login.isError && (
                <div className="mb-4">
                  <ErrorBanner message={login.error instanceof ApiError ? login.error.message : "Could not log in. Try again."} />
                </div>
              )}

              <div className="space-y-4">
                <div className="space-y-1.5">
                  <label htmlFor="username" className="text-xs font-bold uppercase tracking-wider text-slate-300">
                    Email / Username
                  </label>
                  <Input
                    id="username"
                    type="text"
                    autoComplete="username"
                    required
                    autoFocus
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    placeholder="name@company.com"
                  />
                </div>
                <div className="space-y-1.5">
                  <label htmlFor="password" className="text-xs font-bold uppercase tracking-wider text-slate-300">
                    Password
                  </label>
                  <Input
                    id="password"
                    type="password"
                    autoComplete="current-password"
                    required
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="••••••••"
                  />
                </div>
              </div>

              <Button type="submit" disabled={login.isPending} className="mt-7 w-full py-3 text-sm font-bold shadow-glow">
                {login.isPending ? "Authenticating..." : "Sign in to Studio"}
              </Button>
            </form>
          </div>
        </div>

        <Footer className="border-none bg-transparent py-0 text-slate-400" />
      </div>
    </div>
  );
}
