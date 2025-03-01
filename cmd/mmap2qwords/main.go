package main

import (
	"context"
	"fmt"
	"iter"
	"log"
	"os"

	mi "github.com/takanoriyanagitani/go-mmap2ints"
	. "github.com/takanoriyanagitani/go-mmap2ints/util"
)

var envVarByKey func(key string) IO[string] = Lift(
	func(key string) (string, error) {
		val, found := os.LookupEnv(key)
		switch found {
		case true:
			return val, nil
		default:
			return "", fmt.Errorf("env var %s missing", key)
		}
	},
)

func printQword(i uint64) IO[Void] {
	return func(_ context.Context) (Void, error) {
		_, e := fmt.Printf("%016x\n", i)
		return Empty, e
	}
}

func printQwords(ints iter.Seq2[uint64, error]) IO[Void] {
	return func(ctx context.Context) (Void, error) {
		for i, e := range ints {
			select {
			case <-ctx.Done():
				return Empty, ctx.Err()
			default:
			}

			if nil != e {
				return Empty, e
			}

			_, e := printQword(i)(ctx)
			if nil != e {
				return Empty, e
			}
		}

		return Empty, nil
	}
}

var filename IO[string] = envVarByKey("ENV_INTS_FILENAME")

var ints IO[iter.Seq2[uint64, error]] = Bind(
	filename,
	Lift(func(filename string) (iter.Seq2[uint64, error], error) {
		return mi.FilenameToIntegers64(filename), nil
	}),
)

var filename2ints2stdout IO[Void] = Bind(
	ints,
	printQwords,
)

var sub IO[Void] = func(ctx context.Context) (Void, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	return filename2ints2stdout(ctx)
}

func main() {
	_, e := sub(context.Background())
	if nil != e {
		log.Printf("%v\n", e)
	}
}
