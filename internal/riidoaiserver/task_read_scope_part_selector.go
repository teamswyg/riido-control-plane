package riidoaiserver

import (
	"strings"
)

type taskReadPartSelector struct {
	ParentCommentID *string `json:"parentCommentId,omitempty"`
	Page            *int    `json:"page,omitempty"`
	Limit           *int    `json:"limit,omitempty"`
	BoardCursor     *string `json:"boardCursor,omitempty"`
}

func normalizeReadScopePartSelector(kind AssignmentContentKind, part string, selector *taskReadPartSelector) (taskReadPartSelector, error) {
	invalid := taskContextReadError{"invalid"}
	s := taskReadPartSelector{}
	if selector != nil {
		s = *selector
	}
	if !validAssignmentContentKind(kind) {
		return s, invalid
	}
	switch part {
	case "metadata", "document", "description", "document_alternate":
		if s.ParentCommentID != nil || s.Page != nil || s.Limit != nil || s.BoardCursor != nil {
			return s, invalid
		}
	case "comments_page":
		if kind == AssignmentContentMeeting || s.BoardCursor != nil {
			return s, invalid
		}
		page, limit := 1, 20
		if s.Page != nil {
			page = *s.Page
		}
		if s.Limit != nil {
			limit = *s.Limit
		}
		if page < 1 || page > 2147483647 || limit < 1 || limit > 100 || int64(page-1)*int64(limit) > 2147483647 {
			return s, invalid
		}
		s.Page, s.Limit = &page, &limit
		if s.ParentCommentID != nil {
			parent := strings.TrimSpace(*s.ParentCommentID)
			s.ParentCommentID = nil
			if parent != "" {
				if len(parent) > 256 || strings.ContainsAny(parent, "\x00\r\n") {
					return s, invalid
				}
				s.ParentCommentID = &parent
			}
		}
	case "board_page":
		if kind != AssignmentContentMeeting || s.ParentCommentID != nil || s.Page != nil {
			return s, invalid
		}
		limit := 100
		if s.Limit != nil {
			limit = *s.Limit
		}
		if limit < 1 || limit > 200 {
			return s, invalid
		}
		s.Limit = &limit
		if s.BoardCursor != nil {
			cursor := strings.TrimSpace(*s.BoardCursor)
			s.BoardCursor = nil
			if cursor != "" {
				if len(cursor) > 1024 || strings.ContainsAny(cursor, "\x00\r\n") {
					return s, invalid
				}
				s.BoardCursor = &cursor
			}
		}
	default:
		return s, invalid
	}
	return s, nil
}
