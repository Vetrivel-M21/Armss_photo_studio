import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { PaletteHex, StyleProfile, Template } from "../../api/types";
import { templatesApi } from "../../api/templates";
import { Button } from "../common/Button";
import { Input } from "../common/Input";
import { TextArea } from "../common/TextArea";
import { ErrorBanner } from "../common/ErrorBanner";
import { Modal } from "../common/Modal";
import { useToast } from "../common/Toast";
import { ApiError } from "../../api/client";

interface Props {
  template: Template;
  onClose: () => void;
}

const emptyProfile: StyleProfile = {
  summary: "",
  palette: [],
  typography: "",
  layoutPatterns: "",
  imagery: "",
  finish: "",
  tone: "",
  logoTreatment: "",
  fixedElements: [],
  consistency: "",
};

function paletteToText(palette: PaletteHex[]): string {
  return palette.map((p) => `${p.hex},${p.role}`).join("\n");
}

function paletteFromText(text: string): PaletteHex[] {
  return text
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const [hex, ...rest] = line.split(",");
      return { hex: hex.trim(), role: rest.join(",").trim() };
    });
}

// Admin correction beats a model's guess (spec §8) — every field here is a
// plain, hand-editable form, not a JSON blob, so a non-technical admin can fix
// what the Style Analyst got wrong.
export function TemplateEditDrawer({ template, onClose }: Props) {
  const [profile, setProfile] = useState<StyleProfile>(template.styleProfile ?? emptyProfile);
  const [paletteText, setPaletteText] = useState(paletteToText(profile.palette));
  const [tags, setTags] = useState(template.occasionTags.join(", "));
  const [season, setSeason] = useState(template.season ?? "");
  const [isActive, setIsActive] = useState(template.isActive);
  const queryClient = useQueryClient();
  const { toast } = useToast();

  const save = useMutation({
    mutationFn: () =>
      templatesApi.patch(template.id, {
        styleProfile: { ...profile, palette: paletteFromText(paletteText) },
        occasionTags: tags.split(",").map((t) => t.trim()).filter(Boolean),
        season: season || undefined,
        isActive,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-templates"] });
      toast({ title: "Template saved", variant: "success" });
      onClose();
    },
  });

  function field(label: string, key: keyof StyleProfile) {
    return (
      <div className="space-y-1">
        <label className="text-xs font-medium text-neutral-500">{label}</label>
        <Input value={profile[key] as string} onChange={(e) => setProfile({ ...profile, [key]: e.target.value })} />
      </div>
    );
  }

  return (
    <Modal
      open
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
      side="right"
      title={template.title}
      disableOutsideClose
      footer={
        <>
          <Button variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={() => save.mutate()} disabled={save.isPending}>
            {save.isPending ? "Saving..." : "Save"}
          </Button>
        </>
      }
    >
      {save.isError && (
        <ErrorBanner message={save.error instanceof ApiError ? save.error.message : "Could not save this template."} />
      )}

      <div className="space-y-4">
        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-1">
            <label className="text-xs font-medium text-neutral-500">Occasion tags (comma separated)</label>
            <Input value={tags} onChange={(e) => setTags(e.target.value)} />
          </div>
          <div className="space-y-1">
            <label className="text-xs font-medium text-neutral-500">Season</label>
            <Input value={season} onChange={(e) => setSeason(e.target.value)} />
          </div>
        </div>

        <label className="flex items-center gap-2 text-sm text-neutral-700">
          <input type="checkbox" checked={isActive} onChange={(e) => setIsActive(e.target.checked)} />
          Active in the gallery
        </label>

        <hr className="border-neutral-100" />
        <h3 className="text-sm font-semibold text-neutral-700">Style profile</h3>

        <div className="space-y-1">
          <label className="text-xs font-medium text-neutral-500">Summary</label>
          <TextArea rows={5} value={profile.summary} onChange={(e) => setProfile({ ...profile, summary: e.target.value })} />
        </div>

        <div className="space-y-1">
          <label className="text-xs font-medium text-neutral-500">Palette (one "hex,role" per line)</label>
          <TextArea rows={3} value={paletteText} onChange={(e) => setPaletteText(e.target.value)} />
        </div>

        {field("Typography", "typography")}
        {field("Layout patterns", "layoutPatterns")}
        {field("Imagery", "imagery")}
        {field("Finish", "finish")}
        {field("Tone", "tone")}
        {field("Logo treatment", "logoTreatment")}

        <div className="space-y-1">
          <label className="text-xs font-medium text-neutral-500">Fixed elements (comma separated)</label>
          <Input
            value={profile.fixedElements.join(", ")}
            onChange={(e) =>
              setProfile({ ...profile, fixedElements: e.target.value.split(",").map((s) => s.trim()).filter(Boolean) })
            }
          />
        </div>

        <div className="space-y-1">
          <label className="text-xs font-medium text-neutral-500">Consistency</label>
          <TextArea rows={2} value={profile.consistency} onChange={(e) => setProfile({ ...profile, consistency: e.target.value })} />
        </div>
      </div>
    </Modal>
  );
}
