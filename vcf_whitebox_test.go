package vcf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ParseVcfLineSuite struct {
	suite.Suite
}

var defaultHeader = []string{"CHROM", "POS", "ID", "REF", "ALT", "QUAL", "FILTER", "INFO"}

func (s *ParseVcfLineSuite) TestBlankLineShouldReturnError() {
	result, err := parseVcfLine("\t ", defaultHeader)
	assert.Error(s.T(), err, "Line with only blanks should return empty and an error")
	assert.Empty(s.T(), result, "Line with only blanks should return emptyand an error")
}

func (s *ParseVcfLineSuite) TestContinuousLineShouldReturnError() {
	result, err := parseVcfLine("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nunc tellus ligula, faucibus sed nibh sed, fringilla viverra enim.", defaultHeader)
	assert.Error(s.T(), err, "Line with continuous text should return empty and an error")
	assert.Empty(s.T(), result, "Line with continuous text should return empty and an error")
}

func (s *ParseVcfLineSuite) TestEmptyFieldedLineShouldReturnError() {
	result, err := parseVcfLine("\t\t\t\t\t", defaultHeader)
	assert.Error(s.T(), err, "Line with empty fields should return empty and an error")
	assert.Empty(s.T(), result, "Line with empty fields should return empty and an error")
}

func (s *ParseVcfLineSuite) TestWrongFormattedFieldedLineShouldReturnError() {
	result, err := parseVcfLine("A\tB\tC\tD\tE\tF", defaultHeader)
	assert.Error(s.T(), err, "Line with wrong formatted fields should return empty and an error")
	assert.Empty(s.T(), result, "Line with wrong formatted fields should return empty and an error")
}

func (s *ParseVcfLineSuite) TestValidLineShouldReturnOneElementAndNoErrors() {
	result, err := parseVcfLine("1\t847491\trs28407778\tGT\tA\t745.77\tPASS\tAC=1;AF=0.500;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant\tGT:AD:DP:GQ:PL\t0/1:16,25:41:99:774,0,434", defaultHeader)

	assert.NoError(s.T(), err, "Valid VCF line should not return error")
	assert.NotNil(s.T(), result, "Valid VCF line should not return nil")
	assert.Exactly(s.T(), len(result), 1, "Valid VCF should return a list with one element")
	assert.Equal(s.T(), result[0].Chrom, "1", "result.Chrom should be 1")
	assert.Equal(s.T(), result[0].Pos, 847490, "result.Pos should be 0-based to 847490")
	assert.Equal(s.T(), result[0].Ref, "GT", "result.Ref should be GT")
	assert.Equal(s.T(), result[0].Alt, "A", "result.Alt should be A")
}

func (s *ParseVcfLineSuite) TestValidLineWithChrShouldStripIt() {
	result, err := parseVcfLine("chr1\t847491\trs28407778\tGT\tA\t745.77\tPASS\tAC=1;AF=0.500;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant\tGT:AD:DP:GQ:PL\t0/1:16,25:41:99:774,0,434", defaultHeader)

	assert.NoError(s.T(), err, "Valid VCF line should not return error")
	assert.NotNil(s.T(), result, "Valid VCF line should not return nil")
	assert.Exactly(s.T(), len(result), 1, "Valid VCF should return a list with one element")
	assert.Equal(s.T(), result[0].Chrom, "1", "result.Chrom should be 1")
	assert.Equal(s.T(), result[0].Pos, 847490, "result.Pos should be 0-based to 847490")
	assert.Equal(s.T(), result[0].Ref, "GT", "result.Ref should be GT")
	assert.Equal(s.T(), result[0].Alt, "A", "result.Alt should be A")
}

func (s *ParseVcfLineSuite) TestValidLineWithLowercaseRefAndAltShouldReturnOneElementAndNoErrors() {
	result, err := parseVcfLine("1\t847491\trs28407778\tgt\ta\t745.77\tPASS\tAC=1;AF=0.500;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant\tGT:AD:DP:GQ:PL\t0/1:16,25:41:99:774,0,434", defaultHeader)

	assert.NoError(s.T(), err, "Valid VCF line should not return error")
	assert.NotNil(s.T(), result, "Valid VCF line should not return nil")
	assert.Exactly(s.T(), len(result), 1, "Valid VCF should return a list with one element")
	assert.Equal(s.T(), result[0].Chrom, "1", "result.Chrom should be 1")
	assert.Equal(s.T(), result[0].Pos, 847490, "result.Pos should be 0-based to 847490")
	assert.Equal(s.T(), result[0].Ref, "GT", "result.Ref should be GT")
	assert.Equal(s.T(), result[0].Alt, "A", "result.Alt should be A")
}

func (s *ParseVcfLineSuite) TestDotsShouldBeRemovedFromValidLineAlternative() {
	result, err := parseVcfLine("1\t847491\trs28407778\tGTTTA\tG....\t745.77\tPASS\tAC=1;AF=0.500;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant\tGT:AD:DP:GQ:PL\t0/1:16,25:41:99:774,0,434", defaultHeader)

	assert.NoError(s.T(), err, "Valid VCF line should not return error")
	assert.NotNil(s.T(), result, "Valid VCF line should not return nil")
	assert.Exactly(s.T(), len(result), 1, "Valid VCF should return a list with one element")
	assert.Equal(s.T(), result[0].Chrom, "1", "result.Chrom should be 1")
	assert.Equal(s.T(), result[0].Pos, 847490, "result.Pos should be 0-based to 847490")
	assert.Equal(s.T(), result[0].Ref, "GTTTA", "result.Ref should be GTTTA")
	assert.Equal(s.T(), result[0].Alt, "G", "result.Alt should be G")
}

func (s *ParseVcfLineSuite) TestValidLineWithMultipleAlternativesShouldReturnThreeElementsAndNoErrors() {
	result, err := parseVcfLine("1\t847491\trs28407778\tGT\tA,C,G\t745.77\tPASS\tAC=1;AF=0.300,0.300,0.400;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant\tGT:AD:DP:GQ:PL\t0/1:16,25:41:99:774,0,434", defaultHeader)

	assert.NoError(s.T(), err, "Valid VCF line should not return error")
	assert.NotNil(s.T(), result, "Valid VCF line should not return nil")
	assert.Exactly(s.T(), len(result), 3, "Valid VCF should return a list with one element")

	assert.Equal(s.T(), result[0].Chrom, "1", "result[0].Chrom should be 1")
	assert.Equal(s.T(), result[0].Pos, 847490, "result[0].Pos should be 0-based to 847490")
	assert.Equal(s.T(), result[0].Ref, "GT", "result[0].Ref should be GT")
	assert.Equal(s.T(), result[0].Alt, "A", "result[0].Alt should be A")

	assert.Equal(s.T(), result[1].Chrom, "1", "result[1].Chrom should be 1")
	assert.Equal(s.T(), result[1].Pos, 847490, "result[1].Pos should be 0-based to 847490")
	assert.Equal(s.T(), result[1].Ref, "GT", "result[1].Ref should be GT")
	assert.Equal(s.T(), result[1].Alt, "C", "result[1].Alt should be A")

	assert.Equal(s.T(), result[2].Chrom, "1", "result[2].Chrom should be 1")
	assert.Equal(s.T(), result[2].Pos, 847490, "result[2].Pos should be 0-based to 847490")
	assert.Equal(s.T(), result[2].Ref, "GT", "result[2].Ref should be GT")
	assert.Equal(s.T(), result[2].Alt, "G", "result[2].Alt should be A")
}

func (s *ParseVcfLineSuite) TestValidLineWithSampleGenotypeFields() {
	result, err := parseVcfLine("1\t847491\trs28407778\tGTTTA\tG....\t745.77\tPASS\tAC=1;AF=0.500;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant\tGT:AD:DP:GQ:PL\t0/1:16,25:41:99:774,0,434", defaultHeader)

	assert.NoError(s.T(), err, "Valid VCF line should not return error")
	assert.NotNil(s.T(), result, "Valid VCF line should not return nil")
	assert.Exactly(s.T(), len(result), 1, "Valid VCF should return a list with one element")

	samples := result[0].Samples
	assert.NotNil(s.T(), samples, "Valid VCF should contain slice of sample maps")
	assert.Exactly(s.T(), len(samples), 1, "Valid VCF should contain one sample")
	sampleMap := samples[0]
	assert.NotNil(s.T(), sampleMap, "Genotype field mapping should not return nil")
	assert.Exactly(s.T(), len(sampleMap), 5, "Sample map should have as many keys as there are formats")

	gt, ok := sampleMap["GT"]
	assert.True(s.T(), ok, "GT key must be found")
	assert.Equal(s.T(), gt, "0/1", "gt")

	ad, ok := sampleMap["AD"]
	assert.True(s.T(), ok, "AD key must be found")
	assert.Equal(s.T(), ad, "16,25", "ad")

	dp, ok := sampleMap["DP"]
	assert.True(s.T(), ok, "AD key must be found")
	assert.Equal(s.T(), dp, "41", "dp")

	gq, ok := sampleMap["GQ"]
	assert.True(s.T(), ok, "GQ key must be found")
	assert.Equal(s.T(), gq, "99", "gq")

	pl, ok := sampleMap["PL"]
	assert.True(s.T(), ok, "PL key must be found")
	assert.Equal(s.T(), pl, "774,0,434", "pl")
}

func (s *ParseVcfLineSuite) TestInfoFields() {
	result, err := parseVcfLine("1\t847491\trs28407778\tG\tA,C\t745.77\tPASS\tAC=1,2;AF=0.500,0.335;AN=2;BQ=30.00;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;NS=27;H2;H3;SOMATIC;VALIDATED;1000G;MLEAC=1;MLEAF=0.500;END=847492;MQ=60.00;MQ0=0;SB=0.127;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;CIGAR=a;culprit=FS;set=variant\tGT:AD:DP:GQ:PL\t0/1:16,25:41:99:774,0,434", defaultHeader)

	assert.NoError(s.T(), err, "Valid VCF line should not return error")
	assert.NotNil(s.T(), result, "Valid VCF line should not return nil")
	assert.Exactly(s.T(), len(result), 2, "Valid VCF should return a list with two elements")

	info := result[0].Info
	assert.NotNil(s.T(), info, "Valid VCF should contain info map")
	assert.Exactly(s.T(), len(info), 28, "Info should contain 20 keys")

	ac, ok := info["AC"]
	assert.True(s.T(), ok, "AC key must be found")
	assert.Equal(s.T(), ac, "1", "ac")

	af, ok := info["AF"]
	assert.True(s.T(), ok, "AF key must be found")
	assert.Equal(s.T(), af, "0.500", "af")

	af, ok = result[1].Info["AF"]
	assert.True(s.T(), ok, "AF key must be found")
	assert.Equal(s.T(), af, "0.335", "af")

	db, ok := info["DB"]
	assert.True(s.T(), ok, "DB key must be found")
	booldb, isbool := db.(bool)
	assert.True(s.T(), isbool, "DB value must be a boolean")
	assert.True(s.T(), booldb)

	_, ok = info["AA"]
	assert.False(s.T(), ok, "AA key must not be found")

	aa := result[0].AncestralAllele
	assert.Nil(s.T(), aa, "No AA field")

	dp := result[0].Depth
	assert.NotNil(s.T(), dp, "Depth field of first element must be found")
	assert.Equal(s.T(), *dp, 41)

	dp = result[1].Depth
	assert.NotNil(s.T(), dp, "Depth field of second element must be found")
	assert.Equal(s.T(), *dp, 41)

	freq := result[0].AlleleFrequency
	assert.NotNil(s.T(), freq, "AlleleFrequency field must be found")
	assert.Equal(s.T(), *freq, 0.500)
	freq = result[1].AlleleFrequency
	assert.NotNil(s.T(), freq, "AlleleFrequency field must be found")
	assert.Equal(s.T(), *freq, 0.335)

	count := result[0].AlleleCount
	assert.NotNil(s.T(), count, "AlleleCount field must be found")
	assert.Equal(s.T(), *count, 1)
	count = result[1].AlleleCount
	assert.NotNil(s.T(), count, "AlleleCount field must be found")
	assert.Equal(s.T(), *count, 2)

	total := result[0].TotalAlleles
	assert.NotNil(s.T(), total, "TotalAlleles field must be found")
	assert.Equal(s.T(), *total, 2)

	end := result[0].End
	assert.NotNil(s.T(), end, "End field must be found")
	assert.Equal(s.T(), *end, 847492)

	mapq0reads := result[0].MAPQ0Reads
	assert.NotNil(s.T(), mapq0reads, "MAPQ0Reads field must be found")
	assert.Equal(s.T(), *mapq0reads, 0)

	numSamples := result[0].NumberOfSamples
	assert.NotNil(s.T(), numSamples, "NumberOfSamples field must be found")
	assert.Equal(s.T(), *numSamples, 27)

	mq := result[0].MappingQuality
	assert.NotNil(s.T(), mq, "MappingQuality field must be found")
	assert.Equal(s.T(), *mq, 60.0)

	cigar := result[0].Cigar
	assert.NotNil(s.T(), cigar, "Cigar field must be found")
	assert.Equal(s.T(), *cigar, "a")

	dbsnp := result[0].InDBSNP
	assert.NotNil(s.T(), dbsnp, "InDBSNP field must be found")
	assert.True(s.T(), *dbsnp)

	h2 := result[0].InHapmap2
	assert.NotNil(s.T(), h2, "InHapmap2 field must be found")
	assert.True(s.T(), *h2)

	h3 := result[0].InHapmap3
	assert.NotNil(s.T(), h3, "InHapmap3 field must be found")
	assert.True(s.T(), *h3)

	somatic := result[0].IsSomatic
	assert.NotNil(s.T(), somatic, "IsSomatic field must be found")
	assert.True(s.T(), *somatic)

	validated := result[0].IsValidated
	assert.NotNil(s.T(), validated, "IsValidated field must be found")
	assert.True(s.T(), *validated)

	thousand := result[0].In1000G
	assert.NotNil(s.T(), thousand, "In1000G field must be found")
	assert.True(s.T(), *thousand)

	bq := result[0].BaseQuality
	assert.NotNil(s.T(), bq, "BaseQuality field must be found")
	assert.Equal(s.T(), *bq, 30.0)

	strandBias := result[0].StrandBias
	assert.NotNil(s.T(), strandBias, "StrandBias field must be found")
	assert.Equal(s.T(), *strandBias, 0.127)
}

func (s *ParseVcfLineSuite) TestAncestralAllele() {
	result, _ := parseVcfLine("1\t847491\trs28407778\tG\tA,C\t745.77\tPASS\tAC=1;AF=0.500,0.335;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;AA=T;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant\tGT:AD:DP:GQ:PL\t0/1:16,25:41:99:774,0,434", defaultHeader)

	aa := result[0].AncestralAllele
	assert.NotNil(s.T(), aa, "AncestralAllele field must be found")
	assert.Equal(s.T(), *aa, "T")
}

func (s *ParseVcfLineSuite) TestAlternateFormatOptionalField() {
	var result []*Variant
	var err error

	assert.NotPanics(s.T(), func() {
		result, err = parseVcfLine("1\t847491\trs28407778\tG\tA\t745.77\tPASS\tSB=strong;AA\tGT:AD:DP:GQ:PL\t0/1:16,25:41:99:774,0,434", defaultHeader)
	})

	assert.NoError(s.T(), err, "Valid VCF line should not return error")
	assert.NotNil(s.T(), result, "Valid VCF line should not return nil")

	info := result[0].Info
	assert.NotNil(s.T(), info, "Valid VCF should contain info map")
	assert.Exactly(s.T(), len(info), 2, "Info should contain 2 keys")

	sb, ok := info["SB"]
	assert.True(s.T(), ok, "SB key must be found")
	assert.Equal(s.T(), sb, "strong")

	aa, ok := info["AA"]
	assert.True(s.T(), ok, "AA key must be found")
	boolaa, isbool := aa.(bool)
	assert.True(s.T(), isbool, "AA value must be a boolean")
	assert.True(s.T(), boolaa)
}

func TestParseVcfLineSuite(t *testing.T) {
	suite.Run(t, new(ParseVcfLineSuite))
}

func (s *ParseVcfLineSuite) TestValidCNVShouldReturnOneElementAndNoErrors() {
	result, err := parseVcfLine("22\t16533236\tSI_BD_17525\tC\t<CN0>\t100\tPASS\tAC=125;AF=0.0249601;AN=5008;CIEND=-50,141;CIPOS=-141,50;CS=DEL_union;END=16536204;NS=2504;SVLEN=-2968;SVTYPE=DEL;DP=14570;EAS_AF=0;AMR_AF=0.0086;AFR_AF=0.09;EUR_AF=0;SAS_AF=0\tGT", defaultHeader)

	assert.NoError(s.T(), err, "Valid VCF line should not return error")
	assert.NotNil(s.T(), result, "Valid VCF line should not return nil")
	assert.Exactly(s.T(), len(result), 1, "Valid VCF should return a list with one element")
	assert.Equal(s.T(), result[0].Chrom, "22", "result.Chrom should be 2")
	assert.Equal(s.T(), result[0].Pos, 16533235, "result.Pos should be 0-based to 16533235")
	assert.Equal(s.T(), result[0].Ref, "C", "result.Ref should be C")
	assert.Equal(s.T(), result[0].Alt, "<CN0>", "result.Alt should be A")
}

func (s *ParseVcfLineSuite) TestInfoFieldsWhenCNV() {
	result, _ := parseVcfLine("22\t16533236\tSI_BD_17525\tC\t<CN0>\t100\tPASS\tAC=125;AF=0.0249601;AN=5008;CIEND=-50,141;CIPOS=-141,50;CS=DEL_union;END=16536204;NS=2504;SVLEN=-2968;SVTYPE=DEL;DP=14570;EAS_AF=0;AMR_AF=0.0086;AFR_AF=0.09;EUR_AF=0;SAS_AF=0\tGT", defaultHeader)

	c1 := result[0].SVType
	assert.NotNil(s.T(), c1, "SVType field must be found")
	// assert.True(s.T(), *c1)
	assert.Equal(s.T(), *c1, "DEL", "result.SVTYPE should be DEL")

	c2 := result[0].SVLength
	assert.NotNil(s.T(), c2, "SVLength field must be found")
	// assert.True(s.T(), *c2)
	assert.Equal(s.T(), *c2, -2968, "result.SVLEN should be -2968")

	c3 := result[0].End
	assert.NotNil(s.T(), c3, "End field must be found")
	// assert.True(s.T(), *c3)
	assert.Equal(s.T(), *c3, 16536204, "result.End should be 1-based to 16536204")
}

type FixSuffixSuite struct {
	suite.Suite
}

func (s *FixSuffixSuite) TestNoSuffix() {
	variant := Variant{
		Ref: "T",
		Alt: "C",
	}
	result := fixRefAltSuffix(&variant)
	assert.Equal(s.T(), variant.Ref, result.Ref, "no suffix in common should return the same ref")
	assert.Equal(s.T(), variant.Alt, result.Alt, "no suffix in common should return the same alt")
}

func (s *FixSuffixSuite) TestSmallSuffix() {
	variant := Variant{
		Ref: "GC",
		Alt: "TC",
	}
	result := fixRefAltSuffix(&variant)
	assert.Equal(s.T(), "G", result.Ref, "GC -> TC should become ref G")
	assert.Equal(s.T(), "T", result.Alt, "GC -> TC should become alt T")
}

func (s *FixSuffixSuite) TestBigSuffix() {
	variant := Variant{
		Ref: "CGGCCACGTCCCCCTATGGAGGG",
		Alt: "TGGCCACGTCCCCCTATGGAGGG",
	}
	result := fixRefAltSuffix(&variant)
	assert.Equal(s.T(), "C", result.Ref, "CGGCCACGTCCCCCTATGGAGGG -> TGGCCACGTCCCCCTATGGAGGG should become ref C")
	assert.Equal(s.T(), "T", result.Alt, "CGGCCACGTCCCCCTATGGAGGG -> TGGCCACGTCCCCCTATGGAGGG should become alt T")
}

func (s *FixSuffixSuite) TestBigSuffixWithBigResult() {
	variant := Variant{
		Ref: "CGGCCACGTCCCCCTATGGAGGG",
		Alt: "CGGCCACGTCCCCCTATGGAGGGGGCCACGTCCCCCTATGGAGGG",
	}
	result := fixRefAltSuffix(&variant)
	assert.Equal(s.T(), "C", result.Ref, "CGGCCACGTCCCCCTATGGAGGG -> CGGCCACGTCCCCCTATGGAGGGGGCCACGTCCCCCTATGGAGGG should become ref C")
	assert.Equal(s.T(), "CGGCCACGTCCCCCTATGGAGGG", result.Alt, "CGGCCACGTCCCCCTATGGAGGG -> CGGCCACGTCCCCCTATGGAGGGGGCCACGTCCCCCTATGGAGGG should become alt T")
}

func TestFixSuffixSuite(t *testing.T) {
	suite.Run(t, new(FixSuffixSuite))
}
