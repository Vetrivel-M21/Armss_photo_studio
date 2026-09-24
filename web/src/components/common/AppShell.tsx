import type { User } from "../../api/types";
import { Sidebar } from "./Sidebar";
import { RouteTransition } from "./RouteTransition";
import { AmbientBackground } from "./AmbientBackground";
import { Footer } from "./Footer";

export function AppShell({ user }: { user: User }) {
  return (
    <div className="relative min-h-screen bg-[#090B10] text-slate-100 md:flex selection:bg-brand-primary selection:text-white">
      <AmbientBackground />
      <div className="relative z-20">
        <Sidebar user={user} />
      </div>
      <div className="relative z-10 min-w-0 flex-1 flex flex-col">
        <main className="mx-auto max-w-7xl px-4 py-6 sm:px-8 lg:py-8 flex-1 w-full">
          <RouteTransition />
        </main>
        <Footer />
      </div>
    </div>
  );
};
