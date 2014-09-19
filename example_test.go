package vcf

import (
	"fmt"
	"log"
	"os"
)

// Channels should be initialized and passed to the ToChannel function. The client should not close the channels
// This will happen inside ToChannel, when the input is exhausted.
func Example() {
	validVariants := make(chan *Variant, 100)      // buffered channel for correctly parsed variants
	invalidVariants := make(chan InvalidLine, 100) // buffered channel for variants that fail to parse

	filename := "example_vcfs/test.vcf"

	vcfFile, err := os.Open(filename)
	if err != nil {
		log.Fatalln("can't open file", filename)
	}
	defer vcfFile.Close()

	go func() {
		err := ToChannel(vcfFile, validVariants, invalidVariants)
		if err != nil {
			log.Fatalln(err)
		}
	}()

	go func() {
		// consume invalid variants channel asynchronously
		for invalid := range invalidVariants {
			log.Println("failed to parse line", invalid.Line, "with error", invalid.Err)
		}
	}()

	for variant := range validVariants {
		fmt.Println(variant)
		if variant.Qual != nil {
			fmt.Println("Quality:", *variant.Qual)
		}
		fmt.Println("Filter:", variant.Filter)
		fmt.Println("Allele Count:", *variant.AlleleCount)
		fmt.Println("Allele Frequency:", *variant.AlleleFrequency)
		fmt.Println("Total Alleles:", *variant.TotalAlleles)
		fmt.Println("Depth:", *variant.Depth)
		fmt.Println("Mapping Quality:", *variant.MappingQuality)
		fmt.Println("MAPQ0 Reads:", *variant.MAPQ0Reads)

		rawInfo := variant.Info
		vqslod := rawInfo["VQSLOD"]
		fmt.Println("VQSLOD:", vqslod)
	}

	// output:
	// Chromosome: 1 Position: 762588 Reference: G Alternative: C
	// Quality: 40
	// Filter: PASS
	// Allele Count: 2
	// Allele Frequency: 1
	// Total Alleles: 2
	// Depth: 5
	// Mapping Quality: 43.32
	// MAPQ0 Reads: 0
	// VQSLOD: 1.18
}

func ExampleSampleIDs() {
	filename := "example_vcfs/testsamples.vcf"
	vcfFile, err := os.Open(filename)
	if err != nil {
		log.Fatalln("can't open file", filename)
	}
	defer vcfFile.Close()

	sampleIDs, err := SampleIDs(vcfFile)
	if err == nil && sampleIDs != nil {
		for i, sample := range sampleIDs {
			fmt.Printf("sample %d: %s\n", i, sample)
		}
	}
	// output:
	// sample 0: 111222
}
