package json_test

import (
	std "encoding/json"
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
		{"can decode unicode", `"ü"`, Equal("ü"), ``},
		{"can decode escapes", `"\""`, Equal(`"`), ``},
		{"can decode escapes2", `"\u00FC"`, Equal("\u00fc"), ``},
		{"can decode escapes3", `"\u002F\u002f\//"`, Equal("////"), ``},
		{"can decode escapes3", `"\u3042"`, Equal(`あ`), ``},                          // Japanese "a"
		{"can decode escapes from extended range", `"\ud800\udd40"`, Equal("𐅀"), ``}, // Greek Acrophonic Attic One Quarter
		{"errors in strings", `"ab`, BeNil(), `incorrect syntax`},
		// objects
		{"can decode empty object", `{}`, Equal(map[string]interface{}{}), ``},
		{"can decode simple object", `{"key1":"val1"}`, Equal(map[string]interface{}{"key1": "val1"}), ``},
		{"can decode simple object", `{"key1":"val1","key2":"val2"}`, Equal(map[string]interface{}{"key1": "val1", "key2": "val2"}), ``},
		{"can decode simple object", ` { "key1" : 10 , "key2" : true } `, Equal(map[string]interface{}{"key1": 10.0, "key2": true}), ``},
		{"can decode nested objects", `{"k1":{"kk1":10}}`, Equal(map[string]interface{}{"k1": map[string]interface{}{"kk1": 10.0}}), ``},
		{"errors in objects", `{"k1:`, BeNil(), `incorrect syntax`},
		// arrays
		{"can decode empty array", `[]`, Equal([]interface{}{}), ``},
		{"can decode simple array", `[10,20]`, Equal([]interface{}{10.0, 20.0}), ``},
	}
	for n, t := range table {
		n, t := n, t
		Context(fmt.Sprintf("test %d", n), func() {
			It(t.Descr, func() {
				res, err := json.Decode(t.In)
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
	Context("real data test", func() {
		It("should produce equivalent json after reencoding", func() {
			sample := `[1434751206.666,"127.0.0.1",[{"site":"Odnolalalalali","uniq2":"db","env":"Canvas","value":"0","project":"SuperGame","type":"VIRAL.requestSuccess.notification_sns","uniq1":"friend_touch"},{"env":"Canvas","value":1,"project":"SuperGame","uniq2":"db","type":"VIRAL.requestFailed.notification_sns","uniq1":"friend_touch","site":"Odnolalalalali"},{"site":"Odnolalalalali","type":"VIRAL.requestInvalid.notification_sns","uniq1":"friend_touch","uniq2":"db","env":"Canvas","value":"0","project":"SuperGame"},{"type":"Viral.requestShortened.notification_sns","uniq1":"friend_touch","env":"Canvas","uniq2":"text","value":100,"project":"SuperGame","site":"Odnolalalalali"}]]`
			res, err := json.Decode(sample)
			Expect(err).To(Succeed())
			enc, err := std.Marshal(res)
			Expect(err).To(Succeed())
			Expect(enc).To(MatchJSON(sample))
		})
	})
})
