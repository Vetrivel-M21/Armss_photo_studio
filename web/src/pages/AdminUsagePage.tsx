import { useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Activity,
  Coins,
  Cpu,
  Edit2,
  FileImage,
  FolderKanban,
  Save,
} from "lucide-react";
import { usageApi, type UpdateLimitsPayload, type UserUsage } from "../api/usage";
import { Card, CardHeader } from "../components/common/Card";
import { Badge } from "../components/common/Badge";
import { Button } from "../components/common/Button";
import { Input } from "../components/common/Input";
import { Spinner } from "../components/common/Spinner";
import { ErrorBanner } from "../components/common/ErrorBanner";
import { Modal } from "../components/common/Modal";
import { useToast } from "../components/common/Toast";
import { cn } from "../lib/cn";

export function AdminUsagePage() {
  const queryClient = useQueryClient();
  const { toast } = useToast();

  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ["admin-usage"],
    queryFn: usageApi.getOverview,
  });

  // Global limits form state
  const [globalCalls, setGlobalCalls] = useState(30);
  const [globalRenders, setGlobalRenders] = useState(6);
  const [globalCapDollars, setGlobalCapDollars] = useState(15);

  useEffect(() => {
    if (data?.globalLimit) {
      setGlobalCalls(data.globalLimit.maxApiCallsPerHour);
      setGlobalRenders(data.globalLimit.maxRendersPerHour);
      setGlobalCapDollars(Math.round(data.globalLimit.monthlySpendCapCents / 100));
    }
  }, [data]);

  const saveGlobal = useMutation({
    mutationFn: (payload: UpdateLimitsPayload) => usageApi.updateGlobalLimits(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-usage"] });
      toast({ title: "Global rate limits updated successfully", variant: "success" });
    },
    onError: (err) => {
      toast({ title: err instanceof Error ? err.message : "Failed to update global limits", variant: "error" });
    },
  });

  // User limit modal state
  const [editingUser, setEditingUser] = useState<UserUsage | null>(null);
  const [userCalls, setUserCalls] = useState(30);
  const [userRenders, setUserRenders] = useState(6);
  const [userCapDollars, setUserCapDollars] = useState(15);
  const [userUnlimited, setUserUnlimited] = useState(false);

  const openUserModal = (u: UserUsage) => {
    setEditingUser(u);
    setUserCalls(u.limits.maxApiCallsPerHour);
    setUserRenders(u.limits.maxRendersPerHour);
    setUserCapDollars(Math.round(u.limits.monthlySpendCapCents / 100));
    setUserUnlimited(u.limits.isUnlimited);
  };

  const saveUser = useMutation({
    mutationFn: (payload: UpdateLimitsPayload) => {
      if (!editingUser) throw new Error("No user selected");
      return usageApi.updateUserLimits(editingUser.userId, payload);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-usage"] });
      toast({ title: `Quotas for ${editingUser?.email} updated`, variant: "success" });
      setEditingUser(null);
    },
    onError: (err) => {
      toast({ title: err instanceof Error ? err.message : "Failed to update user quota", variant: "error" });
    },
  });

  const handleSaveGlobal = () => {
    saveGlobal.mutate({
      maxApiCallsPerHour: Number(globalCalls),
      maxRendersPerHour: Number(globalRenders),
      monthlySpendCapCents: Math.round(Number(globalCapDollars) * 100),
    });
  };

  const handleSaveUser = () => {
    saveUser.mutate({
      maxApiCallsPerHour: Number(userCalls),
      maxRendersPerHour: Number(userRenders),
      monthlySpendCapCents: Math.round(Number(userCapDollars) * 100),
      isUnlimited: userUnlimited,
    });
  };

  if (isLoading) {
    return (
      <div className="flex justify-center py-20">
        <Spinner label="Loading usage statistics..." />
      </div>
    );
  }

  if (isError || !data) {
    return (
      <ErrorBanner
        message={error instanceof Error ? error.message : "Failed to load usage data."}
        onRetry={() => refetch()}
      />
    );
  }

  const totalTokens = data.totalTokensIn + data.totalTokensOut;
  const totalCost = (data.totalCostCents / 100).toFixed(2);

  return (
    <div className="space-y-8">
      {/* Header bar */}
      <div className="flex flex-wrap items-center justify-between gap-4 rounded-3xl border border-white/10 bg-gradient-to-r from-[#141824]/95 via-[#161B28]/90 to-[#1D162B]/95 p-6 shadow-glass backdrop-blur-xl">
        <div className="flex items-center gap-3.5">
          <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-primary via-brand-secondary to-purple-600 shadow-glow">
            <Activity className="h-6 w-6 text-white" aria-hidden="true" />
          </div>
          <div>
            <h1 className="text-xl font-extrabold tracking-tight text-white sm:text-2xl">
              Usage &amp; Rate Limits
            </h1>
            <p className="mt-0.5 text-xs text-slate-400">
              Real-time API tracking, cost meters, and hourly quota controls
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Badge tone="success" size="md">
            <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse mr-1" />
            Live Metering Active
          </Badge>
        </div>
      </div>

      {/* Top Key Metrics */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card padding="md" className="border-white/10 bg-[#111420]/80">
          <div className="flex items-center justify-between">
            <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Total Tokens</span>
            <Cpu className="h-4 w-4 text-brand-primary" />
          </div>
          <p className="mt-2 text-2xl font-black text-white">{totalTokens.toLocaleString()}</p>
          <p className="mt-1 text-[11px] text-slate-400">
            {data.totalTokensIn.toLocaleString()} in · {data.totalTokensOut.toLocaleString()} out
          </p>
        </Card>

        <Card padding="md" className="border-white/10 bg-[#111420]/80">
          <div className="flex items-center justify-between">
            <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Image Generations</span>
            <FileImage className="h-4 w-4 text-emerald-400" />
          </div>
          <p className="mt-2 text-2xl font-black text-white">{data.totalRenders}</p>
          <p className="mt-1 text-[11px] text-slate-400">Flux, Ideogram &amp; DALL-E</p>
        </Card>

        <Card padding="md" className="border-white/10 bg-[#111420]/80">
          <div className="flex items-center justify-between">
            <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Est. Render Cost</span>
            <Coins className="h-4 w-4 text-amber-400" />
          </div>
          <p className="mt-2 text-2xl font-black text-amber-400">${totalCost}</p>
          <p className="mt-1 text-[11px] text-slate-400">Cumulative provider fees</p>
        </Card>

        <Card padding="md" className="border-white/10 bg-[#111420]/80">
          <div className="flex items-center justify-between">
            <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">Total Posters</span>
            <FolderKanban className="h-4 w-4 text-sky-400" />
          </div>
          <p className="mt-2 text-2xl font-black text-white">{data.totalProjects}</p>
          <p className="mt-1 text-[11px] text-slate-400">{data.users.length} active users</p>
        </Card>
      </div>

      {/* Global Default Limits Settings Card */}
      <Card padding="md" className="border-white/10 bg-[#111420]/90">
        <CardHeader
          title="Global Hourly Rate Limits"
          description="Default quotas applied to all users without a custom override. Rolling 60-minute window."
          actions={
            <Button
              onClick={handleSaveGlobal}
              loading={saveGlobal.isPending}
              loadingText="Saving..."
              className="gap-1.5 text-xs font-bold shadow-glow"
            >
              <Save className="h-3.5 w-3.5" />
              Save Defaults
            </Button>
          }
        />

        <div className="mt-4 grid gap-4 sm:grid-cols-3">
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">
              Max Pipeline Calls / Hour
            </label>
            <Input
              type="number"
              min={1}
              value={globalCalls}
              onChange={(e) => setGlobalCalls(Number(e.target.value))}
              placeholder="30"
              className="text-xs"
            />
            <p className="text-[10px] text-slate-400">
              Intake, concepts, copy &amp; art re-runs
            </p>
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">
              Max Image Renders / Hour
            </label>
            <Input
              type="number"
              min={1}
              value={globalRenders}
              onChange={(e) => setGlobalRenders(Number(e.target.value))}
              placeholder="6"
              className="text-xs"
            />
            <p className="text-[10px] text-slate-400">
              AI image synthesis operations
            </p>
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">
              Monthly Budget Cap ($)
            </label>
            <Input
              type="number"
              min={1}
              value={globalCapDollars}
              onChange={(e) => setGlobalCapDollars(Number(e.target.value))}
              placeholder="15"
              className="text-xs"
            />
            <p className="text-[10px] text-slate-400">
              Maximum monthly spend per user
            </p>
          </div>
        </div>
      </Card>

      {/* User Usage & Quotas Table */}
      <Card padding="none" className="overflow-hidden border-white/10 bg-[#111420]/80">
        <div className="p-5 border-b border-white/10 flex items-center justify-between">
          <div>
            <h3 className="text-sm font-bold text-white">User Consumption &amp; Quotas</h3>
            <p className="text-xs text-slate-400">Monitor live activity and customize individual user thresholds</p>
          </div>
          <span className="text-xs font-medium text-slate-400">
            {data.users.length} registered users
          </span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="border-b border-white/5 bg-white/[0.02] text-[10px] font-bold uppercase tracking-wider text-slate-400">
              <tr>
                <th className="py-3 px-4">User</th>
                <th className="py-3 px-4">Role</th>
                <th className="py-3 px-4">Posters</th>
                <th className="py-3 px-4">Total Tokens</th>
                <th className="py-3 px-4">Renders</th>
                <th className="py-3 px-4">Est. Spend</th>
                <th className="py-3 px-4">Hourly Quota Status</th>
                <th className="py-3 px-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {data.users.map((u) => {
                const userTokens = u.totalTokensIn + u.totalTokensOut;
                const userSpend = (u.totalCostCents / 100).toFixed(2);

                const callPct = u.limits.isUnlimited
                  ? 0
                  : Math.min(100, Math.round((u.hourApiCalls / u.limits.maxApiCallsPerHour) * 100));
                const renderPct = u.limits.isUnlimited
                  ? 0
                  : Math.min(100, Math.round((u.hourRenders / u.limits.maxRendersPerHour) * 100));

                return (
                  <tr key={u.userId} className="hover:bg-white/[0.02] transition-colors">
                    <td className="py-3 px-4">
                      <div className="font-semibold text-white">{u.email}</div>
                      <div className="font-mono text-[10px] text-slate-400">ID: {u.userId.slice(0, 8)}</div>
                    </td>
                    <td className="py-3 px-4">
                      <Badge tone={u.role === "admin" ? "brand" : "neutral"} size="sm">
                        {u.role}
                      </Badge>
                    </td>
                    <td className="py-3 px-4 text-slate-200 font-medium">{u.totalProjects}</td>
                    <td className="py-3 px-4 font-mono text-slate-300">{userTokens.toLocaleString()}</td>
                    <td className="py-3 px-4 text-slate-200">{u.totalRenders}</td>
                    <td className="py-3 px-4 font-bold text-amber-400">${userSpend}</td>
                    <td className="py-3 px-4">
                      {u.limits.isUnlimited ? (
                        <Badge tone="success" size="sm">
                          Unlimited Quota
                        </Badge>
                      ) : (
                        <div className="space-y-1.5 w-44">
                          <div>
                            <div className="flex justify-between text-[10px] text-slate-400 mb-0.5">
                              <span>Calls</span>
                              <span className="font-semibold text-slate-300">
                                {u.hourApiCalls} / {u.limits.maxApiCallsPerHour}
                              </span>
                            </div>
                            <div className="h-1.5 w-full rounded-full bg-white/10 overflow-hidden">
                              <div
                                className={cn(
                                  "h-full rounded-full transition-all duration-300",
                                  callPct >= 90 ? "bg-rose-500" : callPct >= 70 ? "bg-amber-400" : "bg-brand-primary"
                                )}
                                style={{ width: `${callPct}%` }}
                              />
                            </div>
                          </div>

                          <div>
                            <div className="flex justify-between text-[10px] text-slate-400 mb-0.5">
                              <span>Renders</span>
                              <span className="font-semibold text-slate-300">
                                {u.hourRenders} / {u.limits.maxRendersPerHour}
                              </span>
                            </div>
                            <div className="h-1.5 w-full rounded-full bg-white/10 overflow-hidden">
                              <div
                                className={cn(
                                  "h-full rounded-full transition-all duration-300",
                                  renderPct >= 90 ? "bg-rose-500" : renderPct >= 70 ? "bg-amber-400" : "bg-emerald-400"
                                )}
                                style={{ width: `${renderPct}%` }}
                              />
                            </div>
                          </div>
                        </div>
                      )}
                    </td>
                    <td className="py-3 px-4 text-right">
                      <Button
                        variant="secondary"
                        onClick={() => openUserModal(u)}
                        className="gap-1 text-[11px] py-1 px-2.5"
                      >
                        <Edit2 className="h-3 w-3" />
                        Edit Limit
                      </Button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </Card>

      {/* Edit User Quota Modal */}
      {editingUser && (
        <Modal
          open={!!editingUser}
          onOpenChange={(open) => !open && setEditingUser(null)}
          title={`Edit Quota: ${editingUser.email}`}
        >
          <div className="space-y-4 pt-2">
            <div className="flex items-center justify-between rounded-xl bg-white/[0.04] p-3 border border-white/5">
              <div>
                <span className="text-xs font-bold text-white">Unlimited Quota Mode</span>
                <p className="text-[11px] text-slate-400">Bypass all hourly rate limits for this user</p>
              </div>
              <input
                type="checkbox"
                checked={userUnlimited}
                onChange={(e) => setUserUnlimited(e.target.checked)}
                className="h-4 w-4 rounded border-white/20 bg-black/40 text-brand-primary focus:ring-brand-primary cursor-pointer"
              />
            </div>

            {!userUnlimited && (
              <div className="space-y-3">
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-slate-300">
                    Max Pipeline Calls / Hour
                  </label>
                  <Input
                    type="number"
                    min={1}
                    value={userCalls}
                    onChange={(e) => setUserCalls(Number(e.target.value))}
                    className="text-xs"
                  />
                </div>

                <div className="space-y-1">
                  <label className="text-xs font-semibold text-slate-300">
                    Max Image Renders / Hour
                  </label>
                  <Input
                    type="number"
                    min={1}
                    value={userRenders}
                    onChange={(e) => setUserRenders(Number(e.target.value))}
                    className="text-xs"
                  />
                </div>

                <div className="space-y-1">
                  <label className="text-xs font-semibold text-slate-300">
                    Monthly Spend Cap ($)
                  </label>
                  <Input
                    type="number"
                    min={1}
                    value={userCapDollars}
                    onChange={(e) => setUserCapDollars(Number(e.target.value))}
                    className="text-xs"
                  />
                </div>
              </div>
            )}

            <div className="flex justify-end gap-2 pt-3 border-t border-white/10">
              <Button
                variant="ghost"
                onClick={() => setEditingUser(null)}
                className="text-xs"
              >
                Cancel
              </Button>
              <Button
                onClick={handleSaveUser}
                loading={saveUser.isPending}
                loadingText="Updating..."
                className="text-xs font-bold shadow-glow"
              >
                Save User Quota
              </Button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  );
}
