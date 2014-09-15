vcf
===

`vcf` is a `golang` package that parses data from an `io.Reader` adhering to the [Variant Call Format v4.2 Specification](https://samtools.github.io/hts-specs/VCFv4.2.pdf). Data is read asynchronously and returned through two channels, one with correctly parsed variants and one with unknown variants whose parsing failed. Proper initialization and buffering of these channels is a responsibility of the client.

This package is still work in progress, subject to change at any time without notice. Releases will follow [Semantic Versioning 2.0.0](http://semver.org/spec/v2.0.0.html). Major is still in v0 to reflect the early stage development this package is in.

Currently, parsing can handle Samples, optional fields such as ID, Quality and Filter, as well as the INFO field. Info is exposed in two ways:

* As a `map[string]interface{}` exposing all fields found on the INFO for each variant, without any treatment. Key-value pairs are added to this map. In the case of keys such as `DB` which don't have a value, the value used is a `true` boolean.
* As a series of sub-fields listed on section 1.4.1-8 of the VCF 4.2 spec. These sub-fields are provided in a best effort manner. Failure to parse one of these sub-fields will only cause its corresponding pointer to be `nil`, not generating an error. The raw data can always be found on the map.

Genotype fields (section 1.4.2 on the spec) do not have the same kind of treatment yet. They are separated by sample, but the only form represented is a raw map. Easy access to sub-fields is intended in the future.

Structural variants have not been addressed as of version 0.1.0.
