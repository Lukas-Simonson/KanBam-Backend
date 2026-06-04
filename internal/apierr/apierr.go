package apierr

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type APIError struct {
	Status int    `json:"-"`
	Code   string `json:"code"`
	Reason string `json:"reason"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("Code %s: %s", e.Code, e.Reason)
}

func (e APIError) RespondTo(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	json.NewEncoder(w).Encode(e)
}

func WorkspaceNotFound() APIError {
	return APIError{
		Status: 404,
		Code:   "WORKSPACE_NOT_FOUND",
		Reason: "Workspace not found",
	}
}

func InvalidAuth() APIError {
	return APIError{
		Status: 401,
		Code:   "INVALID_AUTH",
		Reason: "Auth header missing or malformed",
	}
}

func NotAuthorized() APIError {
	return APIError{
		Status: 401,
		Code:   "NOT_AUTHORIZED",
		Reason: "User lacks permission for this action",
	}
}

func InvalidCredentials() APIError {
	return APIError{
		Status: 401,
		Code:   "INVALID_CREDENTIALS",
		Reason: "Wrong email or password",
	}
}

func InvalidPassword() APIError {
	return APIError{
		Status: 401,
		Code:   "INVALID_PASSWORD",
		Reason: "Wrong current password",
	}
}

func EmailTaken() APIError {
	return APIError{
		Status: 409,
		Code:   "EMAIL_TAKEN",
		Reason: "Email already registered",
	}
}

func BoardNotFound() APIError {
	return APIError{
		Status: 404,
		Code:   "BOARD_NOT_FOUND",
		Reason: "Board not found",
	}
}

func BoardPrefixTaken() APIError {
	return APIError{
		Status: 409,
		Code:   "BOARD_PREFIX_TAKEN",
		Reason: "Board prefix already used in this workspace",
	}
}

func ColumnNotFound() APIError {
	return APIError{
		Status: 404,
		Code:   "COLUMN_NOT_FOUND",
		Reason: "Column not found",
	}
}

func CardNotFound() APIError {
	return APIError{
		Status: 404,
		Code:   "CARD_NOT_FOUND",
		Reason: "Card not found",
	}
}

func CommentNotFound() APIError {
	return APIError{
		Status: 404,
		Code:   "COMMENT_NOT_FOUND",
		Reason: "Comment not found",
	}
}

func TagNotFound() APIError {
	return APIError{
		Status: 404,
		Code:   "TAG_NOT_FOUND",
		Reason: "Tag not found",
	}
}

func TagNameTaken() APIError {
	return APIError{
		Status: 409,
		Code:   "TAG_NAME_TAKEN",
		Reason: "Tag name already used in this workspace",
	}
}

func ActivityNotFound() APIError {
	return APIError{
		Status: 404,
		Code:   "ACTIVITY_NOT_FOUND",
		Reason: "Activity not found",
	}
}

func AlreadyInWorkspace() APIError {
	return APIError{
		Status: 409,
		Code:   "ALREADY_IN_WORKSPACE",
		Reason: "User is already a workspace member",
	}
}

func MalformedRequestBody() APIError {
	return APIError{
		Status: 400,
		Code:   "MALFORMED_REQUEST_BODY",
		Reason: "The body of the request was missing or malformed.",
	}
}

func InvalidRequest(reason string) APIError {
	return APIError{
		Status: 400,
		Code:   "INVALID_REQUEST",
		Reason: reason,
	}
}

func UnexpectedServerError() APIError {
	return APIError{
		Status: 500,
		Code:   "UNEXPECTED_SERVER_ERROR",
		Reason: "An unexpected server error has occurreed",
	}
}
