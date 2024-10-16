// Code generated by 'yaegi extract github.com/HazelnutParadise/insyra/gplot'. DO NOT EDIT.

package idensyra

import (
	"github.com/HazelnutParadise/insyra/gplot"
	"reflect"
)

func init() {
	Symbols["github.com/HazelnutParadise/insyra/gplot/gplot"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"CreateBarChart":     reflect.ValueOf(gplot.CreateBarChart),
		"CreateFunctionPlot": reflect.ValueOf(gplot.CreateFunctionPlot),
		"CreateHistogram":    reflect.ValueOf(gplot.CreateHistogram),
		"CreateLineChart":    reflect.ValueOf(gplot.CreateLineChart),
		"SaveChart":          reflect.ValueOf(gplot.SaveChart),

		// type definitions
		"BarChartConfig":     reflect.ValueOf((*gplot.BarChartConfig)(nil)),
		"FunctionPlotConfig": reflect.ValueOf((*gplot.FunctionPlotConfig)(nil)),
		"HistogramConfig":    reflect.ValueOf((*gplot.HistogramConfig)(nil)),
		"LineChartConfig":    reflect.ValueOf((*gplot.LineChartConfig)(nil)),
	}
}
