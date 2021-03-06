package mom

import (
	"regexp"
	"strings"
)

func normalizeNamespace(namespace string) string {
	return strings.TrimSpace(strings.ToLower(namespace))
}

func normalizeMappingTarget(to string) string {
	return strings.TrimSpace(to)
}

func normalizeMappingObject(namespace, to string) string {
	namespace = normalizeNamespace(namespace)
	normalizer, exists := normalizerMappings[namespace]
	if !exists || normalizer == nil {
		normalizer = defaultNormalizer
	}
	return normalizer(to)
}

/*
INameNormalizer normalizes name for mapping.
*/
type INameNormalizer func(input string) string

var normalizerMappings = map[string]INameNormalizer{
	"*":             defaultNormalizer,
	"email":         emailNormalizer,
	"email_addr":    emailNormalizer,
	"email_address": emailNormalizer,
	"phone":         phoneNormalizer,
	"phone_num":     phoneNormalizer,
	"mobile":        phoneNormalizer,
	"mobile_num":    phoneNormalizer,
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

var regexpNonDigit = regexp.MustCompile(`[^\d]+`)
var regexpStartWithZeroes = regexp.MustCompile(`^0+`)

/*
phoneNormalizer is used to normalize phone number.
*/
func phoneNormalizer(phone string) string {
	return regexpStartWithZeroes.ReplaceAllString(regexpNonDigit.ReplaceAllString(phone, ""), "")
}
