package engine

import "strconv"

// FormatINR renders a number using the Indian digit-grouping convention
// (last 3 digits, then groups of 2), e.g. 1100000 -> "11,00,000". Exported
// for use by the web layer's templates.
func FormatINR(n float64) string { return formatINR(n) }

// formatINR renders a number using the Indian digit-grouping convention
// (last 3 digits, then groups of 2), e.g. 1100000 -> "11,00,000".
func formatINR(n float64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	whole := int64(n + 0.5)
	s := strconv.FormatInt(whole, 10)
	if len(s) <= 3 {
		if neg {
			return "-" + s
		}
		return s
	}
	last3 := s[len(s)-3:]
	rest := s[:len(s)-3]
	grouped := ""
	for len(rest) > 2 {
		grouped = "," + rest[len(rest)-2:] + grouped
		rest = rest[:len(rest)-2]
	}
	grouped = rest + grouped
	result := grouped + "," + last3
	if neg {
		result = "-" + result
	}
	return result
}

func employmentLabel(e Employment) string {
	switch e {
	case EmploymentSalaried:
		return "salaried"
	case EmploymentSelfEmployed:
		return "self-employed"
	default:
		return "informal-income"
	}
}
