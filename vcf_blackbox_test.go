package vcf_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mendelics/vcf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ChannelSuite struct {
	suite.Suite

	outChannel     chan *vcf.Variant
	invalidChannel chan vcf.InvalidLine
}

func (suite *ChannelSuite) SetupTest() {
	suite.outChannel = make(chan *vcf.Variant, 10)
	suite.invalidChannel = make(chan vcf.InvalidLine, 10)
}

func (s *ChannelSuite) TestNoHeader() {
	vcfLine := `1	847491	rs28407778	GTTTA	G....	745.77	PASS	AC=1;AF=0.500;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant	GT:AD:DP:GQ:PL	0/1:16,25:41:99:774,0,434`
	ioreader := strings.NewReader(vcfLine)
	err := vcf.ToChannel(ioreader, s.outChannel, s.invalidChannel)

	assert.Error(s.T(), err, "VCF line without header should return error")
}

func (s *ChannelSuite) TestInvalidLinesShouldReturnNothing() {
	vcfLine := `#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT	185423
	
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nunc tellus ligula, faucibus sed nibh sed, fringilla viverra enim.
					
A	B	C	D	E	F`

	ioreader := strings.NewReader(vcfLine)
	err := vcf.ToChannel(ioreader, s.outChannel, s.invalidChannel)
	assert.NoError(s.T(), err, "VCF with valid header should not return an error")

	_, hasMore := <-s.outChannel
	assert.False(s.T(), hasMore, "No variant should come out of the channel, it should be closed")

	totalLines := 4
	for i := 0; i < totalLines; i++ {
		invalid := <-s.invalidChannel
		assert.NotNil(s.T(), invalid)
		assert.Error(s.T(), invalid.Err)
	}

	_, hasMore = <-s.invalidChannel
	assert.False(s.T(), hasMore, fmt.Sprintf("More than %d variants came out of the invalid channel, it should be closed", totalLines))
}

func (s *ChannelSuite) TestToChannel() {
	vcfLine := `#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT	185423
1	847491	rs28407778	GTTTA	G....	745.77	PASS	AC=1;AF=0.500;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant	GT:AD:DP:GQ:PL	0/1:16,25:41:99:774,0,434`
	ioreader := strings.NewReader(vcfLine)

	err := vcf.ToChannel(ioreader, s.outChannel, s.invalidChannel)
	assert.NoError(s.T(), err, "Valid VCF line should not return error")

	variant := <-s.outChannel
	assert.NotNil(s.T(), variant, "One variant should come out of channel")

	assert.Equal(s.T(), variant.Chrom, "1")
	assert.Equal(s.T(), variant.Pos, 847490)
	assert.Equal(s.T(), variant.Ref, "GTTTA")
	assert.Equal(s.T(), variant.Alt, "G")
	assert.Equal(s.T(), variant.ID, "rs28407778")
	assert.Equal(s.T(), *variant.Qual, 745.77)
	assert.Equal(s.T(), variant.Filter, "PASS")

	assert.NotNil(s.T(), variant.Info)
	assert.Exactly(s.T(), len(variant.Info), 18)
	ac, ok := variant.Info["AC"]
	assert.True(s.T(), ok, "AC key must be found")
	assert.Equal(s.T(), ac, "1", "ac")
	af, ok := variant.Info["AF"]
	assert.True(s.T(), ok, "AF key must be found")
	assert.Equal(s.T(), af, "0.500", "af")
	db, ok := variant.Info["DB"]
	assert.True(s.T(), ok, "DB key must be found")
	booldb, isbool := db.(bool)
	assert.True(s.T(), isbool, "DB value must be a boolean")
	assert.True(s.T(), booldb)

	assert.NotNil(s.T(), variant.Samples)
	assert.Exactly(s.T(), len(variant.Samples), 1, "Valid VCF should contain one sample")
	sampleMap := variant.Samples[0]
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

	_, hasMore := <-s.outChannel
	assert.False(s.T(), hasMore, "No second variant should come out of the channel, it should be closed")
	_, hasMore = <-s.invalidChannel
	assert.False(s.T(), hasMore, "No variant should come out of invalid channel, it should be closed")
}

func (s *ChannelSuite) TestChrParsedProperly() {
	vcfLine := `#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT	185423
chr1	847491	rs28407778	GTTTA	G....	745.77	PASS	AC=1;AF=0.500;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant	GT:AD:DP:GQ:PL	0/1:16,25:41:99:774,0,434`
	ioreader := strings.NewReader(vcfLine)

	err := vcf.ToChannel(ioreader, s.outChannel, s.invalidChannel)
	assert.NoError(s.T(), err, "Valid VCF line should not return error")

	variant := <-s.outChannel
	assert.NotNil(s.T(), variant, "One variant should come out of channel")

	assert.Equal(s.T(), variant.Chrom, "1")

	_, hasMore := <-s.outChannel
	assert.False(s.T(), hasMore, "No second variant should come out of the channel, it should be closed")
	_, hasMore = <-s.invalidChannel
	assert.False(s.T(), hasMore, "No variant should come out of invalid channel, it should be closed")
}

func (s *ChannelSuite) TestLowercaseRefAlt() {
	vcfLine := `#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT	185423
1	847491	rs28407778	gt	t	745.77	PASS	AC=1;AF=0.500;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant	GT:AD:DP:GQ:PL	0/1:16,25:41:99:774,0,434`
	ioreader := strings.NewReader(vcfLine)

	err := vcf.ToChannel(ioreader, s.outChannel, s.invalidChannel)
	assert.NoError(s.T(), err, "Valid VCF line should not return error")

	variant := <-s.outChannel
	assert.NotNil(s.T(), variant, "One variant should come out of channel")
	assert.Equal(s.T(), variant.Ref, "GT")
	assert.Equal(s.T(), variant.Alt, "T")

	_, hasMore := <-s.outChannel
	assert.False(s.T(), hasMore, "No second variant should come out of the channel, it should be closed")
	_, hasMore = <-s.invalidChannel
	assert.False(s.T(), hasMore, "No variant should come out of invalid channel, it should be closed")
}

func (s *ChannelSuite) TestMultipleAlternatives() {
	vcfLine := `#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT	185423
1	847491	rs28407778	G	A,C	745.77	PASS	AC=1;AF=0.500;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant	GT:AD:DP:GQ:PL	0/1:16,25:41:99:774,0,434`
	ioreader := strings.NewReader(vcfLine)

	err := vcf.ToChannel(ioreader, s.outChannel, s.invalidChannel)
	assert.NoError(s.T(), err, "Valid VCF line should not return error")

	variant := <-s.outChannel
	assert.NotNil(s.T(), variant, "One variant should come out of channel")
	assert.Equal(s.T(), variant.Alt, "A")
	variant = <-s.outChannel
	assert.NotNil(s.T(), variant, "Second variant should come out of channel")
	assert.Equal(s.T(), variant.Alt, "C")

	_, hasMore := <-s.outChannel
	assert.False(s.T(), hasMore, "No third variant should come out of the channel, it should be closed")
	_, hasMore = <-s.invalidChannel
	assert.False(s.T(), hasMore, "No variant should come out of invalid channel, it should be closed")
}

func TestChannelSuite(t *testing.T) {
	suite.Run(t, new(ChannelSuite))
}

type SampleSuite struct {
	suite.Suite
}

func (s *SampleSuite) TestNoHeader() {
	vcfLine := `1	847491	rs28407778	GTTTA	G....	745.77	PASS	AC=1;AF=0.500;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant	GT:AD:DP:GQ:PL	0/1:16,25:41:99:774,0,434`
	ioreader := strings.NewReader(vcfLine)
	sampleIDs, err := vcf.SampleIDs(ioreader)

	assert.Error(s.T(), err, "VCF without header should return error")
	assert.Nil(s.T(), sampleIDs, "No slice of ids is expected on a vcf without header")
}

func (s *SampleSuite) TestValidHeaderNoSample() {
	vcfLine := `#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT
1	847491	rs28407778	GTTTA	G....	745.77	PASS	AC=1;AF=0.500;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant	GT:AD:DP:GQ:PL	0/1:16,25:41:99:774,0,434`
	ioreader := strings.NewReader(vcfLine)
	sampleIDs, err := vcf.SampleIDs(ioreader)

	assert.NoError(s.T(), err, "VCF with valid header should not return error")
	assert.Nil(s.T(), sampleIDs, "No slice of ids should be returned on a vcf with a valid header that doesn't contain any sample")
}

func (s *SampleSuite) TestOneSample() {
	vcfLine := `#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT	185423
1	847491	rs28407778	GTTTA	G....	745.77	PASS	AC=1;AF=0.500;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant	GT:AD:DP:GQ:PL	0/1:16,25:41:99:774,0,434`
	ioreader := strings.NewReader(vcfLine)
	sampleIDs, err := vcf.SampleIDs(ioreader)

	assert.NoError(s.T(), err, "VCF with valid header should not return error")
	assert.NotNil(s.T(), sampleIDs, "A slice of ids should be returned on a vcf with a valid header")
	assert.Exactly(s.T(), len(sampleIDs), 1, "Slice of ids should have only one element")
	assert.Equal(s.T(), sampleIDs[0], "185423")
}

func (s *SampleSuite) TestThreeSamples() {
	vcfLine := `#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT	185423	776182	091635
1	847491	rs28407778	GTTTA	G....	745.77	PASS	AC=1;AF=0.500;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant	GT:AD:DP:GQ:PL	0/1:16,25:41:99:774,0,434`
	ioreader := strings.NewReader(vcfLine)
	sampleIDs, err := vcf.SampleIDs(ioreader)

	assert.NoError(s.T(), err, "VCF with valid header should not return error")
	assert.NotNil(s.T(), sampleIDs, "A slice of ids should be returned on a vcf with a valid header")
	assert.Exactly(s.T(), len(sampleIDs), 3, "Slice of ids should have three elements")
	assert.Equal(s.T(), sampleIDs[0], "185423")
	assert.Equal(s.T(), sampleIDs[1], "776182")
	assert.Equal(s.T(), sampleIDs[2], "091635")
}

func TestSampleSuite(t *testing.T) {
	suite.Run(t, new(SampleSuite))
}

type InfoSuite struct {
	suite.Suite

	outChannel     chan *vcf.Variant
	invalidChannel chan vcf.InvalidLine
}

func (suite *InfoSuite) SetupTest() {
	suite.outChannel = make(chan *vcf.Variant, 10)
	suite.invalidChannel = make(chan vcf.InvalidLine, 10)
}

func (s *InfoSuite) TestInfo() {
	vcfLine := `#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT	185423
1	847491	rs28407778	GTTTA	G....	745.77	PASS	AC=1;AF=0.500;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant	GT:AD:DP:GQ:PL	0/1:16,25:41:99:774,0,434`
	ioreader := strings.NewReader(vcfLine)

	err := vcf.ToChannel(ioreader, s.outChannel, s.invalidChannel)
	assert.NoError(s.T(), err, "Valid VCF line should not return error")

	variant := <-s.outChannel
	assert.NotNil(s.T(), variant, "One variant should come out of channel")

	assert.NotNil(s.T(), variant.AlleleCount)
	assert.Equal(s.T(), *variant.AlleleCount, 1)
	assert.NotNil(s.T(), variant.AlleleFrequency)
	assert.Equal(s.T(), *variant.AlleleFrequency, 0.5)
	assert.NotNil(s.T(), variant.TotalAlleles)
	assert.Equal(s.T(), *variant.TotalAlleles, 2)
	assert.NotNil(s.T(), variant.InDBSNP)
	assert.True(s.T(), *variant.InDBSNP)
	assert.NotNil(s.T(), variant.Depth)
	assert.Equal(s.T(), *variant.Depth, 41)
	assert.NotNil(s.T(), variant.MappingQuality)
	assert.Equal(s.T(), *variant.MappingQuality, 60.0)
	assert.NotNil(s.T(), variant.MAPQ0Reads)
	assert.Equal(s.T(), *variant.MAPQ0Reads, 0)

	assert.Nil(s.T(), variant.AncestralAllele)
	assert.Nil(s.T(), variant.BaseQuality)
	assert.Nil(s.T(), variant.Cigar)
	assert.Nil(s.T(), variant.End)
	assert.Nil(s.T(), variant.InHapmap2)
	assert.Nil(s.T(), variant.InHapmap3)
	assert.Nil(s.T(), variant.NumberOfSamples)
	assert.Nil(s.T(), variant.StrandBias)
	assert.Nil(s.T(), variant.IsSomatic)
	assert.Nil(s.T(), variant.IsValidated)
	assert.Nil(s.T(), variant.In1000G)

	_, hasMore := <-s.outChannel
	assert.False(s.T(), hasMore, "No second variant should come out of the channel, it should be closed")
	_, hasMore = <-s.invalidChannel
	assert.False(s.T(), hasMore, "No variant should come out of invalid channel, it should be closed")
}

func (s *InfoSuite) TestMultiple() {
	vcfLine := `#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT	185423
5	159478089	rs80263784	GTT	G,GT	198.19	.	AC=1,2;AF=0.500,0.600;AN=2;BaseQRankSum=1.827;ClippingRankSum=1.323;DB;DP=20;FS=0.000;MLEAC=1,1;MLEAF=0.500,0.500;MQ=60.00;MQ0=0;MQRankSum=0.441;QD=5.74;ReadPosRankSum=0.063;set=variant5	GT:AD:DP:GQ:PL  1/2:2,9,9:20:99:425,145,183,175,0,166
5	159478089	rs80263784	GTT	G,GT	198.19	.	AC=1,2;AF=0.500,0.600;AN=3,4;BaseQRankSum=1.827;ClippingRankSum=1.323;DB;DP=20;FS=0.000;MLEAC=1,1;MLEAF=0.500,0.500;MQ=60.00;MQ0=0;MQRankSum=0.441;QD=5.74;ReadPosRankSum=0.063;set=variant5	GT:AD:DP:GQ:PL  1/2:2,9,9:20:99:425,145,183,175,0,166`
	ioreader := strings.NewReader(vcfLine)

	err := vcf.ToChannel(ioreader, s.outChannel, s.invalidChannel)
	assert.NoError(s.T(), err, "Valid VCF line should not return error")

	// first variant
	variant := <-s.outChannel
	assert.NotNil(s.T(), variant, "One variant should come out of channel")
	assert.NotNil(s.T(), variant.AlleleCount)
	assert.Equal(s.T(), *variant.AlleleCount, 1)
	assert.NotNil(s.T(), variant.AlleleFrequency)
	assert.Equal(s.T(), *variant.AlleleFrequency, 0.5)
	assert.NotNil(s.T(), variant.TotalAlleles)
	assert.Equal(s.T(), *variant.TotalAlleles, 2)

	// second variant
	variant, hasMore := <-s.outChannel
	assert.True(s.T(), hasMore, "Second variant should be in the channel")
	assert.NotNil(s.T(), variant, "Second variant should come out of channel")
	assert.NotNil(s.T(), variant.AlleleCount)
	assert.Equal(s.T(), *variant.AlleleCount, 2)
	assert.NotNil(s.T(), variant.AlleleFrequency)
	assert.Equal(s.T(), *variant.AlleleFrequency, 0.6)
	assert.NotNil(s.T(), variant.TotalAlleles)
	assert.Equal(s.T(), *variant.TotalAlleles, 2)

	// third variant
	variant, hasMore = <-s.outChannel
	assert.True(s.T(), hasMore, "Third variant should be in the channel")
	assert.NotNil(s.T(), variant, "Third variant should come out of channel")
	assert.NotNil(s.T(), variant.TotalAlleles)
	assert.Equal(s.T(), *variant.TotalAlleles, 3)

	// fourth variant
	variant, hasMore = <-s.outChannel
	assert.True(s.T(), hasMore, "Fourth variant should be in the channel")
	assert.NotNil(s.T(), variant, "Fourth variant should come out of channel")
	assert.NotNil(s.T(), variant.TotalAlleles)
	assert.Equal(s.T(), *variant.TotalAlleles, 4)

	_, hasMore = <-s.outChannel
	assert.False(s.T(), hasMore, "No fifth variant should come out of the channel, it should be closed")
	_, hasMore = <-s.invalidChannel
	assert.False(s.T(), hasMore, "No variant should come out of invalid channel, it should be closed")
}

func TestInfoSuite(t *testing.T) {
	suite.Run(t, new(InfoSuite))
}

type FixSuffixSuite struct {
	suite.Suite

	outChannel     chan *vcf.Variant
	invalidChannel chan vcf.InvalidLine
}

func (suite *FixSuffixSuite) SetupTest() {
	suite.outChannel = make(chan *vcf.Variant, 10)
	suite.invalidChannel = make(chan vcf.InvalidLine, 10)
}

func (s *FixSuffixSuite) TestSimpleVariant() {
	vcfLine := `#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT	185423
1	138829	.	GC	TC,G	198.19	.	AC=1,2;AF=0.500,0.600;AN=2;BaseQRankSum=1.827;ClippingRankSum=1.323;DB;DP=20;FS=0.000;MLEAC=1,1;MLEAF=0.500,0.500;MQ=60.00;MQ0=0;MQRankSum=0.441;QD=5.74;ReadPosRankSum=0.063;set=variant5	GT:AD:DP:GQ:PL  1/2:2,9,9:20:99:425,145,183,175,0,166`
	ioreader := strings.NewReader(vcfLine)

	err := vcf.ToChannel(ioreader, s.outChannel, s.invalidChannel)
	assert.NoError(s.T(), err, "Valid VCF line should not return error")

	// first variant
	variant := <-s.outChannel
	assert.NotNil(s.T(), variant, "One variant should come out of channel")
	assert.Equal(s.T(), variant.Ref, "G")
	assert.Equal(s.T(), variant.Alt, "T")

	// second variant
	variant, hasMore := <-s.outChannel
	assert.True(s.T(), hasMore, "Second variant should be in the channel")
	assert.NotNil(s.T(), variant, "Second variant should come out of channel")
	assert.Equal(s.T(), variant.Ref, "GC")
	assert.Equal(s.T(), variant.Alt, "G")

	_, hasMore = <-s.outChannel
	assert.False(s.T(), hasMore, "No third variant should come out of the channel, it should be closed")
	_, hasMore = <-s.invalidChannel
	assert.False(s.T(), hasMore, "No variant should come out of invalid channel, it should be closed")
}

func (s *FixSuffixSuite) TestBigSuffix() {
	vcfLine := `#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT	185423
1	879415	.	CGGCCACGTCCCCCTATGGAGGG	C,TGGCCACGTCCCCCTATGGAGGG,CGGCCACGTCCCCCTATGGAGGGGGCCACGTCCCCCTATGGAGGG	198.19	.	AC=1,2;AF=0.500,0.600;AN=2;BaseQRankSum=1.827;ClippingRankSum=1.323;DB;DP=20;FS=0.000;MLEAC=1,1;MLEAF=0.500,0.500;MQ=60.00;MQ0=0;MQRankSum=0.441;QD=5.74;ReadPosRankSum=0.063;set=variant5	GT:AD:DP:GQ:PL  1/2:2,9,9:20:99:425,145,183,175,0,166`
	ioreader := strings.NewReader(vcfLine)

	err := vcf.ToChannel(ioreader, s.outChannel, s.invalidChannel)
	assert.NoError(s.T(), err, "Valid VCF line should not return error")

	// first variant
	variant := <-s.outChannel
	assert.NotNil(s.T(), variant, "One variant should come out of channel")
	assert.Equal(s.T(), variant.Ref, "CGGCCACGTCCCCCTATGGAGGG")
	assert.Equal(s.T(), variant.Alt, "C")

	// second variant
	variant, hasMore := <-s.outChannel
	assert.True(s.T(), hasMore, "Second variant should be in the channel")
	assert.NotNil(s.T(), variant, "Second variant should come out of channel")
	assert.Equal(s.T(), variant.Ref, "C")
	assert.Equal(s.T(), variant.Alt, "T")

	// third variant
	variant, hasMore = <-s.outChannel
	assert.True(s.T(), hasMore, "Third variant should be in the channel")
	assert.NotNil(s.T(), variant, "Third variant should come out of channel")
	assert.Equal(s.T(), variant.Ref, "C")
	assert.Equal(s.T(), variant.Alt, "CGGCCACGTCCCCCTATGGAGGG")

	_, hasMore = <-s.outChannel
	assert.False(s.T(), hasMore, "No fourth variant should come out of the channel, it should be closed")
	_, hasMore = <-s.invalidChannel
	assert.False(s.T(), hasMore, "No variant should come out of invalid channel, it should be closed")
}

func TestFixSuffixSuite(t *testing.T) {
	suite.Run(t, new(FixSuffixSuite))
}

type StructuralSuite struct {
	suite.Suite

	outChannel     chan *vcf.Variant
	invalidChannel chan vcf.InvalidLine
}

func (suite *StructuralSuite) SetupTest() {
	suite.outChannel = make(chan *vcf.Variant, 10)
	suite.invalidChannel = make(chan vcf.InvalidLine, 10)
}

func (s *StructuralSuite) TestNoSpecificStructuralVariantFieldsSet() {
	vcfLine := `#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT	185423
1	847491	CNVR8241.1	G	A	745.77	PASS	AC=1	GT	0/1`
	ioreader := strings.NewReader(vcfLine)

	err := vcf.ToChannel(ioreader, s.outChannel, s.invalidChannel)
	assert.NoError(s.T(), err, "Valid VCF line should not return error")

	variant := <-s.outChannel
	assert.NotNil(s.T(), variant, "One variant should come out of channel")

	assert.Equal(s.T(), variant.Chrom, "1")
	assert.Equal(s.T(), variant.Ref, "G")
	assert.Equal(s.T(), variant.Alt, "A")
	assert.Equal(s.T(), *variant.Qual, 745.77)
	assert.Equal(s.T(), variant.Filter, "PASS")

	assert.NotNil(s.T(), variant.Info)
	assert.Exactly(s.T(), len(variant.Info), 1)
	ac, ok := variant.Info["AC"]
	assert.True(s.T(), ok, "AC key must be found")
	assert.Equal(s.T(), ac, "1", "ac")

	assert.Nil(s.T(), variant.Imprecise)
	assert.Nil(s.T(), variant.Novel)
	assert.Nil(s.T(), variant.End)
	assert.Nil(s.T(), variant.StructuralVariantType)
	assert.Nil(s.T(), variant.StructuralVariantLength)
	assert.Nil(s.T(), variant.ConfidenceIntervalAroundPosition)
	assert.Nil(s.T(), variant.ConfidenceIntervalAroundEnd)

	_, hasMore := <-s.outChannel
	assert.False(s.T(), hasMore, "No second variant should come out of the channel, it should be closed")
	_, hasMore = <-s.invalidChannel
	assert.False(s.T(), hasMore, "No variant should come out of invalid channel, it should be closed")
}

func (s *StructuralSuite) TestImpreciseNovel() {
	vcfLine := `#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT	185423
1	847491	CNVR8241.1	G	A	745.77	PASS	IMPRECISE;NOVEL	GT	0/1`
	ioreader := strings.NewReader(vcfLine)

	err := vcf.ToChannel(ioreader, s.outChannel, s.invalidChannel)
	assert.NoError(s.T(), err, "Valid VCF line should not return error")

	variant := <-s.outChannel
	assert.NotNil(s.T(), variant, "One variant should come out of channel")

	assert.Equal(s.T(), variant.Chrom, "1")
	assert.Equal(s.T(), variant.Ref, "G")
	assert.Equal(s.T(), variant.Alt, "A")
	assert.Equal(s.T(), *variant.Qual, 745.77)
	assert.Equal(s.T(), variant.Filter, "PASS")

	assert.NotNil(s.T(), variant.Imprecise)
	assert.True(s.T(), *variant.Imprecise)
	assert.NotNil(s.T(), variant.Novel)
	assert.True(s.T(), *variant.Novel)

	_, hasMore := <-s.outChannel
	assert.False(s.T(), hasMore, "No second variant should come out of the channel, it should be closed")
	_, hasMore = <-s.invalidChannel
	assert.False(s.T(), hasMore, "No variant should come out of invalid channel, it should be closed")
}

func (s *StructuralSuite) TestInfoEnd() {
	vcfLine := `#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT	185423
1	847491	CNVR8241.1	G	A	755.77	PASS	END=1752234	GT	0/1
1	847491	rs28407778	GTTTA	G....	745.77	PASS	AC=1;AF=0.500;AN=2;BaseQRankSum=0.842;ClippingRankSum=0.147;DB;DP=41;FS=0.000;MLEAC=1;MLEAF=0.500;MQ=60.00;MQ0=0;MQRankSum=-1.109;QD=18.19;ReadPosRankSum=0.334;VQSLOD=2.70;culprit=FS;set=variant	GT:AD:DP:GQ:PL	0/1:16,25:41:99:774,0,434`
	ioreader := strings.NewReader(vcfLine)

	err := vcf.ToChannel(ioreader, s.outChannel, s.invalidChannel)
	assert.NoError(s.T(), err, "Valid VCF line should not return error")

	variant := <-s.outChannel
	assert.NotNil(s.T(), variant, "One variant should come out of channel")

	assert.Equal(s.T(), variant.Chrom, "1")
	assert.Equal(s.T(), variant.Ref, "G")
	assert.Equal(s.T(), variant.Alt, "A")
	assert.Equal(s.T(), *variant.Qual, 755.77)
	assert.Equal(s.T(), variant.Filter, "PASS")

	assert.NotNil(s.T(), variant.End)
	assert.Equal(s.T(), *variant.End, 1752234)

	variant = <-s.outChannel
	assert.NotNil(s.T(), variant, "Second variant should come out of channel")

	assert.Equal(s.T(), variant.Chrom, "1")
	assert.Equal(s.T(), variant.Ref, "GTTTA")
	assert.Equal(s.T(), variant.Alt, "G")
	assert.Equal(s.T(), *variant.Qual, 745.77)
	assert.Equal(s.T(), variant.Filter, "PASS")

	assert.Nil(s.T(), variant.End)

	_, hasMore := <-s.outChannel
	assert.False(s.T(), hasMore, "No third variant should come out of the channel, it should be closed")
	_, hasMore = <-s.invalidChannel
	assert.False(s.T(), hasMore, "No variant should come out of invalid channel, it should be closed")
}

func (s *StructuralSuite) TestSVType() {
	vcfLine := `#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT	185423
1	847491	CNVR8241.1	G	A	755.77	PASS	SVTYPE=DEL	GT	0/1
1	847491	CNVR8241.1	G	A	755.77	PASS	SVTYPE=DUP	GT	0/1`
	ioreader := strings.NewReader(vcfLine)

	err := vcf.ToChannel(ioreader, s.outChannel, s.invalidChannel)
	assert.NoError(s.T(), err, "Valid VCF line should not return error")

	variant := <-s.outChannel
	assert.NotNil(s.T(), variant, "One variant should come out of channel")

	assert.Equal(s.T(), variant.Chrom, "1")
	assert.Equal(s.T(), variant.Ref, "G")
	assert.Equal(s.T(), variant.Alt, "A")
	assert.Equal(s.T(), *variant.Qual, 755.77)
	assert.Equal(s.T(), variant.Filter, "PASS")

	assert.NotNil(s.T(), variant.StructuralVariantType)
	assert.Equal(s.T(), *variant.StructuralVariantType, vcf.Deletion)

	variant = <-s.outChannel
	assert.NotNil(s.T(), variant)
	assert.NotNil(s.T(), variant.StructuralVariantType)
	assert.Equal(s.T(), *variant.StructuralVariantType, vcf.Duplication)

	_, hasMore := <-s.outChannel
	assert.False(s.T(), hasMore, "No more variants should come out of the channel, it should be closed")
	_, hasMore = <-s.invalidChannel
	assert.False(s.T(), hasMore, "No variant should come out of invalid channel, it should be closed")
}

func TestStructuralSuite(t *testing.T) {
	suite.Run(t, new(StructuralSuite))
}
