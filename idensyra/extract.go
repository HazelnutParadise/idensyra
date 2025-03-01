//go:generate go install github.com/traefik/yaegi/cmd/yaegi@latest
//go:generate $GOPATH/bin/yaegi extract github.com/HazelnutParadise/insyra
//go:generate $GOPATH/bin/yaegi extract github.com/HazelnutParadise/insyra/isr
//go:generate $GOPATH/bin/yaegi extract github.com/HazelnutParadise/insyra/datafetch
//go:generate $GOPATH/bin/yaegi extract github.com/HazelnutParadise/insyra/stats
//go:generate $GOPATH/bin/yaegi extract github.com/HazelnutParadise/insyra/parallel
//go:generate $GOPATH/bin/yaegi extract github.com/HazelnutParadise/insyra/plot
//go:generate $GOPATH/bin/yaegi extract github.com/HazelnutParadise/insyra/gplot
//go:generate $GOPATH/bin/yaegi extract github.com/HazelnutParadise/insyra/lpgen
//go:generate $GOPATH/bin/yaegi extract github.com/HazelnutParadise/insyra/csvxl
//go:generate $GOPATH/bin/yaegi extract github.com/HazelnutParadise/insyra/py

package idensyra

import "reflect"

var Symbols = map[string]map[string]reflect.Value{}
