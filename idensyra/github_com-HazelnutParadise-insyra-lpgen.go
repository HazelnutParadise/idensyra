// Code generated by 'yaegi extract github.com/HazelnutParadise/insyra/lpgen'. DO NOT EDIT.

package idensyra

import (
	"github.com/HazelnutParadise/insyra/lpgen"
	"reflect"
)

func init() {
	Symbols["github.com/HazelnutParadise/insyra/lpgen/lpgen"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"NewLPModel":          reflect.ValueOf(lpgen.NewLPModel),
		"ParseLingoModel_str": reflect.ValueOf(lpgen.ParseLingoModel_str),
		"ParseLingoModel_txt": reflect.ValueOf(lpgen.ParseLingoModel_txt),

		// type definitions
		"LPModel": reflect.ValueOf((*lpgen.LPModel)(nil)),
	}
}
