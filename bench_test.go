package sjson_test

import (
	"encoding/json"
	"github.com/vovkasm/go-sjson"
	"testing"
)

var (
	sample           = `[1434751206.666,"127.0.0.1",[{"site":"Odnolalalalali","uniq2":"db","env":"Canvas","value":"0","project":"SuperGame","type":"VIRAL.requestSuccess.notification_sns","uniq1":"friend_touch"},{"env":"Canvas","value":1,"project":"SuperGame","uniq2":"db","type":"VIRAL.requestFailed.notification_sns","uniq1":"friend_touch","site":"Odnolalalalali"},{"site":"Odnolalalalali","type":"VIRAL.requestInvalid.notification_sns","uniq1":"friend_touch","uniq2":"db","env":"Canvas","value":"0","project":"SuperGame"},{"type":"Viral.requestShortened.notification_sns","uniq1":"friend_touch","env":"Canvas","uniq2":"text","value":100,"project":"SuperGame","site":"Odnolalalalali"}]]`
	sampleStringFast = `"stringwoquoting"`
	sampleStringEsc  = `"fdvsdfvs\"dfvsdfvfds"`
	sampleNumInt     = `1234`
	sampleNum        = `123.45e-2`
	sampleObject     = `{"key1":"val1","key2":123}`
	sampleArray      = `[true,false,true,false,true,false]`
	sampleArrayLong  = `[true,false,true,false,true,false,true,false,true,false,true,false]`
)

var result interface{}

func BenchmarkSample_sjson(b *testing.B)     { benchSimple(b, sample) }
func BenchmarkSample__json(b *testing.B)     { benchStdjsn(b, sample) }
func BenchmarkFastString_sjson(b *testing.B) { benchSimple(b, sampleStringFast) }
func BenchmarkFastString__json(b *testing.B) { benchStdjsn(b, sampleStringFast) }
func BenchmarkEscString_sjson(b *testing.B)  { benchSimple(b, sampleStringEsc) }
func BenchmarkEscString__json(b *testing.B)  { benchStdjsn(b, sampleStringEsc) }
func BenchmarkNumInt_sjson(b *testing.B)     { benchSimple(b, sampleNumInt) }
func BenchmarkNumInt__json(b *testing.B)     { benchStdjsn(b, sampleNumInt) }
func BenchmarkNum_sjson(b *testing.B)        { benchSimple(b, sampleNum) }
func BenchmarkNum__json(b *testing.B)        { benchStdjsn(b, sampleNum) }
func BenchmarkObject_sjson(b *testing.B)     { benchSimple(b, sampleObject) }
func BenchmarkObject__json(b *testing.B)     { benchStdjsn(b, sampleObject) }
func BenchmarkArray_sjson(b *testing.B)      { benchSimple(b, sampleArray) }
func BenchmarkArray__json(b *testing.B)      { benchStdjsn(b, sampleArray) }
func BenchmarkArrayLong_sjson(b *testing.B)  { benchSimple(b, sampleArrayLong) }
func BenchmarkArrayLong__json(b *testing.B)  { benchStdjsn(b, sampleArrayLong) }

func benchSimple(b *testing.B, data string) {
	for n := 0; n < b.N; n++ {
		r, err := sjson.Decode(data)
		if err != nil {
			b.Fatalf("Error in json: %v\n", err)
		}
		result = r
	}
	b.SetBytes(int64(len(data)))
}

func benchStdjsn(b *testing.B, data string) {
	datab := []byte(data)
	for n := 0; n < b.N; n++ {
		var r interface{}
		err := json.Unmarshal(datab, &r)
		if err != nil {
			b.Fatalf("Error in json: %v\n", err)
		}
		result = r
	}
	b.SetBytes(int64(len(data)))
}
