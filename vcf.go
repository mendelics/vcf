package vcf

import (
	"bufio"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
)

// Variant is a struct representing the fields specified in the VCF 4.2 spec. It does not support structural variants. When the variant is generated through the API of the vcf package, the required fields are guaranteed to be valid, otherwise the parsing for the variant fails and is reported.
// Multiple alternatives are parsed as separated instances of the type Variant
// All other fields are optional and will not cause parsing fails if missing or non-conformant.
type Variant struct {
	// Required fields
	Chrom string
	Pos   int
	Ref   string
	Alt   string

	ID string
	// Qual is a pointer so that it can be set to nil when it is a dot '.'
	Qual   *float64
	Filter string
	// Info is a map containing all the keys present in the INFO field, with their corresponding value. For keys without corresponding values, the value is a `true` bool.
	// No attempt at parsing is made on this field, data is raw. The only exception is for multiple alternatives data. These are reported separately for each variant
	Info map[string]interface{}

	// Genotype fields for each sample
	Samples []map[string]string

	// Optional info fields. These are the reserved fields listed on the VCF 4.2 spec, session 1.4.1, number 8. The parsing is lenient, if the fields do not conform to the expected type listed here, they will be set to nil
	// The fields are meant as helpers for common scenarios, since the generic usage is covered by the Info map
	// Definitions used in the metadata section of the header are not used
	AncestralAllele *string
	Depth           *int
	AlleleFrequency *float64
	AlleleCount     *int
	TotalAlleles    *int
	End             *int
	MAPQ0Reads      *int
	NumberOfSamples *int
	MappingQuality  *float64
	Cigar           *string
	InDBSNP         *bool
	InHapmap2       *bool
	InHapmap3       *bool
	IsSomatic       *bool
	IsValidated     *bool
	In1000G         *bool
	BaseQuality     *float64
	StrandBias      *float64
}

// InvalidLine represents a VCF line that could not be parsed. It encapsulates the problematic line with its corresponding error.
type InvalidLine struct {
	Line string
	Err  error
}

// ToChannel opens a file and puts all variants into an already initialized channel. Variants whose parsing fails go into a specific channel for failing variants
// Both channels are closed when the reader is fully scanned
func ToChannel(reader io.Reader, output chan<- *Variant, invalids chan<- InvalidLine) error {
	scanner := bufio.NewScanner(bufio.NewReader(reader))
	header, err := vcfHeader(scanner)
	if err != nil {
		return err
	}

	for scanner.Scan() {
		if isBlankOrHeaderLine(scanner.Text()) {
			continue
		}
		variants, err := parseVcfLine(scanner.Text(), header)
		if variants != nil && err == nil {
			for _, variant := range variants {
				output <- variant
			}
		} else if err != nil {
			invalids <- InvalidLine{scanner.Text(), err}
		}
	}

	close(output)
	close(invalids)

	return nil
}

// SampleIDs reads a vcf header from an io.Reader and returns a slice with all the sample IDs contained in that header
// If there are no samples on the header, a nil slice is returned
func SampleIDs(reader io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(bufio.NewReader(reader))
	header, err := vcfHeader(scanner)
	if err != nil {
		return nil, err
	}
	if len(header) > 9 {
		return header[9:], nil
	}
	return nil, nil
}

func vcfHeader(scanner *bufio.Scanner) ([]string, error) {
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "#") && !strings.HasPrefix(scanner.Text(), "##") {
			return strings.Split(scanner.Text()[1:], "\t"), nil
		}
	}
	return nil, errors.New("vcf header not found on file")
}

func isBlankOrHeaderLine(line string) bool {
	return strings.HasPrefix(line, "#") || line == ""
}

type vcfLine struct {
	Chr, Pos, ID, Ref, Alt, Qual, Filter, Info string
	Format                                     []string
	Samples                                    []map[string]string
}

func parseVcfLine(line string, header []string) ([]*Variant, error) {
	vcfLine, err := splitVcfFields(line)
	if err != nil {
		return nil, errors.New("unable to parse apparently misformatted VCF line: " + line)
	}

	baseVariant := Variant{}
	baseVariant.Chrom = vcfLine.Chr
	pos, _ := strconv.Atoi(vcfLine.Pos)
	baseVariant.Pos = pos - 1 // converts variant to 0-based
	baseVariant.Ref = strings.ToUpper(vcfLine.Ref)
	baseVariant.Alt = strings.ToUpper(strings.Replace(vcfLine.Alt, ".", "", -1))

	baseVariant.ID = vcfLine.ID
	floatQuality, err := strconv.ParseFloat(vcfLine.Qual, 64)
	if err == nil {
		baseVariant.Qual = &floatQuality
	} else if vcfLine.Qual == "." {
		baseVariant.Qual = nil
	} else {
		baseVariant.Qual = nil
		log.Println("unable to parse quality as float, setting as nil")
	}
	baseVariant.Filter = vcfLine.Filter
	baseVariant.Samples = vcfLine.Samples
	baseVariant.Info = infoToMap(vcfLine.Info)

	alternatives := strings.Split(baseVariant.Alt, ",")

	info := splitMultipleAltInfos(baseVariant.Info, len(alternatives))

	result := make([]*Variant, 0, 64)
	for i, alternative := range alternatives {

		if baseVariant.Chrom != "" && baseVariant.Pos >= 0 && baseVariant.Ref != "" && alternative != "" {

			var altinfo map[string]interface{}
			if i >= len(info) {
				altinfo = info[0]
			} else {
				altinfo = info[i]
			}

			variant := &Variant{
				Chrom:   baseVariant.Chrom,
				Pos:     baseVariant.Pos,
				Ref:     baseVariant.Ref,
				Alt:     alternative,
				ID:      baseVariant.ID,
				Samples: baseVariant.Samples,
				Info:    altinfo,
				Qual:    baseVariant.Qual,
				Filter:  baseVariant.Filter,
			}
			buildInfoSubFields(variant)

			result = append(result, variant)

		} else {
			return nil, errors.New("error parsing variant: '" + line + "'")
		}
	}
	return result, nil
}

func splitVcfFields(line string) (ret *vcfLine, err error) {

	fields := strings.Split(line, "\t")

	if len(fields) < 8 {
		return nil, errors.New("wrong amount of columns: " + string(len(fields)))
	}
	ret = &vcfLine{}

	ret.Chr = fields[0]
	ret.Pos = fields[1]
	ret.ID = fields[2]
	ret.Ref = fields[3]
	ret.Alt = fields[4]
	ret.Qual = fields[5]
	ret.Filter = fields[6]
	ret.Info = fields[7]

	if len(fields) > 8 {
		samples := fields[9:len(fields)]
		ret.Samples = make([]map[string]string, len(fields)-9)
		ret.Format = strings.Split(fields[8], ":")
		for i, sample := range samples {
			ret.Samples[i] = parseSample(ret.Format, sample)
		}
	}

	return
}

func parseSample(format []string, unparsedSample string) map[string]string {
	sampleMapping := make(map[string]string)
	sampleFields := strings.Split(unparsedSample, ":")
	for i, field := range sampleFields {
		sampleMapping[format[i]] = field
	}
	return sampleMapping
}
