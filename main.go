package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/HazelnutParadise/idensyra/idensyra"
	"github.com/HazelnutParadise/insyra"

	_ "embed"

	"sync"

	"runtime"

	"github.com/gorilla/websocket"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

const version = "0.0.6"

var preCode = `package main
`
var endCode = ``

var defaultCode = `import (
	"fmt"
	"log"
	"github.com/HazelnutParadise/insyra/isr"
	"github.com/HazelnutParadise/insyra"
	"github.com/HazelnutParadise/insyra/datafetch"
	"github.com/HazelnutParadise/insyra/stats"
	"github.com/HazelnutParadise/insyra/parallel"
	"github.com/HazelnutParadise/insyra/plot"
	"github.com/HazelnutParadise/insyra/gplot"
	"github.com/HazelnutParadise/insyra/lpgen"
	"github.com/HazelnutParadise/insyra/csvxl"
	"github.com/HazelnutParadise/insyra/py"

	// No lp package support
	// No other third party package support
)

func main() {
	fmt.Println("Hello, World!")
	log.Println("this is a log message")
	dl := insyra.NewDataList(1, 2, 3)
	fmt.Println("This is your data list:", dl.Data())
}`

// 初始化區塊，啟動程式時會自動執行
func init() {
	fmt.Println("starting Idensyra editor...")
	// 這裡可以進行多初始化操作
}

var fyneApp *fyne.App
var fyneWindow *fyne.Window

var webuiInputCode string
var guiInputCode string
var webuiAlive bool = false
var webuiOpened bool = false

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var clients = make(map[*websocket.Conn]bool)
var clientsMu sync.Mutex
var webSocketConn *websocket.Conn

func main() {
	defer func() {
		if webSocketConn != nil {
			if err := webSocketConn.WriteMessage(websocket.TextMessage, []byte("closeWebUI")); err != nil {
				log.Println("WebSocket write error:", err)
			}
		}
		webSocketConn.Close()
	}()
	var liveRun bool = false
	myApp := app.New()
	fyneApp = &myApp
	myWindow := myApp.NewWindow("Idensyra")
	fyneWindow = &myWindow

	// 檢查是否是通過 go run 執行
	execPath, err := os.Executable()
	execDir := filepath.Dir(execPath)
	if err != nil {
		log.Fatalf("無法獲取執行文件路徑: %v", err)
	}
	isGoRun := strings.HasPrefix(execPath, os.TempDir())

	// 如果是通過 go run 執行，則需要切換工作目錄
	if isGoRun {
		// 獲取主程序的目錄
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			log.Fatalf("無法獲取當前文件路徑")
		}
		dir := filepath.Dir(filename)

		// 切換到主程序所在的目錄
		err = os.Chdir(dir)
		if err != nil {
			log.Fatalf("無法切換到主程序目錄: %v", err)
		}
	} else {
		err = os.Chdir(execDir)
		if err != nil {
			log.Fatalf("無法切換到主程序目錄: %v", err)
		}
	}

	// 建立一個資訊標籤
	infoLabel := widget.NewLabel(fmt.Sprintf("Idensyra v%s, with Insyra v%s", version, insyra.Version))

	liveRunCheck := widget.NewCheck("Live Run on Edit", func(checked bool) {
		liveRun = checked // 更新 liveRun 的值
	})

	// 建立一個多行的 widget.Entry 作為編輯器
	codeInput := widget.NewMultiLineEntry()
	codeInput.SetPlaceHolder("// input Go code here...")
	// 預設輸入
	codeInput.SetText(defaultCode)
	go func() {
		for {
			if webuiInputCode != "" {
				codeInput.SetText(webuiInputCode)
				webuiInputCode = ""
			}
			guiInputCode = codeInput.Text
		}
	}()

	webUIModeButton := widget.NewButton("Switch to Web UI", func() {
		// 切換到 Web UI 模式
		fmt.Println("Switching to Web UI mode...")
		webuiOpened = true
		// 啟動伺服器
		port := startServer()
		myWindow.Hide()
		// 在這裡添加切換到 Web UI 的具體實現
		fyne.CurrentApp().OpenURL(&url.URL{Scheme: "http", Host: fmt.Sprintf("localhost:%d", port)})
	})

	gitHubButton := widget.NewButton("View on GitHub", func() {
		// 打開瀏覽器前往 GitHub 頁面
		fyne.CurrentApp().OpenURL(&url.URL{Scheme: "https", Host: "github.com", Path: "/HazelnutParadise/idensyra"})
	})

	hazelnutParadiseButton := widget.NewButton("HazelnutParadise", func() {
		// 打開瀏覽器前往 HazelnutParadise 頁面
		fyne.CurrentApp().OpenURL(&url.URL{Scheme: "https", Host: "hazelnut-paradise.com"})
	})

	// 建一個用於顯示結果的 Label，並包裹在 Scroll 容器中
	resultBinding := binding.NewString()
	resultLabel := widget.NewLabelWithData(resultBinding)
	resultLabel.Wrapping = fyne.TextWrapWord
	resultLabel.TextStyle = fyne.TextStyle{Monospace: true}

	scrollResult := container.NewScroll(resultLabel)

	// 建立複製按鈕
	copyButton := widget.NewButton("Copy Result", func() {
		result, _ := resultBinding.Get()
		if result == "" {
			dialog.ShowInformation("Copy Failed", "No content to copy.", myWindow)
			return
		}
		myWindow.Clipboard().SetContent(result)
		dialog.ShowInformation("Copy Success", "The result has been copied to the clipboard.", myWindow)
	})

	saveResultButton := widget.NewButton("Save Result", func() {
		result, _ := resultBinding.Get()
		dialog.ShowFileSave(func(uc fyne.URIWriteCloser, err error) {
			if err == nil {
				uc.Write([]byte(result))
				dialog.ShowInformation("Save Success", fmt.Sprintf("The result has been saved as %s", uc.URI().Path()), myWindow)
			}
		}, myWindow)
	})

	// 建立執行按鈕，點擊後執行程式碼
	runButton := widget.NewButton("Run Code", func() {
		code := codeInput.Text        // 獲取使用者輸入的程式碼
		result := executeGoCode(code) // 使用 yaegi 執行程式碼
		resultBinding.Set(result)     // 更新顯示結果
	})

	saveCodeButton := widget.NewButton("Save Code", func() {
		code := preCode + "\n" + codeInput.Text + "\n" + endCode
		dialog.ShowFileSave(func(uc fyne.URIWriteCloser, err error) {
			if err == nil {
				uc.Write([]byte(code))
				dialog.ShowInformation("Save Success", fmt.Sprintf("Your code has been saved as %s", uc.URI().Path()), myWindow)
			}
		}, myWindow)
	})

	// 將執行按鈕和複製按鈕放在一個 VBox 中
	buttonBoxLeft := container.NewVBox(
		runButton,
		saveCodeButton,
	)

	buttonBoxRight := container.NewVBox(
		copyButton,
		saveResultButton,
	)

	// 為文字輸入框和結果輸出框添加標籤
	codeInputLabel := widget.NewLabel("Code Input:")
	resultOutputLabel := widget.NewLabel("Result:")

	// 將標籤和對應的顯示容器組合在一起
	codeInputWithLabel := container.NewBorder(codeInputLabel, nil, nil, nil, codeInput)
	resultOutputWithLabel := container.NewBorder(resultOutputLabel, nil, nil, nil, scrollResult)

	// 建立水平分割的區塊，左邊是編輯器，右邊是結果顯示區
	split := container.NewHSplit(
		container.NewBorder(nil, buttonBoxLeft, nil, nil, codeInputWithLabel),
		container.NewBorder(nil, buttonBoxRight, nil, nil, resultOutputWithLabel),
	)

	// 設置初始分割比例，左邊占比較多
	split.SetOffset(0.55)

	// 將資訊標籤放在頂部，並加入分割區域
	content := container.NewBorder(container.NewGridWithColumns(5, infoLabel, liveRunCheck, webUIModeButton, gitHubButton, hazelnutParadiseButton), nil, nil, nil, split)

	go func() {
		firstRun := true
		oldCode := codeInput.Text
		for {
			if liveRun && (oldCode != codeInput.Text || firstRun) {
				firstRun = false
				code := codeInput.Text        // 獲取使用者輸入的程式碼
				result := executeGoCode(code) // 使用 yaegi 執行程式碼
				resultBinding.Set(result)     // 更新顯示結果
				oldCode = codeInput.Text
			} else if !liveRun {
				firstRun = true
			}
		}
	}()

	go func() {
		for {
			time.Sleep(2 * time.Second)
			if !webuiAlive && webuiOpened {
				myWindow.Show()
				webuiOpened = false
			} else if webuiAlive {
				myWindow.Hide()
				webuiOpened = true
			}
		}
	}()

	// 添加 WebSocket 路由
	http.HandleFunc("/ws", handleWebSocket)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(1200, 650))
	myWindow.ShowAndRun()
}

// 使用 yaegi 來執行動態 Go 程式碼並捕獲所有輸出
func executeGoCode(code string) string {
	// 準備一個 bytes.Buffer 來捕獲所有輸出
	var buf bytes.Buffer

	// 初始化 yaegi 直譯器並設置 Stdout 和 Stderr
	i := interp.New(interp.Options{
		Stdout: &buf,
		Stderr: &buf,
	})
	i.Use(stdlib.Symbols)   // 加載標準庫
	i.Use(idensyra.Symbols) // 加載 idensyra 套件

	// 重定向標準輸出和標準錯誤
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	// 創建一個通道來接收輸出
	outputChan := make(chan string)
	go func() {
		var outputBuf bytes.Buffer
		io.Copy(&outputBuf, r)
		outputChan <- outputBuf.String()
	}()

	// 執行代碼
	_, err := i.Eval(code)
	if err != nil {
		return fmt.Sprintf("執行代碼失敗: %v", err)
	}

	// 恢復標準輸出和標準錯誤
	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// 獲取所有輸出
	output := <-outputChan

	// 合併 yaegi 捕獲的輸出和重定向捕獲的輸出
	return buf.String() + output
}

// ===================== Web UI mode =============================

var runningPort int

// go embed 嵌入靜態文件

//go:embed webui/index.tmpl
var indexHTML string

//go:embed webui/insyra.png
var insyraLogo []byte

// startServer 啟動 Web UI 伺服器，使用可用的端口
func startServer() int {
	if runningPort != 0 {
		return runningPort
	}

	symbols := make([]string, 0)

	for packageFullName, symbol := range idensyra.Symbols {
		packageName := strings.Split(packageFullName, "/")[len(strings.Split(packageFullName, "/"))-1]
		for funcName, _ := range symbol {
			if funcName != "init" && funcName != "main" && !strings.HasPrefix(funcName, "_") {
				symbols = append(symbols, packageName+"."+funcName)
			}
		}
	}

	for packageFullName, symbol := range stdlib.Symbols {
		packageName := strings.Split(packageFullName, "/")[len(strings.Split(packageFullName, "/"))-1]
		for funcName, _ := range symbol {
			if funcName != "init" && funcName != "main" && !strings.HasPrefix(funcName, "_") {
				symbols = append(symbols, packageName+"."+funcName)
			}
		}
	}

	go func() {
		port, err := findAvailablePort()
		if err != nil {
			fmt.Println("Error finding available port:", err)
			return
		}
		fmt.Printf("Starting Web UI on port %d\n", port)
		runningPort = port

		// 設置 HTTP 處理程序
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// 用模板引擎渲染 index.html
			tmpl, err := template.New("index").Parse(indexHTML)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			data := struct {
				Version     string
				InsyaLogo   string
				Port        int
				DefaultCode string
				PreCode     string
				EndCode     string
				Symbols     []string
			}{
				Version:     version,
				InsyaLogo:   base64.StdEncoding.EncodeToString(insyraLogo),
				Port:        runningPort,
				DefaultCode: guiInputCode,
				PreCode:     preCode,
				EndCode:     endCode,
				Symbols:     symbols,
			}

			err = tmpl.Execute(w, data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})

		http.HandleFunc("/api/execute", executeCodeHandler)
		http.HandleFunc("/api/backToGui", backToGuiHandler)
		http.HandleFunc("/api/syncCode", WebUICodeSyncHandler)
		http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	}()
	for {
		if runningPort != 0 {
			return runningPort
		}
	}
}

// findAvailablePort 尋找可用的端口
func findAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", ":0") // 0 代表自動選擇可用端口
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

// api 函數
func executeCodeHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Code string `json:"codeInput"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	decodedCode, err := url.QueryUnescape(requestBody.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := executeGoCode(decodedCode)
	w.Write([]byte(result))
}

func backToGuiHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Code string `json:"codeInput"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	decodedCode, err := url.QueryUnescape(requestBody.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	webuiInputCode = decodedCode
	myWindow := *fyneWindow
	myWindow.Show()
}

func WebUICodeSyncHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Code string `json:"codeInput"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	decodedCode, err := url.QueryUnescape(requestBody.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	webuiInputCode = decodedCode
}

// 添加 WebSocket 處理函數
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	webSocketConn = conn
	webuiAlive = true

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	// 添加心跳
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("心跳發送失敗:", err)
				return
			}
		}
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}
	}

	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()
	webuiAlive = false
}
