import { useState, type FormEvent } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

interface ItemFormProps {
  onSubmit: (title: string, quantity: number) => Promise<void>;
}

export function ItemForm({ onSubmit }: ItemFormProps) {
  const [title, setTitle] = useState("");
  const [quantity, setQuantity] = useState(1);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!title.trim()) return;

    setLoading(true);
    try {
      await onSubmit(title.trim(), quantity);
      setTitle("");
      setQuantity(1);
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex items-end gap-2">
      <div className="flex-1">
        <Input
          placeholder="Add item..."
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          required
        />
      </div>
      <div className="w-20">
        <Input
          type="number"
          min={1}
          value={quantity}
          onChange={(e) => setQuantity(Number(e.target.value))}
        />
      </div>
      <Button type="submit" disabled={loading}>
        Add
      </Button>
    </form>
  );
}
