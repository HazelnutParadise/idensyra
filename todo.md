1. [x] 支援執行python檔案
       利用py.RunFile(nil, "path/to/your/python/file.py")函式
       原理是檔使用者要執行一個python文件 自動生成py.RunFile(nil, "path/to/your/python/file.py")的go代碼並丟給yaegi執行

2. [x] 開發 igonb 格式，類似ipynb，可以在裡面執行go或python程式碼，import可以出現在各處(只作用在他出現之後)
3. [x] 將 igonb 模組獨立分離出來，之後可移到獨立repo，讓其他人可以import就支援igonb檔
