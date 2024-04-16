package helpers

import "fmt"

type APIError[T any] struct {
	Status int
	Err    error
	Body   T
}

func (ae APIError[T]) StatusCode() int {
	return ae.Status
}
func (ae APIError[T]) JSONObj() any {
	return ae.Body
}

func (ae APIError[T]) Error() string {
	if ae.Err == nil {
		return fmt.Sprintf("API caused an error with status: %d, body: %v", ae.Status, ae.Body)
	}
	return ae.Err.Error()
}
