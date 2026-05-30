package dto

type ListResponse[T any] struct {
	List  []T `json:"list"`
	Total int `json:"total"`
}
