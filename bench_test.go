package sjson_test

import (
	"encoding/json"
	"github.com/vovkasm/go-sjson"
	"testing"
)

var sample = `[1434751206.666,"127.0.0.1",[{"site":"Odnolalalalali","uniq2":"db","env":"Canvas","value":"0","project":"SuperGame","type":"VIRAL.requestSuccess.notification_sns","uniq1":"friend_touch"},{"env":"Canvas","value":1,"project":"SuperGame","uniq2":"db","type":"VIRAL.requestFailed.notification_sns","uniq1":"friend_touch","site":"Odnolalalalali"},{"site":"Odnolalalalali","type":"VIRAL.requestInvalid.notification_sns","uniq1":"friend_touch","uniq2":"db","env":"Canvas","value":"0","project":"SuperGame"},{"type":"Viral.requestShortened.notification_sns","uniq1":"friend_touch","env":"Canvas","uniq2":"text","value":100,"project":"SuperGame","site":"Odnolalalalali"}]]`
var result interface{}

func BenchmarkSimple(b *testing.B) {
	data := sample
	for n := 0; n < b.N; n++ {
		r, err := sjson.Decode(data)
		if err != nil {
			b.Fatalf("Error in json: %v\n", err)
		}
		result = r
	}
}

func BenchmarkStd(b *testing.B) {
	data := []byte(sample)
	for n := 0; n < b.N; n++ {
		var r interface{}
		err := json.Unmarshal(data, &r)
		if err != nil {
			b.Fatalf("Error in json: %v\n", err)
		}
		result = r
	}
}
