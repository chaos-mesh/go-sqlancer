package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/chaos-mesh/go-sqlancer/pkg/pivot"
	_ "github.com/go-sql-driver/mysql"
)

const (
	nmDSN       = "dsn"
	nmViewCount = "view"
	nmDuration  = "duration"
	sqlDepth    = "depth"
	silentMode  = "silent"
	nmDebug     = "debug"
	nmHint      = "hint"
	nmExprIdx   = "expr-index"
	nmMode      = "mode"
)

var (
	conf      = pivot.NewConfig()
	dsn       = flag.String(nmDSN, "", "dsn of target db for testing")
	viewCount = flag.Int(nmViewCount, 10, "count of views to be created")
	duration  = flag.Duration(nmDuration, 5*time.Minute, "fuzz duration")
	depth     = flag.Int(sqlDepth, 1, "sql depth")
	silent    = flag.Bool(silentMode, false, "silent when verify failed")
	debug     = flag.Bool(nmDebug, false, "enable debug output")
	hint      = flag.Bool(nmHint, false, "enable sql hint for TiDB")
	exprIdx   = flag.Bool(nmExprIdx, false, "enable create expression index")
	mode      = flag.String(nmMode, "pqs|norec", "use NoRec or PQS method or both, split by vertical bar")
)

func main() {
	loadConfig()
	p, err := pivot.NewPivot(conf)
	if err != nil {
		panic(fmt.Sprintf("new pivot failed, error: %+v\n", err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()
	fmt.Printf("start pivot test\n")
	p.Start(ctx)
	fmt.Printf("wait for finish\n")

	<-ctx.Done()
	p.Close()
}

func loadConfig() {
	flag.Parse()

	actualFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		actualFlags[f.Name] = true
	})

	if actualFlags[nmDSN] {
		if err := conf.SetDSN(*dsn); err != nil {
			panic(err)
		}
	} else {
		panic("empty dsn")
	}
	if actualFlags[nmViewCount] {
		conf.ViewCount = *viewCount
	}
	if actualFlags[sqlDepth] {
		conf.Depth = *depth
	}
	if actualFlags[silentMode] {
		conf.Silent = *silent
	}
	if actualFlags[nmDebug] {
		conf.Debug = *debug
	}
	if actualFlags[nmHint] {
		conf.Hint = *hint
	}
	if actualFlags[nmExprIdx] {
		conf.ExprIndex = *exprIdx
	}
	if actualFlags[nmMode] {
		if len(*mode) == 0 {
			panic("empty mode param set")
		}
		approaches := strings.Split(*mode, "|")
		hasSet := false
		for _, i := range approaches {
			if strings.ToLower(i) == "norec" {
				conf.ModeNoREC = true
				hasSet = true
			}
			if strings.ToLower(i) == "pqs" {
				conf.ModePQS = true
				hasSet = true
			}
		}
		if !hasSet {
			panic("no valid mode param set")
		}
	}
}
