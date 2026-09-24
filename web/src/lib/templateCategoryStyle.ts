import {
  Building,
  Building2,
  Clapperboard,
  Circle,
  Cpu,
  Dumbbell,
  GraduationCap,
  Gem,
  Heart,
  Landmark,
  Megaphone,
  Package,
  PartyPopper,
  ShieldCheck,
  Share2,
  Sparkles,
  Sun,
  TrendingUp,
  Users,
  UtensilsCrossed,
  type LucideIcon,
} from "lucide-react";
import type { Template } from "../api/types";

export interface CategoryStyle {
  category: string;
  icon: LucideIcon;
  gradientClass: string;
}

// The 20 starter-library categories (ai_poster_template_library.md) mapped to
// a decorative icon + a gradient built only from the brand palette (plus
// neutral-900 for the one deliberately dark/moody category, Cinematic).
const CATEGORY_TABLE: CategoryStyle[] = [
  { category: "Indian Festival", icon: Sparkles, gradientClass: "bg-gradient-to-br from-brand-primary via-brand-secondary to-brand-highlight" },
  { category: "Corporate", icon: Building2, gradientClass: "bg-gradient-to-r from-brand-primary to-brand-secondary" },
  { category: "Business", icon: TrendingUp, gradientClass: "bg-gradient-to-r from-brand-secondary to-brand-primary" },
  { category: "Luxury", icon: Gem, gradientClass: "bg-gradient-to-b from-brand-primary via-brand-secondary to-brand-highlight" },
  { category: "Event", icon: PartyPopper, gradientClass: "bg-gradient-to-tr from-brand-secondary to-brand-highlight" },
  { category: "Product", icon: Package, gradientClass: "bg-gradient-to-tr from-brand-accent to-brand-highlight" },
  { category: "Technology", icon: Cpu, gradientClass: "bg-gradient-to-bl from-brand-primary to-brand-secondary" },
  { category: "Cultural", icon: Landmark, gradientClass: "bg-gradient-to-tr from-brand-primary via-brand-secondary to-brand-accent" },
  { category: "Wedding", icon: Heart, gradientClass: "bg-gradient-to-br from-brand-secondary via-brand-accent to-brand-highlight" },
  { category: "Social Media", icon: Share2, gradientClass: "bg-gradient-to-br from-brand-accent via-brand-highlight to-brand-secondary" },
  { category: "Promotional", icon: Megaphone, gradientClass: "bg-gradient-to-r from-brand-secondary via-brand-highlight to-brand-accent" },
  { category: "Minimal", icon: Circle, gradientClass: "bg-gradient-to-br from-brand-highlight/20 to-white" },
  { category: "Cinematic", icon: Clapperboard, gradientClass: "bg-gradient-to-br from-neutral-900 via-brand-primary to-brand-secondary" },
  { category: "Seasonal", icon: Sun, gradientClass: "bg-gradient-to-b from-brand-accent to-brand-highlight" },
  { category: "Community", icon: Users, gradientClass: "bg-gradient-to-br from-brand-primary to-brand-secondary" },
  { category: "Education", icon: GraduationCap, gradientClass: "bg-gradient-to-br from-brand-secondary to-brand-highlight" },
  { category: "Real Estate", icon: Building, gradientClass: "bg-gradient-to-b from-brand-primary to-brand-accent" },
  { category: "Food", icon: UtensilsCrossed, gradientClass: "bg-gradient-to-b from-brand-secondary to-brand-highlight" },
  { category: "Fitness", icon: Dumbbell, gradientClass: "bg-gradient-to-tr from-brand-primary via-brand-secondary to-brand-accent" },
  { category: "Professional Services", icon: ShieldCheck, gradientClass: "bg-gradient-to-br from-brand-primary to-brand-accent" },
];

// Keyword aliases for matching against occasionTags/title — lowercase.
const CATEGORY_KEYWORDS: Record<string, string[]> = {
  "Indian Festival": ["festival", "diwali", "holi", "navratri", "pongal", "onam", "rangoli", "diya"],
  Corporate: ["corporate", "office", "meeting"],
  Business: ["business", "startup", "enterprise"],
  Luxury: ["luxury", "premium", "gold", "marble"],
  Event: ["event", "party", "celebration", "confetti"],
  Product: ["product", "hero", "launch"],
  Technology: ["technology", "tech", "futuristic", "digital"],
  Cultural: ["cultural", "culture", "heritage", "temple", "traditional"],
  Wedding: ["wedding", "marriage", "bride", "groom"],
  "Social Media": ["social media", "social", "instagram", "post"],
  Promotional: ["promotional", "promo", "sale", "discount", "offer"],
  Minimal: ["minimal", "ultra minimal", "clean"],
  Cinematic: ["cinematic", "movie", "film", "night", "neon"],
  Seasonal: ["seasonal", "season", "sunrise", "summer", "winter"],
  Community: ["community", "collage", "group"],
  Education: ["education", "academic", "school", "graduation"],
  "Real Estate": ["real estate", "property", "housing"],
  Food: ["food", "restaurant", "cuisine", "photography"],
  Fitness: ["fitness", "gym", "workout", "energy"],
  "Professional Services": ["professional services", "professional", "services", "consulting"],
};

function findByKeyword(haystack: string): CategoryStyle | undefined {
  const lower = haystack.toLowerCase();
  for (const style of CATEGORY_TABLE) {
    const keywords = CATEGORY_KEYWORDS[style.category] ?? [];
    if (keywords.some((kw) => lower.includes(kw))) return style;
  }
  return undefined;
}

function hashFallback(id: string): CategoryStyle {
  const sum = Array.from(id).reduce((acc, ch) => acc + ch.charCodeAt(0), 0);
  return CATEGORY_TABLE[sum % CATEGORY_TABLE.length];
}

// Resolves a category style for a template so every gallery fallback tile
// gets a themed gradient + icon — never flat gray, regardless of how
// complete the template's tagging is.
export function resolveCategoryStyle(template: Pick<Template, "id" | "title" | "occasionTags"> & { category?: string }): CategoryStyle {
  if (template.category) {
    const exact = CATEGORY_TABLE.find((c) => c.category.toLowerCase() === template.category!.toLowerCase());
    if (exact) return exact;
    const byKeyword = findByKeyword(template.category);
    if (byKeyword) return byKeyword;
  }

  const tagsAndTitle = [...(template.occasionTags ?? []), template.title ?? ""].join(" ");
  const byKeyword = findByKeyword(tagsAndTitle);
  if (byKeyword) return byKeyword;

  return hashFallback(template.id);
}
