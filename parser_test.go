package sjson_test

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/vovkasm/go-sjson"
)

func TestDict(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "sjson")
}

func ExpectNoErr() func(err error) {
	f := func(err error) {
		ExpectWithOffset(1, err).To(Succeed())
	}
	return f
}

func ExpectErr(msgLike string) func(err error) {
	f := func(err error) {
		ExpectWithOffset(1, err).To(MatchError(MatchRegexp(msgLike)))
	}
	return f
}

func ExpectSyntaxErr(msgLike string, offset int) func(err error) {
	return func(err error) {
		ExpectWithOffset(1, err).To(MatchError(MatchRegexp(msgLike)))
		syntaxErr, ok := err.(*sjson.SyntaxError)
		ExpectWithOffset(1, ok).To(BeTrue(), "error should implement *sjson.SyntaxError")
		ExpectWithOffset(1, syntaxErr.Offset).To(Equal(offset))
	}
}

var _ = Describe("parser", func() {
	table := []struct {
		Descr    string
		In       string
		Expect   types.GomegaMatcher
		ErrCheck func(err error)
	}{
		{"decode empty string with error", ``, BeNil(), ExpectSyntaxErr(`incorrect syntax`, 0)},
		{"can decode null value", `null`, BeNil(), ExpectNoErr()},
		{"decode incorrect value", `nula`, BeNil(), ExpectSyntaxErr(`'null' expected`, 0)},
		{"can decode false value", `false`, BeFalse(), ExpectNoErr()},
		{"can decode true value", `true`, BeTrue(), ExpectNoErr()},
		{"detect incorrect true", `traaa`, BeNil(), ExpectSyntaxErr(`'true' expected`, 0)},
		{"detect incorrect false", `faaaa`, BeNil(), ExpectSyntaxErr(`'false' expected`, 0)},
		// numbers
		{"can decode numbers (simple)", `5`, Equal(5.0), ExpectNoErr()},
		{"can decode numbers (negative)", `-5`, Equal(-5.0), ExpectNoErr()},
		{"can decode numbers (exp)", `5e1`, Equal(50.0), ExpectNoErr()},
		{"can decode numbers (-333e+0)", `-333e+0`, Equal(-333.0), ExpectNoErr()},
		{"can decode numbers (fractional)", `2.5`, Equal(2.5), ExpectNoErr()},
		{"errors in numbers", `+0`, BeNil(), ExpectErr(`incorrect syntax`)},
		{"errors in numbers", `.2`, BeNil(), ExpectErr(`incorrect syntax`)},
		{"errors in numbers", `-0.`, Equal(0.0), ExpectErr(`incorrect number`)},
		{"errors in numbers", `-0e`, Equal(0.0), ExpectSyntaxErr(`incorrect number`, 3)},
		{"errors in numbers", `-e+1`, Equal(0.0), ExpectSyntaxErr(`incorrect number`, 1)},
		{"parse flost error", `11222132131232132132132321.1e100000`, Equal(math.Inf(1)), ExpectErr(`value out of range`)},
		// strings
		{"can decode empty string", `""`, Equal(""), ExpectNoErr()},
		{"can decode simple string", `"abc"`, Equal("abc"), ExpectNoErr()},
		{"can decode unicode", `"ü"`, Equal("ü"), ExpectNoErr()},
		{"can decode escapes", `"\""`, Equal(`"`), ExpectNoErr()},
		{"can decode escapes2", `"\u00FC"`, Equal("\u00fc"), ExpectNoErr()},
		{"can decode escapes3", `"\u002F\u002f\//"`, Equal("////"), ExpectNoErr()},
		{"can decode escapes3", `"\u3042"`, Equal(`あ`), ExpectNoErr()},                          // Japanese "a"
		{"can decode escapes from extended range", `"\ud800\udd40"`, Equal("𐅀"), ExpectNoErr()}, // Greek Acrophonic Attic One Quarter
		{"errors in strings", `"ab`, Equal(""), ExpectSyntaxErr(`incorrect syntax`, 3)},
		{"errors in strings 2", `"ab\"cd`, Equal(""), ExpectSyntaxErr(`incorrect syntax`, 7)},
		{"many escapes", `"bbb\"\\\b\f\n\r\tあeee"`, Equal("bbb\"\\\b\f\n\r\tあeee"), ExpectNoErr()},
		{"many escapes 2", `"bあbb\"\\\b\f\n\r\teee"`, Equal("bあbb\"\\\b\f\n\r\teee"), ExpectNoErr()},
		{"invalid escapes", `"\a"`, Equal(``), ExpectErr("string contains invalid characters")},
		//TODO: do not allow control characters for fast paths :-(
		{"control characters invalid in strings", "\"\r\\\r\"", Equal(``), ExpectErr("string contains invalid characters")},
		// objects
		{"can decode empty object", `{}`, Equal(map[string]interface{}{}), ExpectNoErr()},
		{"can decode simple object", `{"key1":"val1"}`, Equal(map[string]interface{}{"key1": "val1"}), ExpectNoErr()},
		{"can decode simple object", `{"key1":"val1","key2":"val2"}`, Equal(map[string]interface{}{"key1": "val1", "key2": "val2"}), ExpectNoErr()},
		{"can decode simple object", ` { "key1" : 10 , "key2" : true } `, Equal(map[string]interface{}{"key1": 10.0, "key2": true}), ExpectNoErr()},
		{"can decode nested objects", `{"k1":{"kk1":10}}`, Equal(map[string]interface{}{"k1": map[string]interface{}{"kk1": 10.0}}), ExpectNoErr()},
		{"errors in objects", `{`, Equal(map[string]interface{}{}), ExpectSyntaxErr(`incorrect syntax`, 1)},
		{"errors in objects", `{"k1:`, Equal(map[string]interface{}{}), ExpectSyntaxErr(`incorrect syntax`, 5)},
		{"errors in objects", `{"k1"v`, Equal(map[string]interface{}{}), ExpectSyntaxErr(`incorrect syntax - expect ':' after object key`, 5)},
		{"errors in objects", `{"k1":"v1"hmm`, Equal(map[string]interface{}{"k1": "v1"}), ExpectSyntaxErr(`incorrect syntax - expect object key or incomplete object`, 10)},
		// arrays
		{"can decode empty array", `[]`, Equal([]interface{}{}), ExpectNoErr()},
		{"can decode simple array", `[10,20]`, Equal([]interface{}{10.0, 20.0}), ExpectNoErr()},
		{"can decode long array", ` [ 10 , 20 , 30 , 40 , 50 , 60 , 70 , 80 , 90 , 100 ] `, Equal([]interface{}{10.0, 20.0, 30.0, 40.0, 50.0, 60.0, 70.0, 80.0, 90.0, 100.0}), ExpectNoErr()},
		{"errors in array", `[10,20`, Equal([]interface{}{10.0, 20.0}), ExpectSyntaxErr(`incomplete array`, 6)},
		{"errors in array 2", `[10,20[[[`, Equal([]interface{}{10.0, 20.0}), ExpectSyntaxErr(`incomplete array`, 6)},
		{"error in first item", `[nuuu,ll]`, Equal([]interface{}{nil}), ExpectSyntaxErr(`'null' expected`, 1)},
		{"incomplete array after first item", `[true `, Equal([]interface{}{true}), ExpectSyntaxErr(`incomplete array`, 6)},
		{"error in second item", `[true, tru]`, Equal([]interface{}{true}), ExpectSyntaxErr(`'true' expected`, 7)},
		{"long incomplete array", `[1,2,3,4,5,6,7,8,9 `, Equal([]interface{}{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0}), ExpectSyntaxErr(`incomplete array`, 19)},
		{"error in long incomplete array", `[1,2,3,4,5,6,7,8,tru] `, Equal([]interface{}{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0}), ExpectSyntaxErr(`'true' expected`, 17)},
		// spaces
		{"skip various spaces", " \u000a\u000d\u0009true", BeTrue(), ExpectNoErr()},
		{"don't skip low chars as spaces", " \u0008true", BeNil(), ExpectSyntaxErr(`unrecognized token`, 1)},
	}
	for n, t := range table {
		n, t := n, t
		Context(fmt.Sprintf("test %d", n), func() {
			It(t.Descr, func() {
				res, err := sjson.Decode(t.In)
				t.ErrCheck(err)
				if t.Expect != nil {
					Expect(res).To(t.Expect)
				}
			})
		})
	}
	Context("real data test", func() {
		It("should produce equivalent json after reencoding", func() {
			res, err := sjson.Decode(sample)
			Expect(err).To(Succeed())
			enc, err := json.Marshal(res)
			Expect(err).To(Succeed())
			Expect(enc).To(MatchJSON(sample))
		})
		It("should correctly decode encoding/json test", func() {
			if codeJSON == nil {
				codeInit()
			}
			res, err := sjson.Decode(codeJSONStr)
			Expect(err).To(Succeed())
			enc, err := json.Marshal(res)
			Expect(err).To(Succeed())
			Expect(enc).To(MatchJSON(codeJSONStr))
		})
	})
})

func Example() {
	data := `{"name":"John","age":30}`
	obj, err := sjson.Decode(data)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hi, %s!\n", obj.(map[string]interface{})["name"])
	// Output: Hi, John!
}
