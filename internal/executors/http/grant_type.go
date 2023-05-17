package http

type GrantType uint

const (
	GrantType_CLIENT_CREDENTIALS GrantType = iota
	GrantType_PASSWORD
	GrantType_REFRESH_TOKEN
)

var (
	grantTypeNames = [...]string{"CLIENT_CREDENTIALS", "PASSWORD", "REFRESH_TOKEN"}
)

func (t GrantType) String() string {
	return grantTypeNames[t]
}

func (e GrantType) Ordinal() uint {
	return uint(e)
}
