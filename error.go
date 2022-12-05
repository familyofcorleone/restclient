package restclient

import (
	"fmt"
	"strings"
)

type Error struct {
	Code        int
	Description string
	Parameter   string
}

func (e Error) Error() string {
	return e.Description
}

type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

func (r ErrorResponse) Error() string {
	var inners []string
	for _, inner := range r.Errors {
		inners = append(inners, inner.Error())
	}
	return fmt.Sprintf("API errors: %s", strings.Join(inners, ", "))
}
