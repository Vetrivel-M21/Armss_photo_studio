import { useState, type ReactNode } from "react";
import { NavLink } from "react-router-dom";
import { motion } from "framer-motion";
import { Image, LayoutTemplate, LogOut, Menu, Settings, Sparkles } from "lucide-react";
import type { User } from "../../api/types";
import { useLogout } from "../../hooks/useAuth";
import { Modal } from "./Modal";
import { cn } from "../../lib/cn";
import { useMotionSafe } from "../../hooks/useMotionSafe";
import { springPremium } from "../../lib/motion";

interface Props {
  user: User;
}

const navItemClass = ({ isActive }: { isActive: boolean }) =>
  cn(
    "group relative flex items-center gap-3 rounded-xl px-3.5 py-2.5 text-sm font-medium transition-all duration-200 motion-reduce:transition-none",
    isActive
      ? "bg-gradient-to-r from-brand-primary via-brand-secondary to-purple-600 text-white shadow-glow-sm font-semibold"
      : "text-slate-400 hover:bg-white/[0.06] hover:text-white",
  );

function Logo() {
  return (
    <div className="flex items-center gap-3">
      <div className="relative flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-brand-primary via-brand-secondary to-purple-600 shadow-glow-sm">
        <Sparkles className="h-5 w-5 text-white animate-pulse" aria-hidden="true" />
      </div>
      <div>
        <span className="block text-base font-extrabold tracking-tight text-white">
          ARMSS <span className="bg-gradient-to-r from-brand-primary to-brand-highlight bg-clip-text text-transparent">STUDIO</span>
        </span>
        <span className="block text-[10px] font-medium tracking-wider uppercase text-slate-400">AI Poster Engine</span>
      </div>
    </div>
  );
}

function NavItemContent({ children }: { children: ReactNode }) {
  const motionSafe = useMotionSafe();
  return (
    <motion.span
      className="flex items-center gap-3"
      whileHover={motionSafe ? { x: 3 } : undefined}
      transition={springPremium}
    >
      {children}
    </motion.span>
  );
}

function NavItems({ user, onNavigate }: { user: User; onNavigate?: () => void }) {
  return (
    <nav className="flex flex-1 flex-col gap-1.5 px-3">
      <div className="mb-2 px-3 pt-2 text-[10px] font-semibold uppercase tracking-wider text-slate-500">
        Studio
      </div>
      <NavLink to="/" end className={navItemClass} onClick={onNavigate}>
        <NavItemContent>
          <Sparkles className="h-4 w-4 text-brand-accent group-hover:text-white transition-colors" aria-hidden="true" />
          Create Poster
        </NavItemContent>
      </NavLink>

      {user.role === "admin" && (
        <>
          <div className="mb-1 mt-6 px-3 text-[10px] font-semibold uppercase tracking-wider text-slate-500">
            Admin Suite
          </div>
          <NavLink to="/admin/templates" className={navItemClass} onClick={onNavigate}>
            <NavItemContent>
              <LayoutTemplate className="h-4 w-4 text-purple-400 group-hover:text-white transition-colors" aria-hidden="true" />
              Templates
            </NavItemContent>
          </NavLink>
          <NavLink to="/admin/brand-assets" className={navItemClass} onClick={onNavigate}>
            <NavItemContent>
              <Image className="h-4 w-4 text-amber-400 group-hover:text-white transition-colors" aria-hidden="true" />
              Brand Assets
            </NavItemContent>
          </NavLink>
          <NavLink to="/admin/agent-settings" className={navItemClass} onClick={onNavigate}>
            <NavItemContent>
              <Settings className="h-4 w-4 text-emerald-400 group-hover:text-white transition-colors" aria-hidden="true" />
              AI Providers
            </NavItemContent>
          </NavLink>
        </>
      )}
    </nav>
  );
}

function SidebarFooter({ user }: { user: User }) {
  const logout = useLogout();
  return (
    <div className="mt-auto flex items-center justify-between gap-3 border-t border-white/10 bg-white/[0.02] px-4 py-3.5">
      <div className="flex items-center gap-2.5 min-w-0">
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-gradient-to-tr from-brand-primary to-purple-600 text-xs font-bold text-white shadow-inner">
          {user.email.slice(0, 2).toUpperCase()}
        </div>
        <div className="min-w-0">
          <p className="truncate text-xs font-medium text-slate-200">{user.email}</p>
          <p className="text-[10px] capitalize text-slate-500">{user.role} workspace</p>
        </div>
      </div>
      <button
        type="button"
        onClick={() => logout.mutate()}
        aria-label="Log out"
        className="shrink-0 rounded-lg p-2 text-slate-400 transition-colors hover:bg-white/10 hover:text-white"
      >
        <LogOut className="h-4 w-4" aria-hidden="true" />
      </button>
    </div>
  );
}

// Persistent left nav on desktop; a slim top bar + slide-in Modal on mobile
export function Sidebar({ user }: Props) {
  const [mobileNavOpen, setMobileNavOpen] = useState(false);

  return (
    <>
      {/* Desktop sidebar */}
      <aside className="hidden md:sticky md:top-0 md:flex md:h-screen md:w-64 md:shrink-0 md:flex-col md:border-r md:border-white/10 md:bg-[#0C0F19]/90 md:backdrop-blur-2xl">
        <div className="px-5 py-6">
          <Logo />
        </div>
        <NavItems user={user} />
        <SidebarFooter user={user} />
      </aside>

      {/* Mobile top bar */}
      <div className="sticky top-0 z-30 flex items-center justify-between border-b border-white/10 bg-[#0C0F19]/90 px-4 py-3.5 backdrop-blur-xl md:hidden">
        <Logo />
        <button
          type="button"
          onClick={() => setMobileNavOpen(true)}
          aria-label="Open menu"
          className="rounded-xl p-2 text-slate-300 hover:bg-white/10"
        >
          <Menu className="h-5 w-5" aria-hidden="true" />
        </button>
      </div>

      <Modal side="left" open={mobileNavOpen} onOpenChange={setMobileNavOpen} title="Navigation">
        <div className="-mx-6 flex flex-col pt-2">
          <NavItems user={user} onNavigate={() => setMobileNavOpen(false)} />
          <div className="mt-8">
            <SidebarFooter user={user} />
          </div>
        </div>
      </Modal>
    </>
  );
}
