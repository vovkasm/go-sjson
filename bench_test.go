package json_test

import (
	std "encoding/json"
	"github.com/vovkasm/go-simplejson"
	"testing"
)

var sample = `[1434751206.666,"127.0.0.1",[{"site":"Odnolalalalali","uniq2":"db","env":"Canvas","value":"0","project":"SuperGame","type":"VIRAL.requestSuccess.notification_sns","uniq1":"friend_touch"},{"env":"Canvas","value":1,"project":"SuperGame","uniq2":"db","type":"VIRAL.requestFailed.notification_sns","uniq1":"friend_touch","site":"Odnolalalalali"},{"site":"Odnolalalalali","type":"VIRAL.requestInvalid.notification_sns","uniq1":"friend_touch","uniq2":"db","env":"Canvas","value":"0","project":"SuperGame"},{"type":"Viral.requestShortened.notification_sns","uniq1":"friend_touch","env":"Canvas","uniq2":"text","value":100,"project":"SuperGame","site":"Odnolalalalali"}]]`
var result interface{}

func BenchmarkSimple(b *testing.B) {
	for n := 0; n < b.N; n++ {
		r, err := json.Decode(sample)
		if err != nil {
			b.Fatalf("Error in json: %v\n", err)
		}
		result = r
	}
}

func BenchmarkStd(b *testing.B) {
	json := []byte(sample)
	for n := 0; n < b.N; n++ {
		var r interface{}
		err := std.Unmarshal(json, &r)
		if err != nil {
			b.Fatalf("Error in json: %v\n", err)
		}
		result = r
	}
}
