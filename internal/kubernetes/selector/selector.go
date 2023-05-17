package selector

type Selector struct {
	LabelSelector
	FieldSelector
}

func NewSelector(ls []string, fs []string) Selector {
	return Selector{
		LabelSelector: NewLabelSelector(ls),
		FieldSelector: NewFieldSelector(fs),
	}
}
