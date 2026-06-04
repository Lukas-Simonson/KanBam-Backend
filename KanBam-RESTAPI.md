# KanBam REST API Reference

**Base URL:** `/api/v1`  
**Auth:** All endpoints (except `Auth_login`) require `Authorization: Bearer <JWT>`. `Auth_login` uses `Authorization: Basic <base64(email:password)>`.  
**Content-Type:** `application/json`  
**IDs:** All IDs are UUID v4. Timestamps are ISO 8601.

---

## Resource Hierarchy

```
Workspace
├── Members (UserRole[])
├── Tags (Tag[])
└── Boards
    ├── Columns
    │   └── Cards (via columnID)
    │       └── Comments
    └── Activity
```

---

## Authentication

### POST /auth/register
Register a new user. Returns JWT.

**Body:** `Models.Registration`
```json
{ "email": "string(email)", "name": "string(min:2, /^[a-zA-Z0-9]+$/)", "password": "string(min:8, requires upper+lower+digit+special)" }
```
**201** → `Models.UserToken`  
**409** → `EMAIL_TAKEN`

---

### POST /auth/login
Authenticate and get JWT. Uses **Basic Auth** (not Bearer).

**201** → `Models.UserToken`  
**401** → `INVALID_CREDENTIALS` | `INVALID_AUTH`

---

### POST /auth/updatePassword
Change authenticated user's password.

**Body:** `Models.UpdatePassword`
```json
{ "currentPassword": "string", "newPassword": "string(min:8, requires upper+lower+digit+special)" }
```
**200** → (empty)  
**401** → `INVALID_AUTH` | `INVALID_PASSWORD`

---

## Workspaces

### GET /workspaces
Get all workspaces accessible to the authenticated user.

**200** → `{ workspaces: Workspace[] }`

---

### POST /workspaces/create
Create a workspace owned by the authenticated user.

**Body:** `Models.WorkspaceCreation`
```json
{ "title": "string(required)", "description": "string(optional)" }
```
**201** → `Models.Workspace`

---

### GET /workspaces/{workspaceID}
Get workspace details.

**200** → `Models.Workspace`  
**401** → `INVALID_AUTH` | `NOT_AUTHORIZED`

---

### PATCH /workspaces/{workspaceID}
Update workspace title or description. Only send fields to change.

**Body:** `{ "title"?: string, "description"?: string }`  
**200** → `Models.Workspace`

---

### DELETE /workspaces/{workspaceID}
Delete workspace (must be owner).

**204** → (empty)

---

## Workspace Members

### GET /workspaces/{workspaceID}/members
List all workspace members with roles.

**200** → `{ members: UserRole[] }`

---

### POST /workspaces/{workspaceID}/members
Add a user to the workspace.

**Body:** `Models.Membership`
```json
{ "userID": "uuid", "role": "owner|admin|contributer|viewer" }
```
**204** → (empty)  
**409** → `ALREADY_IN_WORKSPACE`

---

### GET /workspaces/{workspaceID}/members/{userID}
Get a specific member's role.

**200** → `Models.UserRole`

---

### PATCH /workspaces/{workspaceID}/members/{userID}
Change a member's role.

**Body:** `{ "role": "owner|admin|contributer|viewer" }`  
**204** → (empty)

---

### DELETE /workspaces/{workspaceID}/members/{userID}
Remove a member from the workspace.

**204** → (empty)

---

## Boards

### GET /workspaces/{workspaceID}/boards
List all boards in a workspace the user can access.

**200** → `{ boards: Board[] }`

---

### POST /workspaces/{workspaceID}/boards/create
Create a board in a workspace.

**Body:** `Models.BoardCreation`
```json
{ "title": "string(required)", "prefix": "string(min:3, required)", "description": "string(optional)" }
```
**201** → `Models.Board`  
**409** → `BOARD_PREFIX_TAKEN`

---

### GET /boards/{boardID}
Get board details.

**200** → `Models.Board`  
**404** → `BOARD_NOT_FOUND`

---

### PATCH /boards/{boardID}
Update board details or archive it.

**Body:** `Models.BoardUpdate`
```json
{ "title"?: "string", "description"?: "string", "isArchived"?: "boolean" }
```
**200** → `Models.Board`

---

### DELETE /boards/{boardID}
Delete a board.

**204** → (empty)

---

## Columns

### GET /boards/{boardID}/columns
Get all columns in a board, each with their cards.

**200** → `{ columns: ColumnWithCards[] }`  
**404** → `BOARD_NOT_FOUND`

---

### POST /boards/{boardID}/columns/create
Create a column in a board.

**Body:** `Models.ColumnCreation`
```json
{
  "title": "string(required)",
  "isTerminal": "boolean(required)",
  "position": "integer(required)",
  "limit": "integer(optional)",
  "color": "string(optional, hex /^#[0-9a-f]{6}$/)"
}
```
**201** → `Models.Column`

---

### GET /columns/{columnID}
Get column details (without cards).

**200** → `Models.Column`  
**404** → `COLUMN_NOT_FOUND`

---

### PATCH /columns/{columnID}
Update column details. Only send fields to change.

**Body:** `Models.ColumnUpdate`
```json
{ "title"?: "string", "limit"?: "integer", "color"?: "hex string", "isTerminal"?: "boolean", "position"?: "integer" }
```
**200** → `Models.Column`

---

### DELETE /columns/{columnID}
Delete a column.

**204** → (empty)

---

## Cards

### GET /boards/{boardID}/cards
Get all cards in a board, each with their tags.

**200** → `{ cards: CardWithTags[] }`  
**404** → `BOARD_NOT_FOUND`

---

### POST /boards/{boardID}/cards/create
Create a card in a board.

**Body:** `Models.CardCreation`
```json
{
  "title": "string(required)",
  "position": "number(required)",
  "description": "string(markdown, optional)",
  "columnID": "uuid(optional)",
  "userID": "uuid(optional, assignee)",
  "tagIDs": ["uuid"](optional)
}
```
**201** → `Models.CardWithTags`  
**404** → `BOARD_NOT_FOUND` | `COLUMN_NOT_FOUND`

---

### GET /cards/{cardID}
Get card details including tags.

**200** → `Models.CardWithTags`  
**404** → `CARD_NOT_FOUND`

---

### PATCH /cards/{cardID}
Update card details. Move card between columns by sending a new `columnID`. Reorder by sending a new `position`.

**Body:** `Models.CardUpdate`
```json
{
  "title"?: "string",
  "description"?: "string(markdown)",
  "isArchived"?: "boolean",
  "position"?: "number",
  "columnID"?: "uuid",
  "userID"?: "uuid(assignee)"
}
```
**200** → `Models.CardWithTags`  
**404** → `CARD_NOT_FOUND` | `COLUMN_NOT_FOUND`

---

### DELETE /cards/{cardID}
Delete a card.

**204** → (empty)

---

## Comments

### GET /cards/{cardID}/comments
Get all comments on a card.

**200** → `{ comments: Comment[] }`  
**404** → `CARD_NOT_FOUND`

---

### POST /cards/{cardID}/comments/create
Add a comment to a card.

**Body:** `Models.CommentCreation`
```json
{ "body": "string(markdown, required)" }
```
**201** → `Models.Comment`

---

### GET /comments/{commentID}
Get comment details.

**200** → `Models.Comment`  
**404** → `COMMENT_NOT_FOUND`

---

### PATCH /comments/{commentID}
Update comment body.

**Body:** `{ "body": "string(markdown, required)" }`  
**200** → `Models.Comment`

---

### DELETE /comments/{commentID}
Delete a comment.

**204** → (empty)

---

## Tags

### GET /workspaces/{workspaceID}/tags
List all tags in a workspace.

**200** → `{ tags: Tag[] }`

---

### POST /workspaces/{workspaceID}/tags/create
Create a tag in a workspace.

**Body:** `Models.TagCreation`
```json
{ "name": "string(required)", "color": "string(hex, required)" }
```
**201** → `Models.Tag`  
**409** → `TAG_NAME_TAKEN`

---

### GET /tags/{tagID}
Get tag details.

**200** → `Models.Tag`  
**404** → `TAG_NOT_FOUND`

---

### PATCH /tags/{tagID}
Update tag name or color.

**Body:** `{ "name"?: "string", "color"?: "hex string" }`  
**200** → `Models.Tag`  
**409** → `TAG_NAME_TAKEN`

---

### DELETE /tags/{tagID}
Delete a tag.

**204** → (empty)

---

## Activity

Activity is read-only and auto-generated by the API.

### GET /activity/{activityID}
Get a specific activity entry.

**200** → `Models.Activity`  
**404** → `ACTIVITY_NOT_FOUND`

---

### GET /workspaces/{workspaceID}/activity
Get workspace-level activity with optional filters.

**Query params (all optional):**
| Param | Type | Description |
|-------|------|-------------|
| `boardID` | uuid | Filter to a specific board |
| `cardID` | uuid | Filter to a specific card |
| `userID` | uuid | Filter to a specific user |
| `after` | ISO8601 | Activity after this timestamp |
| `before` | ISO8601 | Activity before this timestamp |

**200** → `Models.Activities`

---

### GET /boards/{boardID}/activity
Get board-level activity with optional filters.

**Query params (all optional):** `cardID`, `userID`, `after`, `before`

**200** → `Models.Activities`  
**404** → `BOARD_NOT_FOUND`

---

### GET /cards/{cardID}/activity
Get card-level activity with optional filters.

**Query params (all optional):** `userID`, `after`, `before`

**200** → `Models.Activities`  
**404** → `CARD_NOT_FOUND`

---

## Data Models

### Models.Workspace
```
id: uuid
title: string
description?: string
ownerID: uuid
createdAt: ISO8601
updatedAt?: ISO8601
```

### Models.Board
```
id: uuid
title: string
description?: string
prefix: string (min 3 chars, unique within workspace)
sequence: integer (auto-incremented card number counter)
isArchived: boolean
workspaceID: uuid
createdAt: ISO8601
updatedAt?: ISO8601
```

### Models.Column
```
id: uuid
title: string
limit?: integer (WIP limit)
color?: string (hex, e.g. "#8DFF6E")
isTerminal: boolean (cards here are quickly archiveable)
position: integer
boardID: uuid
createdAt: ISO8601
updatedAt?: ISO8601
```

### Models.ColumnWithCards
Column + `cards: Card[]`

### Models.Card
```
id: uuid
title: string
description?: string (markdown)
isArchived: boolean
position: number (float, used for ordering — insert between two cards using midpoint)
boardID: uuid
columnID?: uuid
userID?: uuid (assignee)
createdAt: ISO8601
updatedAt?: ISO8601
```

### Models.CardWithTags
Card + `tags: Tag[]`

### Models.Tag
```
id: uuid
name: string (unique within workspace)
color?: string (hex)
workspaceID: uuid
createdAt: ISO8601
updatedAt?: ISO8601
```

### Models.Comment
```
id: uuid
body: string (markdown)
cardID: uuid
userID?: uuid (author)
createdAt: ISO8601
updatedAt?: ISO8601
```

### Models.Activity
```
id: uuid
workspaceID: uuid
description: string
createdAt: ISO8601
userID?: uuid
boardID?: uuid
columnID?: uuid
cardID?: uuid
commentID?: uuid
tagID?: uuid
```

### Models.User
```
id: uuid
email: string
name: string (alphanumeric, min 2)
createdAt: ISO8601
```

### Models.UserToken
```
token: string (JWT)
user: User
```

### Models.UserRole
User + `role: Role`

### Models.Role
Enum: `owner` | `admin` | `contributer` | `viewer`

---

## Error Schema

All errors follow:
```json
{ "code": "string", "reason": "string" }
```

### Error Code Catalog
| Code | HTTP | Meaning |
|------|------|---------|
| `INVALID_AUTH` | 401 | Auth header missing or malformed |
| `NOT_AUTHORIZED` | 401 | User lacks permission for this action |
| `INVALID_CREDENTIALS` | 401 | Wrong email or password on login |
| `INVALID_PASSWORD` | 401 | Wrong current password on update |
| `EMAIL_TAKEN` | 409 | Email already registered |
| `BOARD_NOT_FOUND` | 404 | Board UUID does not exist |
| `BOARD_PREFIX_TAKEN` | 409 | Board prefix already used in workspace |
| `COLUMN_NOT_FOUND` | 404 | Column UUID does not exist |
| `CARD_NOT_FOUND` | 404 | Card UUID does not exist |
| `COMMENT_NOT_FOUND` | 404 | Comment UUID does not exist |
| `TAG_NOT_FOUND` | 404 | Tag UUID does not exist |
| `TAG_NAME_TAKEN` | 409 | Tag name already used in workspace |
| `ACTIVITY_NOT_FOUND` | 404 | Activity UUID does not exist |
| `ALREADY_IN_WORKSPACE` | 409 | User is already a workspace member |

---

## Common Workflows

### New User Onboarding
1. `POST /auth/register` → get JWT
2. `POST /workspaces/create` → get `workspaceID`
3. `POST /workspaces/{workspaceID}/boards/create` → get `boardID`
4. `POST /boards/{boardID}/columns/create` (repeat for each column)
5. `POST /boards/{boardID}/cards/create` → get `cardID`

### Move a Card to a Different Column
`PATCH /cards/{cardID}` with `{ "columnID": "<targetColumnID>", "position": <newPosition> }`

Card `position` is a float. To insert between cards at positions `1.0` and `2.0`, send `1.5`. There is no reorder endpoint — all reordering is done via position on `PATCH`.

### Invite a Team Member
1. Member registers via `POST /auth/register`
2. Workspace owner calls `POST /workspaces/{workspaceID}/members` with `{ "userID": "<memberID>", "role": "contributer" }`

### Attach Tags to a Card
Tags must exist in the workspace first.  
- On creation: include `"tagIDs": ["<uuid>", ...]` in `POST /boards/{boardID}/cards/create`  
- Tags are returned on all `CardWithTags` responses; updating tags on existing cards is done through the card update endpoint (include updated `tagIDs` if the field is available — check current spec).

### Archive vs Delete
- Boards and cards support `isArchived: true` via PATCH — soft delete, data preserved.
- Columns, comments, and tags only support hard DELETE.

---

## Security Schemes

| Scheme | Used By | Format |
|--------|---------|--------|
| `BasicAuth` | `POST /auth/login` only | `Authorization: Basic <base64(email:password)>` |
| `BearerAuth` | All other endpoints | `Authorization: Bearer <JWT>` |
