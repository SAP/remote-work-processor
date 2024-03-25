package http

type TokenType uint

const (
	TokenType_ACCESS TokenType = iota
	TokenType_ID
)

var (
	tokenTypeNames = [...]string{"ACCESS", "ID"}
)

func (t TokenType) String() string {
	return tokenTypeNames[t]
}
