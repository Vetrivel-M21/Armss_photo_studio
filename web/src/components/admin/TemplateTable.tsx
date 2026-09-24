import type { Template } from "../../api/types";
import { Card } from "../common/Card";
import { Badge } from "../common/Badge";

interface Props {
  templates: Template[];
  onEdit: (template: Template) => void;
}

export function TemplateTable({ templates, onEdit }: Props) {
  return (
    <Card padding="none" className="overflow-x-auto border-white/10 bg-[#0E121E]/90 backdrop-blur-xl">
      <table className="w-full text-left text-sm">
        <thead className="border-b border-white/10 bg-white/[0.02] text-xs uppercase tracking-wider text-slate-400">
          <tr>
            <th className="px-4 py-3">Template</th>
            <th className="px-4 py-3">Occasion Tags</th>
            <th className="px-4 py-3">Season</th>
            <th className="px-4 py-3">Style Profile</th>
            <th className="px-4 py-3">Active</th>
            <th className="px-4 py-3 text-right">Actions</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-white/5">
          {templates.map((t) => (
            <tr key={t.id} className="transition-colors hover:bg-white/[0.03]">
              <td className="flex items-center gap-3 px-4 py-3">
                {t.thumbUrl || t.imageUrl ? (
                  <img src={t.thumbUrl || t.imageUrl} alt="" className="h-11 w-9 rounded-lg object-cover border border-white/10" />
                ) : (
                  <span className="flex h-11 w-9 items-center justify-center rounded-lg border border-white/10 bg-white/[0.04] text-[10px] font-bold text-slate-400">
                    Prompt
                  </span>
                )}
                <span className="font-semibold text-white tracking-tight">{t.title}</span>
              </td>
              <td className="px-4 py-3 text-slate-300">{t.occasionTags.join(", ") || "—"}</td>
              <td className="px-4 py-3 text-slate-300">{t.season || "—"}</td>
              <td className="px-4 py-3">
                <Badge tone={t.styleProfileStatus === "ready" ? "success" : "warning"}>
                  {t.styleProfileStatus === "ready" ? "Ready" : "Needs Review"}
                </Badge>
              </td>
              <td className="px-4 py-3 text-slate-300 font-medium">{t.isActive ? "Yes" : "No"}</td>
              <td className="px-4 py-3 text-right">
                <button
                  onClick={() => onEdit(t)}
                  className="rounded-lg px-2.5 py-1 text-xs font-semibold text-brand-accent hover:bg-brand-primary/10 hover:text-white transition-colors"
                >
                  Edit Profile
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </Card>
  );
}
