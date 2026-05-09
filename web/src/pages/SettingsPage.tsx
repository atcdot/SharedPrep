import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/lib/auth";

export function SettingsPage() {
  const { user, updateProfile } = useAuth();

  const [displayName, setDisplayName] = useState(user!.displayName ?? user!.firstName);
  const [showTelegram, setShowTelegram] = useState(user!.showTelegram);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [saved, setSaved] = useState(false);

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError("");
    setSaved(false);

    try {
      await updateProfile({
        display_name: displayName.trim(),
        show_telegram: showTelegram,
      });
      setSaved(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="flex flex-col items-center gap-8 pt-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Settings</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSave} className="flex flex-col gap-4">
            <div className="space-y-2">
              <Label htmlFor="displayName">Display name</Label>
              <Input
                id="displayName"
                placeholder={user!.firstName}
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
              />
              <p className="text-xs text-muted-foreground">
                This name will be shown to other participants instead of your Telegram name.
              </p>
            </div>

            <div className="flex items-center justify-between rounded-lg border p-3">
              <div className="space-y-0.5">
                <Label>Show Telegram link</Label>
                <p className="text-xs text-muted-foreground">
                  Let other participants message you on Telegram
                </p>
              </div>
              <input
                type="checkbox"
                checked={showTelegram}
                onChange={(e) => setShowTelegram(e.target.checked)}
                className="h-4 w-4"
              />
            </div>

            {error && <p className="text-sm text-destructive">{error}</p>}
            {saved && <p className="text-sm text-green-600">Saved</p>}

            <Button type="submit" disabled={saving}>
              {saving ? "Saving..." : "Save"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
