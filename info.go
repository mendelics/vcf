package vcf

import (
	"strconv"
	"strings"
)

func infoToMap(info string) map[string]interface{} {
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

func splitInfo(variant *Variant) {
	info := variant.Info
	variant.Depth = infoInt("DP", info)
	variant.AlleleFrequency = infoFloat("AF", info)
	variant.AncestralAllele = infoString("AA", info)
	variant.AlleleCount = infoInt("AC", info)
	variant.TotalAlleles = infoInt("AN", info)
	variant.End = infoInt("END", info)
	variant.MAPQ0Reads = infoInt("MQ0", info)
	variant.NumberOfSamples = infoInt("NS", info)
	variant.MappingQuality = infoFloat("MQ", info)
	variant.Cigar = infoString("CIGAR", info)
	variant.InDBSNP = infoBool("DB", info)
	variant.InHapmap2 = infoBool("H2", info)
	variant.InHapmap3 = infoBool("H3", info)
	variant.IsSomatic = infoBool("SOMATIC", info)
	variant.IsValidated = infoBool("VALIDATED", info)
	variant.In1000G = infoBool("1000G", info)
	variant.BaseQuality = infoFloat("BQ", info)
	variant.StrandBias = infoFloat("SB", info)
}

func infoInt(key string, info map[string]interface{}) *int {
	if value, found := info[key]; found {
		if str, ok := value.(string); ok {
			intvalue, err := strconv.Atoi(str)
			if err == nil {
				return &intvalue
			}
		}
	}
	return nil
}

func infoString(key string, info map[string]interface{}) *string {
	if value, found := info[key]; found {
		if str, ok := value.(string); ok {
			return &str
		}
	}
	return nil
}

func infoFloat(key string, info map[string]interface{}) *float64 {
	if value, found := info[key]; found {
		if str, ok := value.(string); ok {
			floatvalue, err := strconv.ParseFloat(str, 64)
			if err == nil {
				return &floatvalue
			}
		}
	}
	return nil
}

func infoBool(key string, info map[string]interface{}) *bool {
	if value, found := info[key]; found {
		if b, ok := value.(bool); ok {
			return &b
		}
	}
	return nil
}

func splitMultipleAltInfos(info map[string]interface{}) []map[string]interface{} {
	maps := make([]map[string]interface{}, 0, 2)
	separator := ","

	for key, v := range info {
		if value, ok := v.(string); ok {
			alternatives := strings.Split(value, separator)
			for position, alt := range alternatives {
				maps = insertMapSlice(maps, position, key, alt)
			}
		} else {
			maps = insertMapSlice(maps, 0, key, v)
		}
	}

	return maps
}

func insertMapSlice(maps []map[string]interface{}, position int, key string, alt interface{}) []map[string]interface{} {
	if len(maps) <= position {
		for i := len(maps); i <= position; i++ {
			maps = append(maps, make(map[string]interface{}))
		}
	}
	maps[position][key] = alt
	return maps
}
