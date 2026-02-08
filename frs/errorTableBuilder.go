package main

import (
	"fmt"
	"sort"
	"strings"
)

func buildErrorTable(errors map[string]string) string {
	if len(errors) == 0 {
		return "No errors encountered."
	}

	var b strings.Builder

	b.WriteString("This page keeps track of the latest errors that SodiumBot during running the [[WP:FRS|Feedback request service]] job.\n\n")
	b.WriteString("{| class=\"wikitable sortable\"\n")
	b.WriteString("! Page\n")
	b.WriteString("! Error\n")

	// stable ordering for clean diffs
	keys := make([]string, 0, len(errors))
	for k := range errors {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, page := range keys {
		err := errors[page]
		b.WriteString("|-\n")
		b.WriteString(fmt.Sprintf("| [[%s]]\n", page))
		b.WriteString(fmt.Sprintf("| <code><nowiki>%s</nowiki></code>\n", err))
	}

	b.WriteString("|}")

	return b.String()
}
