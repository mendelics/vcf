// Package vcf provides an API for parsing genomic data compliant with the Variant Call Format 4.2 Specification
//
// This API is built with channels, assuming asynchronous computation. Variants parsed successfully are sent
// immediately to the consumer of the API through a channel, as well as variants that fail to be processed.
package vcf
