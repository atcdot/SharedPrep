import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import type { Item } from "@/gen/item/v1/item_pb";
import type { Participant } from "@/gen/participant/v1/participant_pb";

interface ItemCardProps {
  item: Item;
  participants: Participant[];
  isAuthor: boolean;
  canModify: boolean;
  onClaim: (id: string) => void;
  onUnclaim: (id: string) => void;
  onDelete: (id: string) => void;
  onAssign: (itemId: string, participantId: string) => void;
}

export function ItemCard({
  item,
  participants,
  isAuthor,
  canModify,
  onClaim,
  onUnclaim,
  onDelete,
  onAssign,
}: ItemCardProps) {
  const [assigning, setAssigning] = useState(false);
  const claimed = !!item.assignedParticipantId;

  function handleAssign(e: React.ChangeEvent<HTMLSelectElement>) {
    const pid = e.target.value;
    if (pid) {
      onAssign(item.id, pid);
      setAssigning(false);
    }
  }

  return (
    <div className="flex items-center justify-between rounded-lg border p-3">
      <div className="flex items-center gap-3">
        <div>
          <div className="flex items-center gap-2">
            <span className="font-medium">{item.title}</span>
            {item.quantity > 1 && (
              <Badge variant="secondary">x{item.quantity}</Badge>
            )}
          </div>
          <span className="text-sm text-muted-foreground">
            {claimed ? item.assignedParticipantName : "Unassigned"}
          </span>
        </div>
      </div>

      <div className="flex items-center gap-2">
        {canModify && (
          claimed ? (
            <Button variant="outline" size="sm" onClick={() => onUnclaim(item.id)}>
              Unclaim
            </Button>
          ) : (
            <Button size="sm" onClick={() => onClaim(item.id)}>
              Claim
            </Button>
        ))}

        {isAuthor && (
          <>
            {assigning ? (
              <select
                className="h-8 rounded-md border bg-background px-2 text-sm"
                onChange={handleAssign}
                onBlur={() => setAssigning(false)}
                value=""
                autoFocus
              >
                <option value="" disabled>
                  Select participant
                </option>
                {participants.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.displayName || "Unknown"}
                  </option>
                ))}
              </select>
            ) : (
              <Button variant="outline" size="sm" onClick={() => setAssigning(true)}>
                Assign
              </Button>
            )}
          </>
        )}

        {canModify && (
          <Button
            variant="ghost"
            size="sm"
            className="text-destructive"
            onClick={() => onDelete(item.id)}
          >
            Delete
          </Button>
        )}
      </div>
    </div>
  );
}
