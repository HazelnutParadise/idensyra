## igonb

2026/01/03 01:53:23 nil type
ERR | process message error: C{"name":"main.App.ExecuteIgonbCells","args":["{\n \"version\": 1,\n \"cells\": [\n {\n \"id\": \"igonb-1\",\n \"language\": \"go\",\n \"source\": \"for range 10 {\\r\\n\\t\\\"hi\\\"\\r\\n}\",\n \"output\": \"\",\n \"error\": \"1:28: invalid operation: mismatched types untyped string and untyped int\"\n },\n {\n \"id\": \"igonb-2\",\n \"language\": \"python\",\n \"source\": \"\\\"hi\\\"*10\",\n \"output\": \"hihihihihihihihihihi\\n\",\n \"error\": \"\"\n }\n ]\n}",-2],"callbackID":"main.App.ExecuteIgonbCells-4171985527"} -> nil type
ERR | nil type

這種錯誤應該讓儲存格執行錯誤 而不是卡在執行狀態
另外要提供一個強制停止執行的功能
