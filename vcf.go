package vcf

import (
	"bufio"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
)

type Variant struct {
	// required fields
	Chrom string
	Pos   int
	Ref   string
	Alt   string

	// optional
	ID     string
	Qual   *float64
	Filter string
	Info   map[string]interface{}

	// sample data
	Samples []map[string]string
}

type InvalidLine struct {
	Line string
	Err  error
}

// ToChannel opens a file and puts all variants into an already initialized channel
func ToChannel(reader io.Reader, output chan<- *Variant, invalids chan<- InvalidLine) error {
	scanner := bufio.NewScanner(bufio.NewReader(reader))
	header, err := readVcfHeader(scanner)
	if err != nil {
		return err
	} else {
		for scanner.Scan() {
			if !isBlankOrHeaderLine(scanner.Text()) {
				variants, err := parseVcfLine(scanner.Text(), header)
				if variants != nil && err == nil {
					for _, variant := range variants {
						output <- variant
					}
				} else if err != nil {
					invalids <- InvalidLine{scanner.Text(), err}
				}
			}
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
	header, err := readVcfHeader(scanner)
	if err != nil {
		return nil, err
	}
	if len(header) > 9 {
		return header[9:], nil
	}
	return nil, nil
}

func readVcfHeader(scanner *bufio.Scanner) ([]string, error) {
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
	result := make([]*Variant, 0, 64)
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
	fqual, err := strconv.ParseFloat(vcfLine.Qual, 64)
	if err == nil {
		baseVariant.Qual = &fqual
	} else if vcfLine.Qual == "." {
		baseVariant.Qual = nil
	} else {
		baseVariant.Qual = nil
		log.Println("unable to parse quality as float, setting as nil")
	}
	baseVariant.Filter = vcfLine.Filter
	baseVariant.Samples = vcfLine.Samples
	baseVariant.Info = parseInfo(vcfLine.Info)

	alternatives := strings.Split(baseVariant.Alt, ",")

	for _, alternative := range alternatives {

		if baseVariant.Chrom != "" && baseVariant.Pos >= 0 && baseVariant.Ref != "" && alternative != "" {

			result = append(result, &Variant{
				Chrom:   baseVariant.Chrom,
				Pos:     baseVariant.Pos,
				Ref:     baseVariant.Ref,
				Alt:     alternative,
				ID:      baseVariant.ID,
				Samples: baseVariant.Samples,
				Info:    baseVariant.Info,
				Qual:    baseVariant.Qual,
				Filter:  baseVariant.Filter,
			})

		} else {
			return nil, errors.New("error parsing variant: '" + line + "'")
		}
	}
	return result, nil
}

func splitVcfFields(line string) (ret *vcfLine, err error) {

	fields := strings.Split(line, "\t")

	// 7 Fields are mandatory in VCF
	if len(fields) < 8 {
		return nil, errors.New("wrong amount of columns: " + string(len(fields)))
	}
	ret = &vcfLine{}

	// Reading mandatory fields (without type conversions)
	ret.Chr = fields[0]
	ret.Pos = fields[1]
	ret.ID = fields[2]
	ret.Ref = fields[3]
	ret.Alt = fields[4]
	ret.Qual = fields[5]
	ret.Filter = fields[6]
	ret.Info = fields[7]

	// Read sample when have INFO and at least one SAMPLE
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

func parseInfo(info string) map[string]interface{} {
	infoMap := make(map[string]interface{})
	fields := strings.Split(info, ";")
	for _, field := range fields {
		if strings.Contains(field, "=") {
			split := strings.Split(field, "=")
			fieldName, fieldValue := split[0], split[1]
			infoMap[fieldName] = fieldValue
		} else {
			infoMap[field] = true
		}
	}
	return infoMap
}
