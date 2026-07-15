package main

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type CrashMetadataResolver struct {
	sectionNames  map[int]string
	pageNames     map[int]string
	contextNames  map[int]string
	context2Names map[int]string
}

func NewCrashMetadataResolver() *CrashMetadataResolver {
	cmr := new(CrashMetadataResolver)

	cmr.sectionNames = parseEnum(pulsarIdentifiers, "SectionId")
	cmr.pageNames = parseEnum(pulsarIdentifiers, "PageId")
	pulPageNames := parseEnum(pulsarUI, "PulPageId")
	for k, v := range pulPageNames {
		cmr.pageNames[k] = v
	}

	cmr.contextNames = parseEnum(pulsarSystemHpp, "Context")
	cmr.context2Names = parseEnum(pulsarSystemHpp, "Context2")

	return cmr
}

func (cmr *CrashMetadataResolver) getSectionName(id int) string {
	return cmr.sectionNames[id]
}

func (cmr *CrashMetadataResolver) getPageName(id int) string {
	return cmr.pageNames[id]
}

func (cmr *CrashMetadataResolver) getEnabledContexts(context uint, context2 uint) string {
	enabled := []string{}
	appendEnabledContexts(&enabled, cmr.contextNames, context)
	appendEnabledContexts(&enabled, cmr.context2Names, context2)

	if len(enabled) == 0 {
		return "None"
	}

	return strings.Join(enabled, ", ")
}

func appendEnabledContexts(enabled *[]string, names map[int]string, enabledBits uint) {
	for k, v := range names {
		if k < 0 || k >= 32 {
			continue
		}

		if (enabledBits & (1 << k)) != 0 {
			*enabled = append(*enabled, v)
		}
	}

	slices.Sort(*enabled)
}

func parseEnum(resource string, enumName string) map[int]string {
	values := map[int]string{}

	if len(resource) == 0 {
		return values
	}

	patternStr := fmt.Sprintf(`enum\s+%s\s*{(?<body>[ _=,A-Za-z0-9\n/]*)};`, enumName)
	pattern := regexp.MustCompile(patternStr)
	matches := pattern.FindStringSubmatch(resource)
	matchNames := pattern.SubexpNames()

	if len(matches) == 0 {
		return values
	}

	bodyIdx := slices.Index(matchNames, "body")
	bodyText := matches[bodyIdx]
	bodyPattern := regexp.MustCompile("//.*")
	body := bodyPattern.ReplaceAllString(bodyText, "")
	entries := strings.Split(body, ",")
	nextValue := 0

	for _, rawEntry := range entries {
		entry := strings.TrimSpace(rawEntry)
		if len(entry) == 0 {
			continue
		}

		parts := strings.SplitN(entry, "=", 2)
		name := strings.TrimSpace(parts[0])
		if len(name) == 0 {
			continue
		}

		value := nextValue

		if len(parts) == 2 {
			value64, err := tryParseHex(strings.TrimSpace(parts[1]))

			if err != nil {
				continue
			}

			value = int(value64)
		}

		values[value] = name
		nextValue = value + 1
	}

	return values
}

func tryParseHex(value string) (int64, error) {
	value = strings.TrimSpace(value)

	if strings.HasPrefix(value, "0x") {
		return strconv.ParseInt(value[2:], 16, 32)
	}

	if strings.HasPrefix(value, "-0x") {
		num, err := strconv.ParseInt(value[3:], 16, 32)
		return -num, err
	}

	return strconv.ParseInt(value, 16, 32)
}
