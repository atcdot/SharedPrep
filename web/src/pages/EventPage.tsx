import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { ItemCard } from "@/components/ItemCard";
import { ItemForm } from "@/components/ItemForm";
import { ParticipantList } from "@/components/ParticipantList";
import { TelegramLogin } from "@/components/TelegramLogin";
import { eventClient, itemClient, participantClient } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import type { Event } from "@/gen/event/v1/event_pb";
import type { Item } from "@/gen/item/v1/item_pb";
import type { Participant } from "@/gen/participant/v1/participant_pb";

const clientId = import.meta.env.VITE_TELEGRAM_CLIENT_ID as string;

export function EventPage() {
  const { shareCode } = useParams<{ shareCode: string }>();
  const navigate = useNavigate();
  const { user, loading: authLoading, login } = useAuth();

  const [event, setEvent] = useState<Event | null>(null);
  const [items, setItems] = useState<Item[]>([]);
  const [participants, setParticipants] = useState<Participant[]>([]);
  const [myParticipantId, setMyParticipantId] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!shareCode || authLoading) return;
    loadData();
  }, [shareCode, authLoading, user]);

  async function loadData() {
    setError("");
    try {
      const eventResp = await eventClient.getEvent({ shareCode: shareCode! });
      const ev = eventResp.event!;
      setEvent(ev);
      setMyParticipantId(eventResp.myParticipantId);

      const [itemsResp, partsResp] = await Promise.all([
        itemClient.listItems({ eventId: ev.id }),
        participantClient.listParticipants({ eventId: ev.id }),
      ]);
      setItems(itemsResp.items);
      setParticipants(partsResp.participants);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Event not found");
    } finally {
      setLoading(false);
    }
  }

  async function handleJoinAndReload() {
    if (!shareCode) return;
    await participantClient.joinEvent({ shareCode });
    await loadData();
  }

  async function handleAddItem(title: string, quantity: number) {
    if (!event) return;
    const resp = await itemClient.createItem({ eventId: event.id, title, quantity });
    setItems((prev) => [...prev, resp.item!]);
  }

  async function handleClaim(itemId: string) {
    const resp = await itemClient.claimItem({ itemId });
    setItems((prev) => prev.map((i) => (i.id === itemId ? resp.item! : i)));
  }

  async function handleUnclaim(itemId: string) {
    const resp = await itemClient.unclaimItem({ itemId });
    setItems((prev) => prev.map((i) => (i.id === itemId ? resp.item! : i)));
  }

  async function handleAssign(itemId: string, participantId: string) {
    const resp = await itemClient.assignItem({ itemId, participantId });
    setItems((prev) => prev.map((i) => (i.id === itemId ? resp.item! : i)));
  }

  async function handleDelete(itemId: string) {
    await itemClient.deleteItem({ itemId });
    setItems((prev) => prev.filter((i) => i.id !== itemId));
  }

  if (loading || authLoading) {
    return <p className="text-muted-foreground">Loading...</p>;
  }

  if (error || !event) {
    return (
      <div className="flex flex-col items-center gap-4 pt-8">
        <p className="text-destructive">{error || "Event not found"}</p>
        <button onClick={() => navigate("/")} className="text-sm underline">
          Go home
        </button>
      </div>
    );
  }

  function handleCopyLink() {
    navigator.clipboard.writeText(window.location.href).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  }

  const canModify = !!user && !!myParticipantId;
  const isAuthor = canModify && participants.some(
    (p) => p.id === myParticipantId && p.isAuthor,
  );

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">{event.title}</h1>
          {event.description && (
          <p className="mt-1 text-muted-foreground">{event.description}</p>
        )}
        {event.date && (
          <p className="mt-1 text-sm text-muted-foreground">{event.date}</p>
        )}
        {event.link && (
          <a
            href={event.link}
            target="_blank"
            rel="noopener noreferrer"
            className="mt-1 inline-block text-sm text-primary underline"
          >
            {event.link}
          </a>
        )}
        <p className="mt-2 text-sm">
          Created by <span className="font-medium">{event.authorName}</span>
        </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          className={`shrink-0 ${copied ? "border-green-500 text-green-600" : ""}`}
          onClick={handleCopyLink}
        >
          {copied ? "Copied!" : "Copy join link"}
        </Button>
      </div>

      <Separator />

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Participants</CardTitle>
        </CardHeader>
        <CardContent>
          <ParticipantList participants={participants} />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Items</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          {canModify && <ItemForm onSubmit={handleAddItem} />}
          {!user && (
            <div className="flex flex-col items-center gap-3">
              <TelegramLogin clientId={clientId} onAuth={login} />
            </div>
          )}
          {user && !myParticipantId && (
            <p className="text-sm text-muted-foreground">
              <button onClick={handleJoinAndReload} className="underline">
                Join this event
              </button>{" "}
              to add items and claim
            </p>
          )}
          <div className="flex flex-col gap-2">
            {items.length === 0 && (
              <p className="text-sm text-muted-foreground">No items yet</p>
            )}
            {items.map((item) => (
              <ItemCard
                key={item.id}
                item={item}
                participants={participants}
                isAuthor={isAuthor}
                canModify={canModify}
                onClaim={handleClaim}
                onUnclaim={handleUnclaim}
                onAssign={handleAssign}
                onDelete={handleDelete}
              />
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
