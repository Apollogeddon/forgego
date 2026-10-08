package templates

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
