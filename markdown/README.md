Markdown Builder
================

[![Godoc reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/wI2L/fizz/markdown)

This is a simple builder to help you create **Mardown** content in Go.

## Usage

```go
import "github.com/wI2L/fizz/markdown"

builder := markdown.Builder{}

builder.
   H1("Markdown Builder").
   P("A simple builder to help your write Markdown in Go").
   H2("Installation").
   Code("go get -u github.com/wI2L/fizz", "bash").
   H2("Todos").
   BulletedList(
     "write tests",
     builder.Block().NumberedList("A", "B", "C"),
     "add more markdown features",
   )

md := builder.String()
```

    Markdown Builder
    ================

    A simple builder to help your write Markdown in Go.

    Installation
    ----------------------------------

    ```bash
    go get -u github.com/wI2L/fizz", "bash
    ```

    Todos
    -----
    * write tests
    * 1. A
      2. B
      3. C
    * add more markdown features

## Elements

The builder have two kinds of elements, block and inline elements. Block elements are concatenated to the underlying builder buffer while inline elements are directly returned as string.

### Inline

* Alternative titles ⇒ `builder.AltH1` and `builder.AltH2`
* Emphasis ⇒ `builder.Emphasis` and `builder.Italic`
* Strong emphasis ⇒ `builder.StrongEmphasis` and `builder.Bold`
* Combined emphasis ⇒ `builder.StrongEmphasis`
* Strikethrough ⇒ `builder.Strikethrough`
* Code ⇒ `builder.InlineCode`
* Link ⇒ `builder.Link`
* Image ⇒ `builder.Image`

### Block

* Headers ⇒ `builder.H1`, `builder.H2`, `builder.H3`, `builder.H4`, `builder.H5`, `builder.H6`
* Paragraph ⇒ `builder.P`
* Line ⇒ `builder.Line`
  A line is similar to a paragraph but it doesn't insert a line break at the end.
* Line break ⇒ `builder.BR`
* Blockquote ⇒ `builder.Blockquote`
* Horizontal rule ⇒ `builder.HR`
* Code ⇒ `builder.Code`
* Lists
  1. Bulleted ⇒ `builder.BulletedList`
  2. Numbered ⇒ `builder.NumberedList`

### Github Flavor

The builder also support the tables extension of the *GitHub Flavored Markdown Spec*.

```go
builder.Table(
  [][]string{
    []string{"Letter", "Title", "ID"},
    []string{"A", "The Good", "500"},
    []string{"B", "The Very very Bad Man", "2885645"},
    []string{"C", "The Ugly"},
    []string{"D", "The\nGopher", "800"},
  },[]markdown.TableAlignment{
      markdown.AlignCenter,
      markdown.AlignCenter,
      markdown.AlignRight,
  }
)
```
```
| Letter |         Title         |      ID |
| :----: | :-------------------: | ------: |
|   A    |       The Good        |     500 |
|   B    | The Very very Bad Man | 2885645 |
|   C    |       The Ugly        |         |
|   D    |      The Gopher       |     800 |
```

## Credits

This work is based on this PHP [markdown builder](https://github.com/DavidBadura/markdown-builder).
