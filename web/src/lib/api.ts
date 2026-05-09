import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

import { EventService } from "@/gen/event/v1/event_pb";
import { ParticipantService } from "@/gen/participant/v1/participant_pb";
import { ItemService } from "@/gen/item/v1/item_pb";

const transport = createConnectTransport({
  baseUrl: window.location.origin,
  useBinaryFormat: false,
  fetch: (input, init) =>
    globalThis.fetch(input, { ...init, credentials: "include" }),
});

export const eventClient = createClient(EventService, transport);
export const participantClient = createClient(ParticipantService, transport);
export const itemClient = createClient(ItemService, transport);
