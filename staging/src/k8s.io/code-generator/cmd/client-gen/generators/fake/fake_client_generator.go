/*
Copyright 2015 The Kubernetes Authors.

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

package fake

import (
	"path/filepath"
	"strings"

	"k8s.io/gengo/v2/generator"
	"k8s.io/gengo/v2/types"

	clientgenargs "k8s.io/code-generator/cmd/client-gen/args"
	scheme "k8s.io/code-generator/cmd/client-gen/generators/scheme"
	"k8s.io/code-generator/cmd/client-gen/generators/util"
	clientgentypes "k8s.io/code-generator/cmd/client-gen/types"
)

func TargetForGroup(gv clientgentypes.GroupVersion, typeList []*types.Type, clientsetDir, clientsetPkg string, groupPkgName string, groupGoName string, inputPkg string, applyBuilderPackage string, boilerplate []byte) generator.Target {
	// TODO: should make this a function, called by here and in client-generator.go
	subdir := filepath.Join("typed", strings.ToLower(groupPkgName), strings.ToLower(gv.Version.NonEmpty()))
	outputDir := filepath.Join(clientsetDir, subdir, "fake")
	outputPkg := filepath.Join(clientsetPkg, subdir, "fake")
	realClientPkg := filepath.Join(clientsetPkg, subdir)

	return &generator.SimpleTarget{
		PkgName:       "fake",
		PkgPath:       outputPkg,
		PkgDir:        outputDir,
		HeaderComment: boilerplate,
		PkgDocComment: []byte("// Package fake has the automatically generated clients.\n"),
		// GeneratorsFunc returns a list of generators. Each generator makes a
		// single file.
		GeneratorsFunc: func(c *generator.Context) (generators []generator.Generator) {
			generators = []generator.Generator{
				// Always generate a "doc.go" file.
				generator.DefaultGen{OptionalName: "doc"},
			}
			// Since we want a file per type that we generate a client for, we
			// have to provide a function for this.
			for _, t := range typeList {
				generators = append(generators, &genFakeForType{
					DefaultGen: generator.DefaultGen{
						OptionalName: "fake_" + strings.ToLower(c.Namers["private"].Name(t)),
					},
					outputPackage:             outputPkg,
					inputPackage:              inputPkg,
					group:                     gv.Group.NonEmpty(),
					version:                   gv.Version.String(),
					groupGoName:               groupGoName,
					typeToMatch:               t,
					imports:                   generator.NewImportTracker(),
					applyConfigurationPackage: applyBuilderPackage,
				})
			}

			generators = append(generators, &genFakeForGroup{
				DefaultGen: generator.DefaultGen{
					OptionalName: "fake_" + groupPkgName + "_client",
				},
				outputPackage:     outputPkg,
				realClientPackage: realClientPkg,
				group:             gv.Group.NonEmpty(),
				version:           gv.Version.String(),
				groupGoName:       groupGoName,
				types:             typeList,
				imports:           generator.NewImportTracker(),
			})
			return generators
		},
		FilterFunc: func(c *generator.Context, t *types.Type) bool {
			return util.MustParseClientGenTags(append(t.SecondClosestCommentLines, t.CommentLines...)).GenerateClient
		},
	}
}

func TargetForClientset(customArgs *clientgenargs.CustomArgs, clientsetDir, clientsetPkg string, groupGoNames map[clientgentypes.GroupVersion]string, boilerplate []byte) generator.Target {
	return &generator.SimpleTarget{
		// TODO: we'll generate fake clientset for different release in the future.
		// Package name and path are hard coded for now.
		PkgName:       "fake",
		PkgPath:       filepath.Join(clientsetPkg, "fake"),
		PkgDir:        filepath.Join(clientsetDir, "fake"),
		HeaderComment: boilerplate,
		PkgDocComment: []byte("// This package has the automatically generated fake clientset.\n"),
		// GeneratorsFunc returns a list of generators. Each generator generates a
		// single file.
		GeneratorsFunc: func(c *generator.Context) (generators []generator.Generator) {
			generators = []generator.Generator{
				// Always generate a "doc.go" file.
				generator.DefaultGen{OptionalName: "doc"},

				&genClientset{
					DefaultGen: generator.DefaultGen{
						OptionalName: "clientset_generated",
					},
					groups:               customArgs.Groups,
					groupGoNames:         groupGoNames,
					fakeClientsetPackage: clientsetPkg,
					imports:              generator.NewImportTracker(),
					realClientsetPackage: clientsetPkg,
				},
				&scheme.GenScheme{
					DefaultGen: generator.DefaultGen{
						OptionalName: "register",
					},
					InputPackages: customArgs.GroupVersionPackages(),
					OutputPackage: clientsetPkg,
					Groups:        customArgs.Groups,
					GroupGoNames:  groupGoNames,
					ImportTracker: generator.NewImportTracker(),
					PrivateScheme: true,
				},
			}
			return generators
		},
	}
}
