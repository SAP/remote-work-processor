package selector

import (
	"log"

	"k8s.io/apimachinery/pkg/labels"
)

type LabelSelector struct {
	labels.Selector
}

func NewLabelSelector(selectors []string) LabelSelector {
	if len(selectors) == 0 {
		return LabelSelector{
			Selector: labels.Everything(),
		}
	}

	ls := labels.NewSelector()

	for _, s := range selectors {
		r, err := labels.ParseToRequirements(s)
		if err != nil {
			// Ignored
			log.Println(err.Error())
		}

		ls = ls.Add(r[0])
	}

	return LabelSelector{
		Selector: ls,
	}
}
