package json_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/vovkasm/go-simplejson"
	"testing"
)

func TestDict(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "simplejson")
}

var _ = Describe("parser", func() {
	table := []struct {
		Descr  string
		In     string
		Expect types.GomegaMatcher
		Err    string
	}{
		{"decode empty string with error", ``, BeNil(), `incorrect syntax`},
		{"can decode null value", `null`, BeNil(), ``},
		{"decode incorrect value", `nula`, BeNil(), `'null' expected`},
		{"can decode false value", `false`, BeFalse(), ``},
		{"can decode true value", `true`, BeTrue(), ``},
		// numbers
		{"can decode numbers (simple)", `5`, Equal(5.0), ``},
		{"can decode numbers (negative)", `-5`, Equal(-5.0), ``},
		{"can decode numbers (exp)", `5e1`, Equal(50.0), ``},
		{"can decode numbers (-333e+0)", `-333e+0`, Equal(-333.0), ``},
		{"can decode numbers (fractional)", `2.5`, Equal(2.5), ``},
		{"errors in numbers", `+0`, BeNil(), `incorrect syntax`},
		{"errors in numbers", `.2`, BeNil(), `incorrect syntax`},
		{"errors in numbers", `-0.`, BeNil(), `incorrect number`},
		{"errors in numbers", `-0e`, BeNil(), `incorrect number`},
		{"errors in numbers", `-e+1`, BeNil(), `incorrect number`},
		// strings
		{"can decode empty string", `""`, Equal(""), ``},
		{"can decode simple string", `"abc"`, Equal("abc"), ``},
		{"errors in strings", `"ab`, BeNil(), `incorrect syntax`},
		// objects
		{"can decode empty object", `{}`, Equal(map[string]interface{}{}), ``},
	}
	for n, t := range table {
		n, t := n, t
		Context(fmt.Sprintf("test %d", n), func() {
			It(t.Descr, func() {
				res, err := json.Decode([]byte(t.In))
				if len(t.Err) == 0 {
					Expect(err).To(Succeed())
				} else {
					Expect(err).ToNot(Succeed())
					Expect(err.Error()).To(MatchRegexp(t.Err))
				}
				if t.Expect != nil {
					Expect(res).To(t.Expect)
				}
			})
		})
	}
})
