
CREATE TABLE IF NOT EXISTS kb_user(
    id            UUID PRIMARY KEY,
    name          TEXT NOT NULL,
    email         CITEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS workspace(
    id          UUID PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT,
    owner_id    UUID NOT NULL REFERENCES kb_user(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ
);

CREATE TYPE role AS ENUM ('owner', 'admin', 'contributer', 'viewer');

CREATE TABLE IF NOT EXISTS workspace_user(
    user_id      UUID NOT NULL REFERENCES kb_user(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    pending      BOOLEAN NOT NULL DEFAULT true,
    role         role NOT NULL
);

CREATE TABLE IF NOT EXISTS board(
    id           UUID PRIMARY KEY,
    title        TEXT NOT NULL,
    description  TEXT,
    prefix       CITEXT NOT NULL,
    sequence     INT NOT NULL DEFAULT 0,
    is_archived  BOOLEAN NOT NULL DEFAULT false,
    workspace_id UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ,

    UNIQUE (prefix, workspace_id)
);

-- Custom Color Type
CREATE DOMAIN hex_color AS VARCHAR(7)
CHECK (VALUE ~* '^#[0-9a-f]{6}$');

CREATE TABLE IF NOT EXISTS kb_column(
    id UUID     PRIMARY KEY,
    title       TEXT NOT NULL,
    wip_limit   INTEGER,
    color       hex_color,
    is_terminal BOOLEAN NOT NULL DEFAULT false,

    position INTEGER NOT NULL,
    board_id UUID NOT NULL REFERENCES board(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS card(
    id          UUID PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT,
    is_archived BOOLEAN NOT NULL DEFAULT false,
    position    REAL NOT NULL,
    reference   TEXT NOT NULL, -- Example: #CHOR-42
    
    workspace_id UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    board_id     UUID NOT NULL REFERENCES board(id) ON DELETE CASCADE,
    column_id    UUID REFERENCES kb_column(id) ON DELETE SET NULL,
    user_id      UUID REFERENCES kb_user(id) ON DELETE SET NULL,

    UNIQUE(reference, workspace_id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS comment(
    id UUID PRIMARY KEY,
    body TEXT NOT NULL,

    card_id UUID NOT NULL REFERENCES card(id) ON DELETE CASCADE,
    user_id UUID REFERENCES kb_user(id) ON DELETE SET NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS tag(
    id    UUID PRIMARY KEY,
    name  CITEXT NOT NULL,
    color hex_color NOT NULL,

    workspace_id UUID REFERENCES workspace(id) ON DELETE CASCADE,

    UNIQUE(name, workspace_id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS card_tag(
    card_id UUID NOT NULL REFERENCES card(id) ON DELETE CASCADE,
    tag_id  UUID NOT NULL REFERENCES tag(id) ON DELETE CASCADE,

    PRIMARY KEY(card_id, tag_id)
);

CREATE TABLE IF NOT EXISTS activity(
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    user_id UUID REFERENCES kb_user(id) ON DELETE SET NULL,
    board_id UUID REFERENCES board(id) ON DELETE SET NULL,
    column_id UUID REFERENCES kb_column(id) ON DELETE SET NULL,
    card_id UUID REFERENCES card(id) ON DELETE SET NULL,
    comment_id UUID REFERENCES comment(id) ON DELETE SET NULL,
    tag_id UUID REFERENCES tag(id) ON DELETE SET NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);