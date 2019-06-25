package template_test

import (
	"testing"

	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/files"
)

func TestDocumentOverlays(t *testing.T) {
	yamlTplData := []byte(`
array:
- name: item1
  subarray:
  - item1
`)

	yamlOverlayTplData := []byte(`
#@ load("@ytt:overlay", "overlay")
#@ load("funcs/funcs.lib.yml", "yamlfunc")
#@overlay/match by=overlay.all
---
array:
#@overlay/match by="name"
- name: item1
  #@overlay/match missing_ok=True
  subarray2:
  - #@ yamlfunc()
`)

	expectedYAMLTplData := `array:
- name: item1
  subarray:
  - item1
  subarray2:
  - yamlfunc: yamlfunc
`

	yamlFuncsData := []byte(`
#@ def/end yamlfunc():
yamlfunc: yamlfunc`)

	filesToProcess := []*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("overlay.yml", yamlOverlayTplData)),
		files.MustNewFileFromSource(files.NewBytesSource("funcs/funcs.lib.yml", yamlFuncsData)),
	}

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but was error: %s", out.Err)
	}

	if len(out.Files) != 1 {
		t.Fatalf("Expected number of output files to be 1, but was %d", len(out.Files))
	}

	file := out.Files[0]

	if file.RelativePath() != "tpl.yml" {
		t.Fatalf("Expected output file to be tpl.yml, but was %#v", file.RelativePath())
	}

	if string(file.Bytes()) != expectedYAMLTplData {
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<<", file.Bytes())
	}
}

func TestDocumentOverlays2(t *testing.T) {
	yamlTplData := []byte(`
array:
- name: item1
  subarray:
  - item1
`)

	yamlOverlayTplData1 := []byte(`
#@ load("@ytt:overlay", "overlay")
#@ load("funcs/funcs.lib.yml", "yamlfunc")
#@overlay/match by=overlay.all
---
array:
#@overlay/match by="name"
- name: item1
  #@overlay/match missing_ok=True
  subarray2:
  - #@ yamlfunc()
`)

	yamlOverlayTplData2 := []byte(`
#@ load("@ytt:overlay", "overlay")
#@overlay/match by=overlay.all
---
array:
#@overlay/match by="name"
- name: item1
  #@overlay/remove
  subarray2:
`)

	// subarray2 is not present because it was removed
	// by overlay1 that comes after overlay2
	expectedYAMLTplData := `array:
- name: item1
  subarray:
  - item1
`

	yamlFuncsData := []byte(`
#@ def/end yamlfunc():
yamlfunc: yamlfunc`)

	filesToProcess := files.NewSortedFiles([]*files.File{
		files.MustNewFileFromSource(files.NewBytesSource("tpl.yml", yamlTplData)),
		// Note that overlay1 is alphanumerically before overlay2
		// but sorting of the files puts them in overlay2 then overlay1 order
		files.MustNewFileFromSource(files.NewBytesSource("overlay2.yml", yamlOverlayTplData1)),
		files.MustNewFileFromSource(files.NewBytesSource("overlay1.yml", yamlOverlayTplData2)),
		files.MustNewFileFromSource(files.NewBytesSource("funcs/funcs.lib.yml", yamlFuncsData)),
	})

	ui := cmdcore.NewPlainUI(false)
	opts := cmdtpl.NewOptions()

	out := opts.RunWithFiles(cmdtpl.TemplateInput{Files: filesToProcess}, ui)
	if out.Err != nil {
		t.Fatalf("Expected RunWithFiles to succeed, but was error: %s", out.Err)
	}

	if len(out.Files) != 1 {
		t.Fatalf("Expected number of output files to be 1, but was %d", len(out.Files))
	}

	file := out.Files[0]

	if file.RelativePath() != "tpl.yml" {
		t.Fatalf("Expected output file to be tpl.yml, but was %#v", file.RelativePath())
	}

	if string(file.Bytes()) != expectedYAMLTplData {
		t.Fatalf("Expected output file to have specific data, but was: >>>%s<<<", file.Bytes())
	}
}