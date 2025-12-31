package internal

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

// AnsiToHTMLWithBG 支援亮/深背景，將 ANSI 轉換為 HTML
func AnsiToHTMLWithBG(input string, bg string) string {
	ansiToHTML := map[string]string{
		// 重置
		"0": "</span>",

		// 樣式
		"1": "<span class='ansi-bold'>",
		"2": "<span class='ansi-dim'>",
		"3": "<span class='ansi-italic'>",
		"4": "<span class='ansi-underline'>",

		// 前景色 (標準色)
		"30": "<span class='ansi-fg-30'>",
		"31": "<span class='ansi-fg-31'>",
		"32": "<span class='ansi-fg-32'>",
		"33": "<span class='ansi-fg-33'>",
		"34": "<span class='ansi-fg-34'>",
		"35": "<span class='ansi-fg-35'>",
		"36": "<span class='ansi-fg-36'>",
		"37": "<span class='ansi-fg-37'>",

		// 前景色 (亮色)
		"90": "<span class='ansi-fg-90'>",
		"91": "<span class='ansi-fg-91'>",
		"92": "<span class='ansi-fg-92'>",
		"93": "<span class='ansi-fg-93'>",
		"94": "<span class='ansi-fg-94'>",
		"95": "<span class='ansi-fg-95'>",
		"96": "<span class='ansi-fg-96'>",
		"97": "<span class='ansi-fg-97'>",

		// 背景色 (標準色)
		"40": "<span class='ansi-bg-40'>",
		"41": "<span class='ansi-bg-41'>",
		"42": "<span class='ansi-bg-42'>",
		"43": "<span class='ansi-bg-43'>",
		"44": "<span class='ansi-bg-44'>",
		"45": "<span class='ansi-bg-45'>",
		"46": "<span class='ansi-bg-46'>",
		"47": "<span class='ansi-bg-47'>",

		// 背景色 (亮色)
		"100": "<span class='ansi-bg-100'>",
		"101": "<span class='ansi-bg-101'>",
		"102": "<span class='ansi-bg-102'>",
		"103": "<span class='ansi-bg-103'>",
		"104": "<span class='ansi-bg-104'>",
		"105": "<span class='ansi-bg-105'>",
		"106": "<span class='ansi-bg-106'>",
		"107": "<span class='ansi-bg-107'>",
	}

	// 匹配 ANSI 轉義序列
	re := regexp.MustCompile(`\x1b\[([0-9;]+)m`)

	output := re.ReplaceAllStringFunc(input, func(match string) string {
		codes := re.FindStringSubmatch(match)[1]
		var html string

		// 處理多個代碼（用分號分隔）
		for _, code := range strings.Split(codes, ";") {
			if v, ok := ansiToHTML[code]; ok {
				html += v
			}
		}

		return html
	})

	// 確保所有開啟的 span 標籤都有對應的關閉標籤
	openSpans := strings.Count(output, "<span")
	closeSpans := strings.Count(output, "</span>")
	if openSpans > closeSpans {
		output += strings.Repeat("</span>", openSpans-closeSpans)
	}

	return output
}
