package model

type Registration struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2"`
	Password string `json:"password" validate:"required,min=8,password_complexity"`
}

type UpdatePassword struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8,password_complexity"`
}

type WorkspaceCreation struct {
	Title       string  `json:"title" validate:"required"`
	Description *string `json:"description"`
}

type WorkspaceUpdate struct {
	Title       *string `json:"title" validate:"omitempty,min=1"`
	Description *string `json:"description"`
}

type BoardCreation struct {
	Title       string  `json:"title" validate:"required"`
	Prefix      string  `json:"prefix" validate:"required,min=3"`
	Description *string `json:"description"`
}

type BoardUpdate struct {
	Title       *string `json:"title" validate:"omitempty,min=1"`
	Description *string `json:"description"`
	IsArchived  *bool   `json:"isArchived"`
}

type ColumnCreation struct {
	Title      string  `json:"title" validate:"required"`
	IsTerminal bool    `json:"isTerminal"`
	Position   int32   `json:"position"`
	WipLimit   *int32  `json:"wipLimit"`
	Color      *string `json:"color" validate:"omitempty,hexcolor"`
}

type ColumnUpdate struct {
	Title      *string `json:"title" validate:"omitempty,min=1"`
	WipLimit   *int32  `json:"wipLimit"`
	Color      *string `json:"color" validate:"omitempty,hexcolor"`
	IsTerminal *bool   `json:"isTerminal"`
	Position   *int32  `json:"position"`
}

type CardCreation struct {
	Title       string   `json:"title" validate:"required"`
	Position    float32  `json:"position"`
	Description *string  `json:"description"`
	ColumnID    *string  `json:"columnID" validate:"omitempty,uuid4"`
	UserID      *string  `json:"userID" validate:"omitempty,uuid4"`
	TagIDs      []string `json:"tagIDs" validate:"omitempty,dive,uuid4"`
}

type CardUpdate struct {
	Title       *string  `json:"title" validate:"omitempty,min=1"`
	Description *string  `json:"description"`
	IsArchived  *bool    `json:"isArchived"`
	Position    *float32 `json:"position"`
	ColumnID    *string  `json:"columnID" validate:"omitempty,uuid4"`
	UserID      *string  `json:"userID" validate:"omitempty,uuid4"`
}

type CommentCreation struct {
	Body string `json:"body" validate:"required"`
}

type CommentUpdate struct {
	Body string `json:"body" validate:"required"`
}

type TagCreation struct {
	Name  string `json:"name" validate:"required"`
	Color string `json:"color" validate:"required,hexcolor"`
}

type TagUpdate struct {
	Name  *string `json:"name" validate:"omitempty,min=1"`
	Color *string `json:"color" validate:"omitempty,hexcolor"`
}

type Membership struct {
	UserID string `json:"userID" validate:"required,uuid4"`
	Role   string `json:"role" validate:"required,oneof=owner admin contributer viewer"`
}

type MembershipUpdate struct {
	Role string `json:"role" validate:"required,oneof=owner admin contributer viewer"`
}
