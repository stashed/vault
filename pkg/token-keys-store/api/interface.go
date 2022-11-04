package api

type TokenKeyInterface interface {
	Get() (string, error)
	TokenName() string
}
