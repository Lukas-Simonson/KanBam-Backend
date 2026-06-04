package model

type Registration struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type UpdatePassword struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type WorkspaceCreation struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
}

type WorkspaceUpdate struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type BoardCreation struct {
	Title       string  `json:"title"`
	Prefix      string  `json:"prefix"`
	Description *string `json:"description"`
}

type BoardUpdate struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	IsArchived  *bool   `json:"isArchived"`
}

type ColumnCreation struct {
	Title      string  `json:"title"`
	IsTerminal bool    `json:"isTerminal"`
	Position   int32   `json:"position"`
	WipLimit   *int32  `json:"wipLimit"`
	Color      *string `json:"color"`
}

type ColumnUpdate struct {
	Title      *string `json:"title"`
	WipLimit   *int32  `json:"wipLimit"`
	Color      *string `json:"color"`
	IsTerminal *bool   `json:"isTerminal"`
	Position   *int32  `json:"position"`
}

type CardCreation struct {
	Title       string   `json:"title"`
	Position    float32  `json:"position"`
	Description *string  `json:"description"`
	ColumnID    *string  `json:"columnID"`
	UserID      *string  `json:"userID"`
	TagIDs      []string `json:"tagIDs"`
}

type CardUpdate struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	IsArchived  *bool    `json:"isArchived"`
	Position    *float32 `json:"position"`
	ColumnID    *string  `json:"columnID"`
	UserID      *string  `json:"userID"`
}

type CommentCreation struct {
	Body string `json:"body"`
}

type CommentUpdate struct {
	Body string `json:"body"`
}

type TagCreation struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type TagUpdate struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

type Membership struct {
	UserID string `json:"userID"`
	Role   string `json:"role"`
}

type MembershipUpdate struct {
	Role string `json:"role"`
}
