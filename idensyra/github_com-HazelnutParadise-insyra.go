// Code generated by 'yaegi extract github.com/HazelnutParadise/insyra'. DO NOT EDIT.

package idensyra

import (
	"github.com/HazelnutParadise/insyra"
	"go/constant"
	"go/token"
	"reflect"
)

func init() {
	Symbols["github.com/HazelnutParadise/insyra/insyra"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Config":                reflect.ValueOf(&insyra.Config).Elem(),
		"ConvertLongDataToWide": reflect.ValueOf(insyra.ConvertLongDataToWide),
		"LogDebug":              reflect.ValueOf(insyra.LogDebug),
		"LogFatal":              reflect.ValueOf(insyra.LogFatal),
		"LogInfo":               reflect.ValueOf(insyra.LogInfo),
		"LogLevelDebug":         reflect.ValueOf(insyra.LogLevelDebug),
		"LogLevelFatal":         reflect.ValueOf(insyra.LogLevelFatal),
		"LogLevelInfo":          reflect.ValueOf(insyra.LogLevelInfo),
		"LogLevelWarning":       reflect.ValueOf(insyra.LogLevelWarning),
		"LogWarning":            reflect.ValueOf(insyra.LogWarning),
		"NewDataList":           reflect.ValueOf(insyra.NewDataList),
		"NewDataTable":          reflect.ValueOf(insyra.NewDataTable),
		"PowRat":                reflect.ValueOf(insyra.PowRat),
		"ProcessData":           reflect.ValueOf(insyra.ProcessData),
		"SetDefaultConfig":      reflect.ValueOf(insyra.SetDefaultConfig),
		"SliceToF64":            reflect.ValueOf(insyra.SliceToF64),
		"SqrtRat":               reflect.ValueOf(insyra.SqrtRat),
		"ToFloat64":             reflect.ValueOf(insyra.ToFloat64),
		"ToFloat64Safe":         reflect.ValueOf(insyra.ToFloat64Safe),
		"Version":               reflect.ValueOf(constant.MakeFromLiteral("\"0.0.14\"", token.STRING, 0)),

		// type definitions
		"DataList":    reflect.ValueOf((*insyra.DataList)(nil)),
		"DataTable":   reflect.ValueOf((*insyra.DataTable)(nil)),
		"FilterFunc":  reflect.ValueOf((*insyra.FilterFunc)(nil)),
		"IDataList":   reflect.ValueOf((*insyra.IDataList)(nil)),
		"IDataTable":  reflect.ValueOf((*insyra.IDataTable)(nil)),
		"LogLevel":    reflect.ValueOf((*insyra.LogLevel)(nil)),
		"NameManager": reflect.ValueOf((*insyra.NameManager)(nil)),

		// interface wrapper definitions
		"_IDataList":  reflect.ValueOf((*_github_com_HazelnutParadise_insyra_IDataList)(nil)),
		"_IDataTable": reflect.ValueOf((*_github_com_HazelnutParadise_insyra_IDataTable)(nil)),
	}
}

// _github_com_HazelnutParadise_insyra_IDataList is an interface wrapper for IDataList type
type _github_com_HazelnutParadise_insyra_IDataList struct {
	IValue                        interface{}
	WAppend                       func(values ...interface{})
	WCapitalize                   func() *insyra.DataList
	WClear                        func() *insyra.DataList
	WClearNaNs                    func() *insyra.DataList
	WClearNumbers                 func() *insyra.DataList
	WClearOutliers                func(a0 float64) *insyra.DataList
	WClearStrings                 func() *insyra.DataList
	WClone                        func() *insyra.DataList
	WCount                        func(value interface{}) int
	WCounter                      func() map[interface{}]int
	WData                         func() []interface{}
	WDifference                   func() *insyra.DataList
	WDoubleExponentialSmoothing   func(a0 float64, a1 float64) *insyra.DataList
	WDrop                         func(index int) *insyra.DataList
	WDropAll                      func(a0 ...interface{}) *insyra.DataList
	WDropIfContains               func(a0 interface{}) *insyra.DataList
	WExponentialSmoothing         func(a0 float64) *insyra.DataList
	WFillNaNWithMean              func() *insyra.DataList
	WFilter                       func(a0 func(interface{}) bool) *insyra.DataList
	WFindAll                      func(a0 interface{}) []int
	WFindFirst                    func(a0 interface{}) interface{}
	WFindLast                     func(a0 interface{}) interface{}
	WGMean                        func() float64
	WGet                          func(index int) interface{}
	WGetCreationTimestamp         func() int64
	WGetLastModifiedTimestamp     func() int64
	WGetName                      func() string
	WHermiteInterpolation         func(a0 float64, a1 []float64) float64
	WIQR                          func() float64
	WInsertAt                     func(index int, value interface{})
	WIsEqualTo                    func(a0 *insyra.DataList) bool
	WIsTheSameAs                  func(a0 *insyra.DataList) bool
	WLagrangeInterpolation        func(a0 float64) float64
	WLen                          func() int
	WLinearInterpolation          func(a0 float64) float64
	WLower                        func() *insyra.DataList
	WMAD                          func() float64
	WMax                          func() float64
	WMean                         func() float64
	WMedian                       func() float64
	WMin                          func() float64
	WMode                         func() float64
	WMovingAverage                func(a0 int) *insyra.DataList
	WMovingStdev                  func(a0 int) *insyra.DataList
	WNearestNeighborInterpolation func(a0 float64) float64
	WNewtonInterpolation          func(a0 float64) float64
	WNormalize                    func() *insyra.DataList
	WParseNumbers                 func() *insyra.DataList
	WParseStrings                 func() *insyra.DataList
	WPercentile                   func(a0 float64) float64
	WPop                          func() interface{}
	WQuadraticInterpolation       func(a0 float64) float64
	WQuartile                     func(a0 int) float64
	WRange                        func() float64
	WRank                         func() *insyra.DataList
	WReplaceAll                   func(a0 interface{}, a1 interface{})
	WReplaceFirst                 func(a0 interface{}, a1 interface{})
	WReplaceLast                  func(a0 interface{}, a1 interface{})
	WReplaceOutliers              func(a0 float64, a1 float64) *insyra.DataList
	WReverse                      func() *insyra.DataList
	WSetName                      func(a0 string) *insyra.DataList
	WSort                         func(acending ...bool) *insyra.DataList
	WStandardize                  func() *insyra.DataList
	WStdev                        func() float64
	WStdevP                       func() float64
	WSum                          func() float64
	WToF64Slice                   func() []float64
	WToStringSlice                func() []string
	WUpdate                       func(index int, value interface{})
	WUpper                        func() *insyra.DataList
	WVar                          func() float64
	WVarP                         func() float64
	WWeightedMean                 func(weights interface{}) float64
	WWeightedMovingAverage        func(a0 int, a1 interface{}) *insyra.DataList
}

func (W _github_com_HazelnutParadise_insyra_IDataList) Append(values ...interface{}) {
	W.WAppend(values...)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Capitalize() *insyra.DataList {
	return W.WCapitalize()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Clear() *insyra.DataList {
	return W.WClear()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) ClearNaNs() *insyra.DataList {
	return W.WClearNaNs()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) ClearNumbers() *insyra.DataList {
	return W.WClearNumbers()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) ClearOutliers(a0 float64) *insyra.DataList {
	return W.WClearOutliers(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) ClearStrings() *insyra.DataList {
	return W.WClearStrings()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Clone() *insyra.DataList {
	return W.WClone()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Count(value interface{}) int {
	return W.WCount(value)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Counter() map[interface{}]int {
	return W.WCounter()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Data() []interface{} {
	return W.WData()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Difference() *insyra.DataList {
	return W.WDifference()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) DoubleExponentialSmoothing(a0 float64, a1 float64) *insyra.DataList {
	return W.WDoubleExponentialSmoothing(a0, a1)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Drop(index int) *insyra.DataList {
	return W.WDrop(index)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) DropAll(a0 ...interface{}) *insyra.DataList {
	return W.WDropAll(a0...)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) DropIfContains(a0 interface{}) *insyra.DataList {
	return W.WDropIfContains(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) ExponentialSmoothing(a0 float64) *insyra.DataList {
	return W.WExponentialSmoothing(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) FillNaNWithMean() *insyra.DataList {
	return W.WFillNaNWithMean()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Filter(a0 func(interface{}) bool) *insyra.DataList {
	return W.WFilter(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) FindAll(a0 interface{}) []int {
	return W.WFindAll(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) FindFirst(a0 interface{}) interface{} {
	return W.WFindFirst(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) FindLast(a0 interface{}) interface{} {
	return W.WFindLast(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) GMean() float64 {
	return W.WGMean()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Get(index int) interface{} {
	return W.WGet(index)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) GetCreationTimestamp() int64 {
	return W.WGetCreationTimestamp()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) GetLastModifiedTimestamp() int64 {
	return W.WGetLastModifiedTimestamp()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) GetName() string {
	return W.WGetName()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) HermiteInterpolation(a0 float64, a1 []float64) float64 {
	return W.WHermiteInterpolation(a0, a1)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) IQR() float64 {
	return W.WIQR()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) InsertAt(index int, value interface{}) {
	W.WInsertAt(index, value)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) IsEqualTo(a0 *insyra.DataList) bool {
	return W.WIsEqualTo(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) IsTheSameAs(a0 *insyra.DataList) bool {
	return W.WIsTheSameAs(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) LagrangeInterpolation(a0 float64) float64 {
	return W.WLagrangeInterpolation(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Len() int {
	return W.WLen()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) LinearInterpolation(a0 float64) float64 {
	return W.WLinearInterpolation(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Lower() *insyra.DataList {
	return W.WLower()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) MAD() float64 {
	return W.WMAD()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Max() float64 {
	return W.WMax()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Mean() float64 {
	return W.WMean()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Median() float64 {
	return W.WMedian()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Min() float64 {
	return W.WMin()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Mode() float64 {
	return W.WMode()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) MovingAverage(a0 int) *insyra.DataList {
	return W.WMovingAverage(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) MovingStdev(a0 int) *insyra.DataList {
	return W.WMovingStdev(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) NearestNeighborInterpolation(a0 float64) float64 {
	return W.WNearestNeighborInterpolation(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) NewtonInterpolation(a0 float64) float64 {
	return W.WNewtonInterpolation(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Normalize() *insyra.DataList {
	return W.WNormalize()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) ParseNumbers() *insyra.DataList {
	return W.WParseNumbers()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) ParseStrings() *insyra.DataList {
	return W.WParseStrings()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Percentile(a0 float64) float64 {
	return W.WPercentile(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Pop() interface{} {
	return W.WPop()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) QuadraticInterpolation(a0 float64) float64 {
	return W.WQuadraticInterpolation(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Quartile(a0 int) float64 {
	return W.WQuartile(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Range() float64 {
	return W.WRange()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Rank() *insyra.DataList {
	return W.WRank()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) ReplaceAll(a0 interface{}, a1 interface{}) {
	W.WReplaceAll(a0, a1)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) ReplaceFirst(a0 interface{}, a1 interface{}) {
	W.WReplaceFirst(a0, a1)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) ReplaceLast(a0 interface{}, a1 interface{}) {
	W.WReplaceLast(a0, a1)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) ReplaceOutliers(a0 float64, a1 float64) *insyra.DataList {
	return W.WReplaceOutliers(a0, a1)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Reverse() *insyra.DataList {
	return W.WReverse()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) SetName(a0 string) *insyra.DataList {
	return W.WSetName(a0)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Sort(acending ...bool) *insyra.DataList {
	return W.WSort(acending...)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Standardize() *insyra.DataList {
	return W.WStandardize()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Stdev() float64 {
	return W.WStdev()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) StdevP() float64 {
	return W.WStdevP()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Sum() float64 {
	return W.WSum()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) ToF64Slice() []float64 {
	return W.WToF64Slice()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) ToStringSlice() []string {
	return W.WToStringSlice()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Update(index int, value interface{}) {
	W.WUpdate(index, value)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Upper() *insyra.DataList {
	return W.WUpper()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) Var() float64 {
	return W.WVar()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) VarP() float64 {
	return W.WVarP()
}
func (W _github_com_HazelnutParadise_insyra_IDataList) WeightedMean(weights interface{}) float64 {
	return W.WWeightedMean(weights)
}
func (W _github_com_HazelnutParadise_insyra_IDataList) WeightedMovingAverage(a0 int, a1 interface{}) *insyra.DataList {
	return W.WWeightedMovingAverage(a0, a1)
}

// _github_com_HazelnutParadise_insyra_IDataTable is an interface wrapper for IDataTable type
type _github_com_HazelnutParadise_insyra_IDataTable struct {
	IValue                                 interface{}
	WAppendCols                            func(columns ...*insyra.DataList) *insyra.DataTable
	WAppendRowsByColIndex                  func(rowsData ...map[string]interface{}) *insyra.DataTable
	WAppendRowsByColName                   func(rowsData ...map[string]interface{}) *insyra.DataTable
	WAppendRowsFromDataList                func(rowsData ...*insyra.DataList) *insyra.DataTable
	WCount                                 func(value interface{}) int
	WData                                  func(useNamesAsKeys ...bool) map[string][]interface{}
	WDropColsByIndex                       func(columnIndices ...string)
	WDropColsByName                        func(columnNames ...string)
	WDropColsByNumber                      func(columnIndices ...int)
	WDropColsContainNil                    func()
	WDropColsContainNumbers                func()
	WDropColsContainStringElements         func()
	WDropRowsByIndex                       func(rowIndices ...int)
	WDropRowsByName                        func(rowNames ...string)
	WDropRowsContainNil                    func()
	WDropRowsContainNumbers                func()
	WDropRowsContainStringElements         func()
	WFilter                                func(filterFunc insyra.FilterFunc) *insyra.DataTable
	WFilterByColIndexEqualTo               func(index string) *insyra.DataTable
	WFilterByColIndexGreaterThan           func(threshold string) *insyra.DataTable
	WFilterByColIndexGreaterThanOrEqualTo  func(threshold string) *insyra.DataTable
	WFilterByColIndexLessThan              func(threshold string) *insyra.DataTable
	WFilterByColIndexLessThanOrEqualTo     func(threshold string) *insyra.DataTable
	WFilterByColNameContains               func(substring string) *insyra.DataTable
	WFilterByColNameEqualTo                func(name string) *insyra.DataTable
	WFilterByCustomElement                 func(f func(value interface{}) bool) *insyra.DataTable
	WFilterByRowIndexEqualTo               func(index int) *insyra.DataTable
	WFilterByRowIndexGreaterThan           func(threshold int) *insyra.DataTable
	WFilterByRowIndexGreaterThanOrEqualTo  func(threshold int) *insyra.DataTable
	WFilterByRowIndexLessThan              func(threshold int) *insyra.DataTable
	WFilterByRowIndexLessThanOrEqualTo     func(threshold int) *insyra.DataTable
	WFilterByRowNameContains               func(substring string) *insyra.DataTable
	WFilterByRowNameEqualTo                func(name string) *insyra.DataTable
	WFindColsIfAllElementsContainSubstring func(substring string) []string
	WFindColsIfAnyElementContainsSubstring func(substring string) []string
	WFindColsIfContains                    func(value interface{}) []string
	WFindColsIfContainsAll                 func(values ...interface{}) []string
	WFindRowsIfAllElementsContainSubstring func(substring string) []int
	WFindRowsIfAnyElementContainsSubstring func(substring string) []int
	WFindRowsIfContains                    func(value interface{}) []int
	WFindRowsIfContainsAll                 func(values ...interface{}) []int
	WGetCol                                func(index string) *insyra.DataList
	WGetColByNumber                        func(index int) *insyra.DataList
	WGetCreationTimestamp                  func() int64
	WGetElement                            func(rowIndex int, columnIndex string) interface{}
	WGetElementByNumberIndex               func(rowIndex int, columnIndex int) interface{}
	WGetLastModifiedTimestamp              func() int64
	WGetRow                                func(index int) *insyra.DataList
	WGetRowNameByIndex                     func(index int) string
	WLoadFromCSV                           func(filePath string, setFirstColToRowNames bool, setFirstRowToColNames bool) error
	WMean                                  func() interface{}
	WSetColToRowNames                      func(columnIndex string) *insyra.DataTable
	WSetRowNameByIndex                     func(index int, name string)
	WSetRowToColNames                      func(rowIndex int) *insyra.DataTable
	WShow                                  func()
	WShowTypes                             func()
	WSize                                  func() (int, int)
	WToCSV                                 func(filePath string, setRowNamesToFirstCol bool, setColNamesToFirstRow bool) error
	WTranspose                             func() *insyra.DataTable
	WUpdateCol                             func(index string, dl *insyra.DataList)
	WUpdateColByNumber                     func(index int, dl *insyra.DataList)
	WUpdateElement                         func(rowIndex int, columnIndex string, value interface{})
	WUpdateRow                             func(index int, dl *insyra.DataList)
}

func (W _github_com_HazelnutParadise_insyra_IDataTable) AppendCols(columns ...*insyra.DataList) *insyra.DataTable {
	return W.WAppendCols(columns...)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) AppendRowsByColIndex(rowsData ...map[string]interface{}) *insyra.DataTable {
	return W.WAppendRowsByColIndex(rowsData...)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) AppendRowsByColName(rowsData ...map[string]interface{}) *insyra.DataTable {
	return W.WAppendRowsByColName(rowsData...)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) AppendRowsFromDataList(rowsData ...*insyra.DataList) *insyra.DataTable {
	return W.WAppendRowsFromDataList(rowsData...)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) Count(value interface{}) int {
	return W.WCount(value)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) Data(useNamesAsKeys ...bool) map[string][]interface{} {
	return W.WData(useNamesAsKeys...)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) DropColsByIndex(columnIndices ...string) {
	W.WDropColsByIndex(columnIndices...)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) DropColsByName(columnNames ...string) {
	W.WDropColsByName(columnNames...)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) DropColsByNumber(columnIndices ...int) {
	W.WDropColsByNumber(columnIndices...)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) DropColsContainNil() {
	W.WDropColsContainNil()
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) DropColsContainNumbers() {
	W.WDropColsContainNumbers()
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) DropColsContainStringElements() {
	W.WDropColsContainStringElements()
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) DropRowsByIndex(rowIndices ...int) {
	W.WDropRowsByIndex(rowIndices...)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) DropRowsByName(rowNames ...string) {
	W.WDropRowsByName(rowNames...)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) DropRowsContainNil() {
	W.WDropRowsContainNil()
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) DropRowsContainNumbers() {
	W.WDropRowsContainNumbers()
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) DropRowsContainStringElements() {
	W.WDropRowsContainStringElements()
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) Filter(filterFunc insyra.FilterFunc) *insyra.DataTable {
	return W.WFilter(filterFunc)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FilterByColIndexEqualTo(index string) *insyra.DataTable {
	return W.WFilterByColIndexEqualTo(index)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FilterByColIndexGreaterThan(threshold string) *insyra.DataTable {
	return W.WFilterByColIndexGreaterThan(threshold)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FilterByColIndexGreaterThanOrEqualTo(threshold string) *insyra.DataTable {
	return W.WFilterByColIndexGreaterThanOrEqualTo(threshold)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FilterByColIndexLessThan(threshold string) *insyra.DataTable {
	return W.WFilterByColIndexLessThan(threshold)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FilterByColIndexLessThanOrEqualTo(threshold string) *insyra.DataTable {
	return W.WFilterByColIndexLessThanOrEqualTo(threshold)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FilterByColNameContains(substring string) *insyra.DataTable {
	return W.WFilterByColNameContains(substring)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FilterByColNameEqualTo(name string) *insyra.DataTable {
	return W.WFilterByColNameEqualTo(name)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FilterByCustomElement(f func(value interface{}) bool) *insyra.DataTable {
	return W.WFilterByCustomElement(f)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FilterByRowIndexEqualTo(index int) *insyra.DataTable {
	return W.WFilterByRowIndexEqualTo(index)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FilterByRowIndexGreaterThan(threshold int) *insyra.DataTable {
	return W.WFilterByRowIndexGreaterThan(threshold)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FilterByRowIndexGreaterThanOrEqualTo(threshold int) *insyra.DataTable {
	return W.WFilterByRowIndexGreaterThanOrEqualTo(threshold)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FilterByRowIndexLessThan(threshold int) *insyra.DataTable {
	return W.WFilterByRowIndexLessThan(threshold)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FilterByRowIndexLessThanOrEqualTo(threshold int) *insyra.DataTable {
	return W.WFilterByRowIndexLessThanOrEqualTo(threshold)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FilterByRowNameContains(substring string) *insyra.DataTable {
	return W.WFilterByRowNameContains(substring)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FilterByRowNameEqualTo(name string) *insyra.DataTable {
	return W.WFilterByRowNameEqualTo(name)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FindColsIfAllElementsContainSubstring(substring string) []string {
	return W.WFindColsIfAllElementsContainSubstring(substring)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FindColsIfAnyElementContainsSubstring(substring string) []string {
	return W.WFindColsIfAnyElementContainsSubstring(substring)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FindColsIfContains(value interface{}) []string {
	return W.WFindColsIfContains(value)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FindColsIfContainsAll(values ...interface{}) []string {
	return W.WFindColsIfContainsAll(values...)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FindRowsIfAllElementsContainSubstring(substring string) []int {
	return W.WFindRowsIfAllElementsContainSubstring(substring)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FindRowsIfAnyElementContainsSubstring(substring string) []int {
	return W.WFindRowsIfAnyElementContainsSubstring(substring)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FindRowsIfContains(value interface{}) []int {
	return W.WFindRowsIfContains(value)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) FindRowsIfContainsAll(values ...interface{}) []int {
	return W.WFindRowsIfContainsAll(values...)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) GetCol(index string) *insyra.DataList {
	return W.WGetCol(index)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) GetColByNumber(index int) *insyra.DataList {
	return W.WGetColByNumber(index)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) GetCreationTimestamp() int64 {
	return W.WGetCreationTimestamp()
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) GetElement(rowIndex int, columnIndex string) interface{} {
	return W.WGetElement(rowIndex, columnIndex)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) GetElementByNumberIndex(rowIndex int, columnIndex int) interface{} {
	return W.WGetElementByNumberIndex(rowIndex, columnIndex)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) GetLastModifiedTimestamp() int64 {
	return W.WGetLastModifiedTimestamp()
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) GetRow(index int) *insyra.DataList {
	return W.WGetRow(index)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) GetRowNameByIndex(index int) string {
	return W.WGetRowNameByIndex(index)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) LoadFromCSV(filePath string, setFirstColToRowNames bool, setFirstRowToColNames bool) error {
	return W.WLoadFromCSV(filePath, setFirstColToRowNames, setFirstRowToColNames)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) Mean() interface{} {
	return W.WMean()
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) SetColToRowNames(columnIndex string) *insyra.DataTable {
	return W.WSetColToRowNames(columnIndex)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) SetRowNameByIndex(index int, name string) {
	W.WSetRowNameByIndex(index, name)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) SetRowToColNames(rowIndex int) *insyra.DataTable {
	return W.WSetRowToColNames(rowIndex)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) Show() {
	W.WShow()
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) ShowTypes() {
	W.WShowTypes()
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) Size() (int, int) {
	return W.WSize()
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) ToCSV(filePath string, setRowNamesToFirstCol bool, setColNamesToFirstRow bool) error {
	return W.WToCSV(filePath, setRowNamesToFirstCol, setColNamesToFirstRow)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) Transpose() *insyra.DataTable {
	return W.WTranspose()
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) UpdateCol(index string, dl *insyra.DataList) {
	W.WUpdateCol(index, dl)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) UpdateColByNumber(index int, dl *insyra.DataList) {
	W.WUpdateColByNumber(index, dl)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) UpdateElement(rowIndex int, columnIndex string, value interface{}) {
	W.WUpdateElement(rowIndex, columnIndex, value)
}
func (W _github_com_HazelnutParadise_insyra_IDataTable) UpdateRow(index int, dl *insyra.DataList) {
	W.WUpdateRow(index, dl)
}
