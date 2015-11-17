package vcf

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

// Variant is a struct representing the fields specified in the VCF 4.2 spec.
//
// When the variant is generated through the API of the vcf package, the required fields are guaranteed to be valid,
// otherwise the parsing for the variant fails and is reported.
//
// Multiple alternatives are parsed as separated instances of the type Variant. All other fields are optional and will
// not cause parsing fails if missing or non-conformant.
type Variant struct {
	// Required fields
	Chrom string
	Pos   int
	Ref   string
	Alt   string

	ID string

	// Qual is a pointer so that it can be set to nil when it is a dot '.'
	Qual *float64

	Filter string

	// Info is a map containing all the keys present in the INFO field, with their corresponding value.
	// For keys without corresponding values, the value is a `true` bool.
	// No attempt at parsing is made on this field, data is raw.
	// The only exception is for multiple alternatives data. These are reported separately for each variant.
	Info map[string]interface{}

	// Genotype fields for each sample
	Samples []map[string]string

	// Optional info fields. These are the reserved fields listed on the VCF 4.2 spec, session 1.4.1, number 8.
	// The parsing is lenient, if the fields do not conform to the expected type listed here, they will be set to nil.
	// The fields are meant as helpers for common scenarios, since the generic usage is covered by the Info map.
	// Definitions used in the metadata section of the header are not used.
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

	// Structural variants
	SVType   *string
	SVLength *int
}

// String provides a representation of the variant key: the fields Chrom, Pos, Ref and Alt
// compatible with fmt.Stringer
func (v *Variant) String() string {
	return fmt.Sprintf("Chromosome: %s Position: %d Reference: %s Alternative: %s", v.Chrom, v.Pos, v.Ref, v.Alt)
}

// InvalidLine represents a VCF line that could not be parsed.
// It encapsulates the problematic line with its corresponding error.
type InvalidLine struct {
	Line string
	Err  error
}

// ToChannel reads from an io.Reader and puts all variants into an already initialized channel.
// Variants whose parsing fails go into a specific channel for failing variants.
// If any of the two channels are full, ToChannel will block.
// The consumer must guarantee there is enough buffer space on the channels.
// Both channels are closed when the reader is fully scanned.
func ToChannel(reader io.Reader, output chan<- *Variant, invalids chan<- InvalidLine) error {
	bufferedReader := bufio.NewReaderSize(reader, 100*1024)
	header, err := vcfHeader(bufferedReader)
	if err != nil {
		return err
	}

	for {
		line, readError := bufferedReader.ReadString('\n')
		if readError != nil && readError != io.EOF {
			// If an error that is not an EOF happens break immediately
			err = readError
			break
		}
		if line == "" && readError == io.EOF {
			// If there is an empty line at EOF, end the loop without propagating the error
			break
		}
		if isHeaderLine(line) {
			continue
		}
		variants, err := parseVcfLine(line, header)
		if variants != nil && err == nil {
			for _, variant := range variants {
				fixedVariant := fixRefAltSuffix(variant)
				output <- fixedVariant
			}
		} else if err != nil {
			invalids <- InvalidLine{line, err}
		}
		// Check again for a read error. This is only possible on EOF
		if readError != nil {
			break
		}
	}

	close(output)
	close(invalids)

	return err
}

// SampleIDs reads a vcf header from an io.Reader and returns a slice with all the sample IDs contained in that header.
// If there are no samples on the header, a nil slice is returned
func SampleIDs(reader io.Reader) ([]string, error) {
	bufferedReader := bufio.NewReaderSize(reader, 100*1024)
	header, err := vcfHeader(bufferedReader)
	if err != nil {
		return nil, err
	}
	if len(header) > 9 {
		return header[9:], nil
	}
	return nil, nil
}

func vcfHeader(bufferedReader *bufio.Reader) ([]string, error) {
	for {
		line, err := bufferedReader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "##") {
			line = strings.TrimSpace(line)
			return strings.Split(line[1:], "\t"), nil
		}
	}
	return nil, errors.New("vcf header not found on file")
}

func isHeaderLine(line string) bool {
	return strings.HasPrefix(line, "#")
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

	if strings.Contains(vcfLine.Chr, "chr") {
		baseVariant.Chrom = strings.Replace(vcfLine.Chr, "chr", "", -1)
	} else {
		baseVariant.Chrom = vcfLine.Chr
	}
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

func fixRefAltSuffix(variant *Variant) *Variant {
	ref := variant.Ref
	alt := variant.Alt
	i := len(ref) - 1
	j := len(alt) - 1
	for ref[i] == alt[j] && i > 0 && j > 0 {
		i--
		j--
	}
	newRef := ref[:i+1]
	newAlt := alt[:j+1]
	variant.Ref = newRef
	variant.Alt = newAlt
	return variant
}
