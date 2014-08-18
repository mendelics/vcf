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

func (s *ParseVcfLineSuite) TestValidLineWithLowecaseRefAndAltShouldReturnOneElementAndNoErrors() {
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
	result, err := parseVcfLine("1\t847491\trs28407778\tG\tA,C\t745.77\tPASS\tAC=1;AF=0.500,0.335;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant\tGT:AD:DP:GQ:PL\t0/1:16,25:41:99:774,0,434", defaultHeader)

	assert.NoError(s.T(), err, "Valid VCF line should not return error")
	assert.NotNil(s.T(), result, "Valid VCF line should not return nil")
	assert.Exactly(s.T(), len(result), 2, "Valid VCF should return a list with two elements")

	info := result[0].Info
	assert.NotNil(s.T(), info, "Valid VCF should contain info map")
	assert.Exactly(s.T(), len(info), 18, "Info should contain 18 keys")

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
	assert.NotNil(s.T(), dp, "Depth field must be found")
	assert.Equal(s.T(), *dp, 41)

	freq := result[0].AlleleFrequency
	assert.NotNil(s.T(), freq, "AlleleFrequency field must be found")
	assert.Equal(s.T(), *freq, 0.500)
	freq = result[1].AlleleFrequency
	assert.NotNil(s.T(), freq, "AlleleFrequency field must be found")
	assert.Equal(s.T(), *freq, 0.335)
}

func (s *ParseVcfLineSuite) TestAncestralAllele() {
	result, _ := parseVcfLine("1\t847491\trs28407778\tG\tA,C\t745.77\tPASS\tAC=1;AF=0.500,0.335;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;AA=T;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant\tGT:AD:DP:GQ:PL\t0/1:16,25:41:99:774,0,434", defaultHeader)

	aa := result[0].AncestralAllele
	assert.NotNil(s.T(), aa, "AncestralAllele field must be found")
	assert.Equal(s.T(), *aa, "T")
}

func TestParseVcfLineSuite(t *testing.T) {
	suite.Run(t, new(ParseVcfLineSuite))
}
