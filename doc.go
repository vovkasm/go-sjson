/*
Package sjson provides fast and simple JSON decoder.

Sjson is designed to be fast and simple, for now it supports only dynamic deserialization.
It can decode JSON Text as defined in ECMA-404.
Simple benchmark test shows ~2x speedup against encoding/json standard parser.
	$ go test -bench=Sample\|Code -benchtime=5s
	BenchmarkSample_sjson    1000000              7582 ns/op          87.43 MB/s // Equivalent of our production JSON
	BenchmarkSample__json     300000             19579 ns/op          33.86 MB/s
	BenchmarkCode_sjson          300          28384877 ns/op          68.36 MB/s // JSON Text from the encoding/json package
	BenchmarkCode__json          100          60002297 ns/op          32.34 MB/s

Links

Some useful links.
	* http://json.org - Info about JSON
	* http://www.ecma-international.org/publications/files/ECMA-ST/ECMA-404.pdf - ECMA-404 specification
	* https://metacpan.org/pod/JSON%3A%3AXS - JSON::XS Perl module

Thanks

Development of the project was sponsored by CP Decision Limited (http://cpdecision.com)
as part of a project on processing statistical data.

Some ideas was borrowed from excellent Marc A. Lehmann's JSON::XS code.

*/
package sjson
