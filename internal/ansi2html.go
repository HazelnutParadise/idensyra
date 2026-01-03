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
	// 追蹤當前打開的 span 標籤數量
	openSpanCount := 0

	// 匹配 ANSI 轉義序列
	re := regexp.MustCompile(`\x1b\[([0-9;]+)m`)

	var result strings.Builder
	lastIndex := 0

	matches := re.FindAllStringSubmatchIndex(input, -1)
	for _, match := range matches {
		// 添加匹配之前的文字
		result.WriteString(input[lastIndex:match[0]])

		// 獲取 ANSI 代碼
		codes := input[match[2]:match[3]]

		// 處理多個代碼（用分號分隔）
		for _, code := range strings.Split(codes, ";") {
			switch code {
			case "0": // 重置 - 關閉所有打開的 span
				for i := 0; i < openSpanCount; i++ {
					result.WriteString("</span>")
				}
				openSpanCount = 0
			case "1": // bold
				result.WriteString("<span class='ansi-bold'>")
				openSpanCount++
			case "2": // dim
				result.WriteString("<span class='ansi-dim'>")
				openSpanCount++
			case "3": // italic
				result.WriteString("<span class='ansi-italic'>")
				openSpanCount++
			case "4": // underline
				result.WriteString("<span class='ansi-underline'>")
				openSpanCount++
			case "22": // 關閉 bold/dim
				if openSpanCount > 0 {
					result.WriteString("</span>")
					openSpanCount--
				}
			case "23": // 關閉 italic
				if openSpanCount > 0 {
					result.WriteString("</span>")
					openSpanCount--
				}
			case "24": // 關閉 underline
				if openSpanCount > 0 {
					result.WriteString("</span>")
					openSpanCount--
				}
			case "30":
				result.WriteString("<span class='ansi-fg-30'>")
				openSpanCount++
			case "31":
				result.WriteString("<span class='ansi-fg-31'>")
				openSpanCount++
			case "32":
				result.WriteString("<span class='ansi-fg-32'>")
				openSpanCount++
			case "33":
				result.WriteString("<span class='ansi-fg-33'>")
				openSpanCount++
			case "34":
				result.WriteString("<span class='ansi-fg-34'>")
				openSpanCount++
			case "35":
				result.WriteString("<span class='ansi-fg-35'>")
				openSpanCount++
			case "36":
				result.WriteString("<span class='ansi-fg-36'>")
				openSpanCount++
			case "37":
				result.WriteString("<span class='ansi-fg-37'>")
				openSpanCount++
			case "39": // 重置前景色
				if openSpanCount > 0 {
					result.WriteString("</span>")
					openSpanCount--
				}
			case "40":
				result.WriteString("<span class='ansi-bg-40'>")
				openSpanCount++
			case "41":
				result.WriteString("<span class='ansi-bg-41'>")
				openSpanCount++
			case "42":
				result.WriteString("<span class='ansi-bg-42'>")
				openSpanCount++
			case "43":
				result.WriteString("<span class='ansi-bg-43'>")
				openSpanCount++
			case "44":
				result.WriteString("<span class='ansi-bg-44'>")
				openSpanCount++
			case "45":
				result.WriteString("<span class='ansi-bg-45'>")
				openSpanCount++
			case "46":
				result.WriteString("<span class='ansi-bg-46'>")
				openSpanCount++
			case "47":
				result.WriteString("<span class='ansi-bg-47'>")
				openSpanCount++
			case "49": // 重置背景色
				if openSpanCount > 0 {
					result.WriteString("</span>")
					openSpanCount--
				}
			case "90":
				result.WriteString("<span class='ansi-fg-90'>")
				openSpanCount++
			case "91":
				result.WriteString("<span class='ansi-fg-91'>")
				openSpanCount++
			case "92":
				result.WriteString("<span class='ansi-fg-92'>")
				openSpanCount++
			case "93":
				result.WriteString("<span class='ansi-fg-93'>")
				openSpanCount++
			case "94":
				result.WriteString("<span class='ansi-fg-94'>")
				openSpanCount++
			case "95":
				result.WriteString("<span class='ansi-fg-95'>")
				openSpanCount++
			case "96":
				result.WriteString("<span class='ansi-fg-96'>")
				openSpanCount++
			case "97":
				result.WriteString("<span class='ansi-fg-97'>")
				openSpanCount++
			case "100":
				result.WriteString("<span class='ansi-bg-100'>")
				openSpanCount++
			case "101":
				result.WriteString("<span class='ansi-bg-101'>")
				openSpanCount++
			case "102":
				result.WriteString("<span class='ansi-bg-102'>")
				openSpanCount++
			case "103":
				result.WriteString("<span class='ansi-bg-103'>")
				openSpanCount++
			case "104":
				result.WriteString("<span class='ansi-bg-104'>")
				openSpanCount++
			case "105":
				result.WriteString("<span class='ansi-bg-105'>")
				openSpanCount++
			case "106":
				result.WriteString("<span class='ansi-bg-106'>")
				openSpanCount++
			case "107":
				result.WriteString("<span class='ansi-bg-107'>")
				openSpanCount++
			}
		}

		lastIndex = match[1]
	}

	// 添加剩餘的文字
	result.WriteString(input[lastIndex:])

	// 確保所有開啟的 span 標籤都有對應的關閉標籤
	for i := 0; i < openSpanCount; i++ {
		result.WriteString("</span>")
	}

	return result.String()
}
