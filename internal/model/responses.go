package model

type Workspace struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	OwnerID     string  `json:"ownerID"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   *string `json:"updatedAt,omitempty"`
}

type Board struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Prefix      string  `json:"prefix"`
	Sequence    int32   `json:"sequence"`
	IsArchived  bool    `json:"isArchived"`
	WorkspaceID string  `json:"workspaceID"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   *string `json:"updatedAt,omitempty"`
}

type Column struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	WipLimit   *int32  `json:"wipLimit,omitempty"`
	Color      string  `json:"color,omitempty"`
	IsTerminal bool    `json:"isTerminal"`
	Position   int32   `json:"position"`
	BoardID    string  `json:"boardID"`
	CreatedAt  string  `json:"createdAt"`
	UpdatedAt  *string `json:"updatedAt,omitempty"`
}

type ColumnWithCards struct {
	Column
	Cards []Card `json:"cards"`
}

type Card struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	IsArchived  bool    `json:"isArchived"`
	Position    float32 `json:"position"`
	Reference   string  `json:"reference"`
	BoardID     string  `json:"boardID"`
	ColumnID    *string `json:"columnID,omitempty"`
	UserID      *string `json:"userID,omitempty"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   *string `json:"updatedAt,omitempty"`
}

type CardWithTags struct {
	Card
	Tags []Tag `json:"tags"`
}

type Tag struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Color       string  `json:"color"`
	WorkspaceID string  `json:"workspaceID"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   *string `json:"updatedAt,omitempty"`
}

type Comment struct {
	ID        string  `json:"id"`
	Body      string  `json:"body"`
	CardID    string  `json:"cardID"`
	UserID    *string `json:"userID,omitempty"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt *string `json:"updatedAt,omitempty"`
}

type UserRole struct {
	User
	Role string `json:"role"`
}

type Activity struct {
	ID          string  `json:"id"`
	WorkspaceID string  `json:"workspaceID"`
	Description string  `json:"description"`
	CreatedAt   string  `json:"createdAt"`
	UserID      *string `json:"userID,omitempty"`
	BoardID     *string `json:"boardID,omitempty"`
	ColumnID    *string `json:"columnID,omitempty"`
	CardID      *string `json:"cardID,omitempty"`
	CommentID   *string `json:"commentID,omitempty"`
	TagID       *string `json:"tagID,omitempty"`
}

type User struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
}

type UserToken struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
