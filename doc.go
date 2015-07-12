/*
Package sjson provides decoding of JSON Text as defined in ECMA-404.

Sjson designed to be fast and simple, for now it supports only dynamic deserialization.
Simple benchmark shows ~2x speedup against encoding/json standard parser.
	BenchmarkSimple   300000             10216 ns/op
	BenchmarkStd      200000             22413 ns/op

Links

Some useful links.
	* http://json.org - Info about JSON
	* http://www.ecma-international.org/publications/files/ECMA-ST/ECMA-404.pdf - ECMA-404 specification
	* https://metacpan.org/pod/JSON%3A%3AXS - JSON::XS Perl module

Thanks

Development of the project was sponsored by "Crazy Panda" (http://cpdecision.com)
for processing a statistical data.

Some ideas was borrowed from excellent Marc A. Lehmanns JSON::XS code.

*/
package sjson
