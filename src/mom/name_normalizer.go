package mom

import "strings"

func normalizeMappingName(namespace, input string) string {
	normalizer, exists := normalizerMappings[namespace]
	if !exists || normalizer == nil {
		normalizer = defaultNormalizer
	}
	return normalizer(input)
}

/*
INameNormalizer normalizes name for mapping.
*/
type INameNormalizer func(input string) string

var normalizerMappings = map[string]INameNormalizer{
	"*":     defaultNormalizer,
	"email": emailNormalizer,
}

/*
defaultNormalizer trims leading and trailing spaces off input.
*/
func defaultNormalizer(input string) string {
	return strings.TrimSpace(input)
}

/*
emailNormalizer is used to normalize email address.
*/
func emailNormalizer(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
