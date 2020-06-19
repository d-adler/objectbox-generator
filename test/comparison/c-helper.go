/*
 * Copyright (C) 2020 ObjectBox Ltd. All rights reserved.
 * https://objectbox.io
 *
 * This file is part of ObjectBox Generator.
 *
 * ObjectBox Generator is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 * ObjectBox Generator is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with ObjectBox Generator.  If not, see <http://www.gnu.org/licenses/>.
 */

package comparison

import (
	"io/ioutil"
	"path"
	"path/filepath"
	"testing"

	"github.com/objectbox/objectbox-generator/internal/generator"
	cgenerator "github.com/objectbox/objectbox-generator/internal/generator/c"
	"github.com/objectbox/objectbox-generator/test/assert"
	"github.com/objectbox/objectbox-generator/test/cmake"
)

type cTestHelper struct {
	cpp        bool
	canCompile bool
}

func (h *cTestHelper) init(t *testing.T, conf testSpec) {
	if !testing.Short() {
		h.canCompile = true

		{ // check objectbox lib
			var includeFiles = []string{"objectbox.h"}
			if h.cpp {
				includeFiles = append(includeFiles, "objectbox-cpp.h")
			}
			assert.NoErr(t, cmake.LibraryExists("objectbox", includeFiles))
		}

		// check flatbuffers library availability
		if h.cpp {
			// Cpp compilation is mandatory.
			// Note: we don't need flatbuffers library explicitly, it's part of objectbox at the moment.
			assert.NoErr(t, cmake.LibraryExists("", []string{"flatbuffers/flatbuffers.h"}))
		} else {

			err := cmake.LibraryExists("flatccrt", []string{"flatcc/flatcc.h", "flatcc/flatcc_builder.h"})
			if err != nil {
				t.Logf("C compilation not available, it will be skipped during tests, because %s", err)
				h.canCompile = false
			}
		}
	}
}

func (h cTestHelper) generatorFor(t *testing.T, conf testSpec, sourceFile string, genDir string) generator.CodeGenerator {
	// make a copy of the default generator
	var gen = *conf.generator.(*cgenerator.CGenerator)
	gen.OutPath = genDir
	return &gen
}

func (cTestHelper) prepareTempDir(t *testing.T, conf testSpec, srcDir, tempDir, tempRoot string) func(err error) error {
	return nil
}

func (h cTestHelper) build(t *testing.T, conf testSpec, dir string, expectedError error, errorTransformer func(err error) error) {
	if !h.canCompile {
		t.Skip("Compilation not available")
	}

	includeDir, err := filepath.Abs(dir) // main.c/cpp will include generated headers from here
	assert.NoErr(t, err)

	build := cmake.Cmake{
		Name:        "compilation-test",
		IsCpp:       h.cpp,
		IncludeDirs: []string{includeDir},
		LinkLibs:    []string{"objectbox"},
	}
	assert.NoErr(t, build.CreateTempDirs())
	defer build.RemoveTempDirs()

	var mainFile string
	if build.IsCpp {
		build.Standard = 11
		mainFile = path.Join(build.ConfDir, "main.cpp")
	} else {
		build.Standard = 99
		mainFile = path.Join(build.ConfDir, "main.c")
	}

	build.Files = append(build.Files, mainFile)

	assert.NoErr(t, build.WriteCMakeListsTxt())

	{ // write main.c/cpp to the conf dir - a simple one, just include all sources
		var mainSrc = ""
		if build.IsCpp {
			mainSrc = mainSrc + "#include \"objectbox-cpp.h\"\n"
		} else {
			mainSrc = mainSrc + "#include \"objectbox.h\"\n"
		}

		files, err := ioutil.ReadDir(includeDir)
		assert.NoErr(t, err)
		for _, file := range files {
			if conf.generator.IsGeneratedFile(file.Name()) {
				mainSrc = mainSrc + "#include \"" + file.Name() + "\"\n"
			}
		}

		mainSrc = mainSrc + "int main(){ return 0; }\n\n"
		t.Logf("main.c/cpp file contents \n%s", mainSrc)
		assert.NoErr(t, ioutil.WriteFile(mainFile, []byte(mainSrc), 0600))
	}

	if stdOut, stdErr, err := build.Configure(); err != nil {
		assert.Failf(t, "cmake configuration failed: \n%s\n%s\n%s", stdOut, stdErr, err)
	} else {
		t.Logf("configuration output:\n%s", string(stdOut))
	}

	if stdOut, stdErr, err := build.Build(); err != nil {
		checkBuildError(t, errorTransformer, stdOut, stdErr, err, expectedError)
		assert.Failf(t, "cmake build failed: \n%s\n%s\n%s", stdOut, stdErr, err)
	} else {
		t.Logf("build output:\n%s", string(stdOut))
	}
}
