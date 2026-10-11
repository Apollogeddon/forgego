package templates

import "embed"

// The vendored flexsearch/ follows Hextra's pin: FlexSearchVersion is the release HextraVersion's
// search partial loads, so a Hextra that moves to a new FlexSearch needs a new copy.
const (
	HextraVersion     = "v0.13.0"
	FlexSearchVersion = "0.8.143"
)

// FlexSearchDir is where a site's vendored FlexSearch lives, under the site's assets/.
const FlexSearchDir = "assets/forgego/flexsearch"

// Hextra fetches FlexSearch from a CDN at build time unless the site points it at a local asset.
//
//go:embed flexsearch
var flexSearchFiles embed.FS

// SiteFile is a file forgego writes into a website and keeps up to date.
type SiteFile struct {
	Path, Content string
}

// SiteFiles are the files forgego manages in a website: FlexSearch and its licence.
func SiteFiles() []SiteFile {
	var files []SiteFile
	for _, name := range []string{"flexsearch.bundle.min.js", "LICENSE"} {
		b, err := flexSearchFiles.ReadFile("flexsearch/" + name)
		if err != nil {
			panic(err)
		}
		files = append(files, SiteFile{FlexSearchDir + "/" + name, string(b)})
	}
	return files
}

// SiteGoMod starts a website's go.mod, pinning the Hextra release the vendored FlexSearch matches.
const SiteGoMod = GoMod + `
require github.com/imfing/hextra ` + HextraVersion + ` // indirect
`

// HugoConfig is a documentation site using the Hextra theme as a Hugo module, so the
// theme is versioned in go.mod like any other dependency.
const HugoConfig = `baseURL = "/"
title = "__NAME__"
locale = "en"
enableGitInfo = false

[module]
  [[module.imports]]
    path = "github.com/imfing/hextra"

[markup.goldmark.renderer]
  unsafe = false

[menu]
  [[menu.main]]
    name = "Docs"
    pageRef = "/docs"
    weight = 1
  [[menu.main]]
    name = "Search"
    weight = 2
    [menu.main.params]
      type = "search"

[params.navbar]
  displayTitle = true

[params.footer]
  displayPoweredBy = false

# FlexSearch from the site's assets, so the site builds without network access
[params.search.flexsearch]
  js = "forgego/flexsearch/flexsearch.bundle.min.js"
`

// SiteHome is the site's landing page.
const SiteHome = `---
title: __NAME__
layout: hextra-home
---

{{< hextra/hero-headline >}}
  __NAME__
{{< /hextra/hero-headline >}}

{{< hextra/hero-subtitle >}}
  Documentation for __NAME__.
{{< /hextra/hero-subtitle >}}

{{< hextra/hero-button text="Get Started" link="docs" >}}
`

// SiteDocsIndex is the first documentation page.
const SiteDocsIndex = `---
title: Getting Started
weight: 1
---

Write your documentation in Markdown under ` + "`content/docs/`" + `. Run ` + "`task dev`" + ` to preview
it with live reload, and ` + "`task build`" + ` to build the site into ` + "`public/`" + `.
`
