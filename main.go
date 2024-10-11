package main

import (
	"bytes"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Idensyra")

	// 建立一個多行的 widget.Entry 作為編輯器
	codeInput := widget.NewMultiLineEntry()
	codeInput.SetPlaceHolder("// 在這裡輸入 Go 程式碼...")
	// 預設輸入
	codeInput.SetText(`import (
	"fmt"
	"log"
)
func main() {
	fmt.Println("Hello, World!")
	log.Println("這是一個 log 訊息")
}`)

	// 建立一個用於顯示結果的 Label，並包裹在 Scroll 容器中
	resultBinding := binding.NewString()
	resultLabel := widget.NewLabelWithData(resultBinding)
	resultLabel.Wrapping = fyne.TextWrapWord
	resultLabel.TextStyle = fyne.TextStyle{Monospace: true}

	scrollResult := container.NewScroll(resultLabel)

	// 建立複製按鈕
	copyButton := widget.NewButton("複製結果", func() {
		result, _ := resultBinding.Get()
		if result == "" {
			dialog.ShowInformation("複製失敗", "沒有內容可供複製。", myWindow)
			return
		}
		myWindow.Clipboard().SetContent(result) // 修正此行
		dialog.ShowInformation("複製成功", "執行結果已複製到剪貼簿。", myWindow)
	})

	// 建立執行按鈕，點擊後執行程式碼
	runButton := widget.NewButton("執行程式碼", func() {
		code := codeInput.Text        // 獲取使用者輸入的程式碼
		result := executeGoCode(code) // 使用 yaegi 執行程式碼
		resultBinding.Set(result)     // 更新顯示結果
	})

	// 將執行按鈕和複製按鈕放在一個 VBox 中
	buttonBox := container.NewVBox(
		runButton,
		copyButton,
	)

	// 為文字輸入框和結果輸出框添加標籤
	codeInputLabel := widget.NewLabel("程式碼輸入:")
	resultOutputLabel := widget.NewLabel("執行結果:")

	// 將標籤和對應的顯示容器組合在一起
	codeInputWithLabel := container.NewBorder(codeInputLabel, nil, nil, nil, codeInput)
	resultOutputWithLabel := container.NewBorder(resultOutputLabel, nil, nil, nil, scrollResult)

	// 建立水平分割的區域，左邊是編輯器，右邊是結果顯示區
	split := container.NewHSplit(
		container.NewBorder(nil, buttonBox, nil, nil, codeInputWithLabel),
		resultOutputWithLabel,
	)

	// 設置初始分割比例，左邊占比較多
	split.SetOffset(0.55)

	myWindow.SetContent(split)
	myWindow.Resize(fyne.NewSize(1200, 650))
	myWindow.ShowAndRun()
}

// 使用 yaegi 來執行動態 Go 程式碼並捕獲標準輸出與 log 輸出
func executeGoCode(code string) string {
	// 構建完整的 Go 程式碼
	preCode := `
package main
`
	endCode := `
`

	code = preCode + code + endCode

	// 準備一個 bytes.Buffer 來捕獲標準輸出和 log 輸出
	var buf bytes.Buffer

	// 初始化 yaegi 直譯器並設置 Stdout 和 Stderr
	i := interp.New(interp.Options{
		Stdout: &buf,
		Stderr: &buf,
	})
	i.Use(stdlib.Symbols) // 加載標準庫

	// 執行傳入的 Go 程式碼
	_, err := i.Eval(code)

	if err != nil {
		return fmt.Sprintf("執行錯誤: %v", err)
	}

	// 返回執行過程中的所有標準輸出與 log 輸出
	return buf.String()
}
