import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { eventClient } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { TelegramLogin } from "@/components/TelegramLogin";
import type { Event } from "@/gen/event/v1/event_pb";

const clientId = import.meta.env.VITE_TELEGRAM_CLIENT_ID as string;

export function HomePage() {
  const navigate = useNavigate();
  const { user, loading: authLoading, login } = useAuth();

  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (authLoading) return;
    if (!user) return;
    loadEvents();
  }, [user, authLoading]);

  async function loadEvents() {
    try {
      const resp = await eventClient.listMyEvents({});
      setEvents(resp.events);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load events");
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
        <div className="text-center">
          <h1 className="text-3xl font-bold tracking-tight">
            Plan your next event together
          </h1>
          <p className="mt-2 text-muted-foreground">
            Create a shared list, invite participants, assign items
          </p>
        </div>
        <TelegramLogin clientId={clientId} onAuth={login} />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 pt-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">My Events</h1>
        <Button onClick={() => navigate("/create")}>Create Event</Button>
      </div>

      {error && <p className="text-destructive">{error}</p>}

      {loading ? (
        <p className="text-muted-foreground">Loading...</p>
      ) : events.length === 0 ? (
        <p className="text-muted-foreground">
          You haven't joined any events yet.{" "}
          <Link to="/create" className="underline">
            Create one
          </Link>
          .
        </p>
      ) : (
        <div className="flex flex-col gap-3">
          {events.map((ev) => (
            <Link key={ev.id} to={`/event/${ev.shareCode}`}>
              <Card className="transition-colors hover:bg-muted/50">
                <CardContent className="flex flex-col gap-1 py-4">
                  <span className="font-medium">{ev.title}</span>
                  {ev.description && (
                    <span className="text-sm text-muted-foreground line-clamp-1">
                      {ev.description}
                    </span>
                  )}
                  <div className="flex gap-3 text-xs text-muted-foreground">
                    {ev.date && <span>{ev.date}</span>}
                    <span>by {ev.authorName}</span>
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
