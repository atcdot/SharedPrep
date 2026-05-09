import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { TelegramLogin } from "@/components/TelegramLogin";
import { participantClient } from "@/lib/api";
import { useAuth } from "@/lib/auth";

const clientId = import.meta.env.VITE_TELEGRAM_CLIENT_ID as string;

export function JoinPage() {
  const { shareCode } = useParams<{ shareCode: string }>();
  const navigate = useNavigate();
  const { user, loading: authLoading, login } = useAuth();

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleJoin() {
    if (!shareCode) return;
    setLoading(true);
    setError("");

    try {
      await participantClient.joinEvent({ shareCode });
      navigate(`/event/${shareCode}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to join event");
    } finally {
      setLoading(false);
    }
  }

  if (authLoading) {
    return <p className="text-muted-foreground pt-8 text-center">Loading...</p>;
  }

  if (!user) {
    return (
      <div className="flex flex-col items-center gap-8 pt-8">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Join Event</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col items-center gap-4">
            <p className="text-muted-foreground">Log in with Telegram to join</p>
            <TelegramLogin clientId={clientId} onAuth={login} />
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center gap-8 pt-8">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Join Event</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col items-center gap-4">
          {error && <p className="text-sm text-destructive">{error}</p>}
          <Button onClick={handleJoin} disabled={loading}>
            {loading ? "Joining..." : `Join as ${user.firstName}`}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
