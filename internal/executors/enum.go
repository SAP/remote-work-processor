package executors

import "fmt"

type Enumer interface {
	fmt.Stringer
	Ordinal() uint
}
