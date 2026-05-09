# SharedPrep API Examples

All endpoints use Connect-RPC with JSON over HTTP. Base URL: `http://localhost:8080`.

## Create Event

```
POST /event.v1.EventService/CreateEvent
Content-Type: application/json

{
  "title": "Шашлыки",
  "description": "На даче у Сергея",
  "date": "2025-06-15",
  "authorName": "Богдан"
}
```

Response:
```json
{
  "event": {
    "id": "uuid",
    "shareCode": "MZ896A",
    "title": "Шашлыки",
    "description": "На даче у Сергея",
    "date": "2025-06-15",
    "authorName": "Богдан",
    "createdAt": "2025-06-01T12:00:00Z"
  },
  "participantToken": "uuid"
}
```

> Save `participantToken` in cookie `participant_token` for subsequent requests.

## Get Event

```
POST /event.v1.EventService/GetEvent
Content-Type: application/json

{"shareCode": "MZ896A"}
```

## Update Event

```
POST /event.v1.EventService/UpdateEvent
Content-Type: application/json

{
  "eventId": "uuid",
  "title": "Шашлыки v2",
  "description": "Перенесли на озеро"
}
```

## Delete Event

```
POST /event.v1.EventService/DeleteEvent
Content-Type: application/json

{"eventId": "uuid"}
```

## Join Event

```
POST /participant.v1.ParticipantService/JoinEvent
Content-Type: application/json

{
  "shareCode": "MZ896A",
  "name": "Алексей"
}
```

Response includes `participantToken` — save in cookie.

## List Participants

```
POST /participant.v1.ParticipantService/ListParticipants
Content-Type: application/json

{"eventId": "uuid"}
```

## Create Item

```
POST /item.v1.ItemService/CreateItem
Content-Type: application/json
Cookie: participant_token=<token>

{
  "eventId": "uuid",
  "title": "Мясо",
  "quantity": 3
}
```

## Update Item

```
POST /item.v1.ItemService/UpdateItem
Content-Type: application/json

{
  "itemId": "uuid",
  "title": "Курица",
  "quantity": 5
}
```

## Delete Item

```
POST /item.v1.ItemService/DeleteItem
Content-Type: application/json

{"itemId": "uuid"}
```

## Claim Item

```
POST /item.v1.ItemService/ClaimItem
Content-Type: application/json
Cookie: participant_token=<token>

{"itemId": "uuid"}
```

## Unclaim Item

```
POST /item.v1.ItemService/UnclaimItem
Content-Type: application/json

{"itemId": "uuid"}
```

## List Items

```
POST /item.v1.ItemService/ListItems
Content-Type: application/json

{"eventId": "uuid"}
```

Response includes `assignedParticipantId` and `assignedParticipantName` when claimed.
