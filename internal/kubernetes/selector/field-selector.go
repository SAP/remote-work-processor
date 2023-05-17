package selector

import (
	"fmt"
	"log"

	"github.com/itchyny/gojq"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FieldSelector struct {
	jqs []*gojq.Code
}

func NewFieldSelector(selectors []string) FieldSelector {
	if len(selectors) == 0 {
		return FieldSelector{
			jqs: []*gojq.Code{},
		}
	}

	jqs := []*gojq.Code{}

	for _, s := range selectors {
		q, err := gojq.Parse(s)
		if err != nil {
			// Ignored
			fmt.Println(err.Error())
			continue
		}

		c, err := gojq.Compile(q)
		if err != nil {
			// Ignored
			fmt.Println(err.Error())
			continue
		}

		jqs = append(jqs, c)
	}

	return FieldSelector{
		jqs: jqs,
	}
}

func (fs *FieldSelector) Matches(o client.Object) bool {
	fields, err := runtime.DefaultUnstructuredConverter.ToUnstructured(o)
	if err != nil {
		log.Printf("Failed to convert object to a unstructured one: %v\n", err)
		return false
	}

	for _, jq := range fs.jqs {
		r, ok := jq.Run(fields).Next()
		if isFalsy(r) || !ok {
			return false
		}
	}

	return true
}

func isFalsy(r any) bool {
	if r == nil {
		return true
	}

	v, ok := r.(bool)
	if ok {
		return !v
	}

	return false
}