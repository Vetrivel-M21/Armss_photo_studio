
interface FooterProps {
  className?: string;
}

export function Footer({ className = "" }: FooterProps) {
  return (
    <footer
      className={`py-4 px-6 text-center text-xs text-slate-400 border-t border-white/5 bg-black/20 backdrop-blur-sm ${className}`}
    >
      <span>Website designed and developed by </span>
      <span className="font-bold text-blue-400 drop-shadow-sm tracking-wide">
        ARMSS INFO TECH
      </span>
    </footer>
  );
}

export default Footer;
