/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package merge_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	. "sigs.k8s.io/structured-merge-diff/internal/fixture"
	"sigs.k8s.io/structured-merge-diff/typed"
)

func testdata(file string) string {
	return filepath.Join("..", "internal", "testdata", file)
}

func read(file string) []byte {
	s, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return s
}

func lastPart(s string) string {
	return s[strings.LastIndex(s, ".")+1:]
}

var parser = func() *typed.Parser {
	s := read(testdata("k8s-schema.yaml"))
	parser, err := typed.NewParser(typed.YAMLObject(s))
	if err != nil {
		panic(err)
	}
	return parser
}()

func BenchmarkOperations(b *testing.B) {
	benches := []struct {
		typename string
		obj      typed.YAMLObject
	}{
		{
			typename: "io.k8s.api.core.v1.Pod",
			obj:      typed.YAMLObject(read(testdata("pod.yaml"))),
		},
		{
			typename: "io.k8s.api.core.v1.Node",
			obj:      typed.YAMLObject(read(testdata("node.yaml"))),
		},
		{
			typename: "io.k8s.api.core.v1.Endpoints",
			obj:      typed.YAMLObject(read(testdata("endpoints.yaml"))),
		},
	}

	for _, bench := range benches {

		b.Run(lastPart(bench.typename), func(b *testing.B) {
			ops := []Operation{
				Update{
					Manager:    "controller",
					APIVersion: "v1",
					Object:     bench.obj,
				},
				Apply{
					Manager:    "controller",
					APIVersion: "v1",
					Object:     bench.obj,
				},
			}
			for _, op := range ops {
				b.Run(lastPart(fmt.Sprintf("%T", op)), func(b *testing.B) {
					test := TestCase{
						Ops: []Operation{op},
					}

					pt := parser.Type(bench.typename)
					test.PreprocessOperations(pt)

					b.ReportAllocs()
					b.ResetTimer()
					for n := 0; n < b.N; n++ {
						if err := test.Bench(pt); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		})
	}
}
