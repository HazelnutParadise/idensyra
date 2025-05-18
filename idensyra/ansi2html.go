package idensyra

import (
	"regexp"
	"strings"
)

func AnsiToHTML(input string) string {
	return AnsiToHTMLWithBG(input, "dark")
}

// AnsiToPlain 移除所有 ANSI 標籤
func AnsiToPlain(input string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]+m`)
	return re.ReplaceAllString(input, "")
}

// AnsiToHTMLWithBG 支援亮/深背景
func AnsiToHTMLWithBG(input string, bg string) string {
	ansiToHTML := map[string]string{
		"0":  "</span>",
		"1":  "<span class='ansi-bold'>",
		"3":  "<span class='ansi-italic'>",
		"4":  "<span class='ansi-underline'>",
		"30": "<span class='ansi-fg-30'>",
		"31": "<span class='ansi-fg-31'>",
		"32": "<span class='ansi-fg-32'>",
		"33": "<span class='ansi-fg-33'>",
		"34": "<span class='ansi-fg-34'>",
		"35": "<span class='ansi-fg-35'>",
		"36": "<span class='ansi-fg-36'>",
		"37": "<span class='ansi-fg-37'>",
		"90": "<span class='ansi-fg-90'>",
		"91": "<span class='ansi-fg-91'>",
		"92": "<span class='ansi-fg-92'>",
		"93": "<span class='ansi-fg-93'>",
		"94": "<span class='ansi-fg-94'>",
		"95": "<span class='ansi-fg-95'>",
		"96": "<span class='ansi-fg-96'>",
		"97": "<span class='ansi-fg-97'>",
	}
	re := regexp.MustCompile(`\x1b\[([0-9;]+)m`)
	output := re.ReplaceAllStringFunc(input, func(match string) string {
		codes := re.FindStringSubmatch(match)[1]
		var html string
		for _, code := range strings.Split(codes, ";") {
			if v, ok := ansiToHTML[code]; ok {
				html += v
			}
		}
		return html
	})
	output += strings.Repeat("</span>", strings.Count(output, "<span")-strings.Count(output, "</span>"))
	return output
}
