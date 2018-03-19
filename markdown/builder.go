package markdown

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// TableAlignment represents the alignment
// of a table cell.
type TableAlignment uint8

// Alignment constants.
const (
	AlignLeft TableAlignment = iota
	AlignCenter
	AlignRight
)

var reCRLN = regexp.MustCompile(`\r?\n`)

// Builder is a simple Markdown builder that
// simplify the creation of content.
type Builder struct {
	markdown string
}

// String implements fmt.Stringer for Builder.
func (b *Builder) String() string {
	return strings.TrimSpace(b.nl().markdown)
}

// Block returns a new markdown block.
func (b *Builder) Block() *Builder {
	return new(Builder)
}

// Line adds a new line with no new line at
// the end to the markdown.
func (b *Builder) Line(text string) *Builder {
	return b.write(text)
}

// P adds a new paragraph to the markdown.
func (b *Builder) P(text string) *Builder {
	return b.writeln(text).nl()
}

// H1 adds a level 1 header to the markdown.
func (b *Builder) H1(text string) *Builder {
	return b.writeln(header(text, 1)).nl()
}

// H2 adds a level 2 header to the markdown.
func (b *Builder) H2(text string) *Builder {
	return b.writeln(header(text, 2)).nl()
}

// H3 adds a level 3 header to the markdown.
func (b *Builder) H3(text string) *Builder {
	return b.writeln(header(text, 3)).nl()
}

// H4 adds a level 4 header to the markdown.
func (b *Builder) H4(text string) *Builder {
	return b.writeln(header(text, 4)).nl()
}

// H5 adds a level 5 header to the markdown.
func (b *Builder) H5(text string) *Builder {
	return b.writeln(header(text, 5)).nl()
}

// H6 adds a level 6 header to the markdown.
func (b *Builder) H6(text string) *Builder {
	return b.writeln(header(text, 6)).nl()
}

func header(text string, level int) string {
	text = reCRLN.ReplaceAllString(text, " ")
	return fmt.Sprintf("%s %s", strings.Repeat("#", level), strings.TrimSpace(text))
}

func (b *Builder) nl() *Builder {
	return b.write("\n")
}

// AltH1 adds a level 1 header to the markdown.
// with an underline-ish style.
func (b *Builder) AltH1(header string) *Builder {
	header = reCRLN.ReplaceAllString(header, " ")
	header = strings.TrimSpace(header)
	return b.writeln(header).writeln(strings.Repeat("=", charLen(header))).nl()
}

// AltH2 adds a level 2 header to the markdown
// with an underline-ish style.
func (b *Builder) AltH2(header string) *Builder {
	header = reCRLN.ReplaceAllString(header, " ")
	header = strings.TrimSpace(header)
	return b.writeln(header).writeln(strings.Repeat("-", charLen(header))).nl()
}

// HR adds a horizontal rule to the markdown.
func (b *Builder) HR() *Builder {
	return b.P("---------------------------------------")
}

// BR adds a line break to the markdown.
func (b *Builder) BR() *Builder {
	return b.write("   \n")
}

// InlineCode returns a quoted inline code.
func (b *Builder) InlineCode(code string) string {
	return fmt.Sprintf("`%s`", code)
}

// Code adds a code portion with the given
// language format to the markdown.
func (b *Builder) Code(code, lang string) *Builder {
	return b.
		writelnf("```%s", lang).
		writeln(code).
		writeln("```").
		nl()
}

// Emphasis is an alias for Italic.
func (b *Builder) Emphasis(text string) string {
	return b.Italic(text)
}

// Italic returns an italic inlined text.
func (b *Builder) Italic(text string) string {
	return fmt.Sprintf("*%s*", text)
}

// StrongEmphasis is an alias for Bold.
func (b *Builder) StrongEmphasis(text string) string {
	return b.Bold(text)
}

// Bold returns a bold inlined text.
func (b *Builder) Bold(text string) string {
	return fmt.Sprintf("**%s**", text)
}

// CombinedEmphasis returns a bold/italic text.
func (b *Builder) CombinedEmphasis(text string) string {
	return fmt.Sprintf("**_%s_**", text)
}

// Strikethrough returns a strikethrough inlined text.
func (b *Builder) Strikethrough(text string) string {
	return fmt.Sprintf("~~%s~~", text)
}

// Link returns an inlined link.
func (b *Builder) Link(url, title string) string {
	return fmt.Sprintf("[%s](%s)", title, url)
}

// Image returns an inlined image.
func (b *Builder) Image(url, title string) string {
	return fmt.Sprintf("![%s](%s)", title, url)
}

// Blockquote adds a blockquote to the markdown.
func (b *Builder) Blockquote(text string) *Builder {
	lines := strings.Split(text, "\n")

	var newLines []string
	for _, l := range lines {
		newLines = append(newLines, strings.TrimSpace(">  "+l))
	}
	content := strings.Join(newLines, "\n")

	return b.P(content)
}

// BulletedList adds the given list as a bulleted
// list to the markdown.
func (b *Builder) BulletedList(list ...interface{}) *Builder {
	for _, el := range list {
		lines := strings.Split(fmt.Sprintf("%s", el), "\n")
		for i, l := range lines {
			if i == 0 {
				b.writelnf("* %s", l)
			} else {
				b.writelnf("  %s", l)
			}
		}
	}
	return b.nl()
}

// NumberedList adds the given list as a numbered
// list to the markdown.
func (b *Builder) NumberedList(list ...interface{}) *Builder {
	for i, el := range list {
		lines := strings.Split(fmt.Sprintf("%s", el), "\n")
		for j, l := range lines {
			if j == 0 {
				b.writelnf("%d. %s", i+1, l)
			} else {
				b.writelnf("   %s", l)
			}
		}
	}
	return b.nl()
}

const (
	minCellSize  = 3
	defaultAlign = AlignLeft
)

// Table adds a table from the given two-dimensional interface
// array to the markdown content.
// The first row of the array is considered as the table header.
// Align represents the alignment of each column, left to right.
// If the alignment of a column is not defined, the rigth alignment
// will be used instead as default.
func (b *Builder) Table(table [][]string, align []TableAlignment) *Builder {
	// The table must have at least two line ;
	// the header and one content line.
	if len(table) < 2 {
		return b
	}
	// This functions protects accesses to undefined
	// indexes in the given alignment slice. The default
	// alignment is returned if the given index exceed
	// the size of the slice.
	idxAlign := func(idx int) TableAlignment {
		alen := len(align)
		if alen == 0 || idx > alen-1 {
			return defaultAlign
		}
		return align[idx]
	}
	headerLen := len(table[0])

	// Count max char len of each column.
	colsLens := make([]int, headerLen)
	for x, line := range table {
		for y := range line {
			cl := max(charLen(table[x][y]), colsLens[y])
			if cl < minCellSize {
				colsLens[y] = minCellSize
			} else {
				colsLens[y] = cl
			}
		}
	}
	sb := strings.Builder{}

	for x, line := range table {
		// Add the line separator just after
		// the header line.
		if x == 1 {
			var cols []string
			for j, l := range colsLens {
				switch idxAlign(j) {
				case AlignRight:
					cols = append(cols, fmt.Sprintf("%s:", strings.Repeat("-", l-1)))
				case AlignCenter:
					cols = append(cols, fmt.Sprintf(":%s:", strings.Repeat("-", l-2)))
				case AlignLeft: // default
					cols = append(cols, strings.Repeat("-", l))
				}
			}
			sb.WriteString(joinCols(cols))
			sb.WriteByte('\n')
		}
		// Pad all the columns of the rows and
		// join the values with the separator.
		cols := make([]string, len(table[0]))
		for y, cl := range colsLens {
			var cell string
			if y <= len(line)-1 {
				cell = line[y]
			}
			if cell == "" {
				// Fill empty column with spaces.
				cols[y] = strings.Repeat(" ", cl)
			} else {
				cell = reCRLN.ReplaceAllString(cell, " ")
				cols[y] = padSpaces(cell, cl, idxAlign(y))
			}
		}
		sb.WriteString(joinCols(cols))
		sb.WriteByte('\n')
	}
	b.writeln(sb.String())

	return b
}

// write writes s to the markdown.
func (b *Builder) write(s string) *Builder {
	b.markdown += s
	return b
}

// writeln writes s with a new line to the markdown.
func (b *Builder) writeln(s string) *Builder {
	return b.write(s).nl()
}

// writelnf writes s with a new line to the markdown.
func (b *Builder) writelnf(format string, a ...interface{}) *Builder {
	return b.write(fmt.Sprintf(format, a...)).nl()
}

// charLen returns the number of caracters present in
// the string s. A caracter is defined as:
// - a sequence of runes that starts with a starter.
// - a rune that does not modify or combine backwards with
//   any other rune.
// - followed by possibly empty sequence of non-starters,
//   that is, runes that do (typically accents).
func charLen(s string) int {
	var (
		ia  norm.Iter
		len int
	)
	ia.InitString(norm.NFKD, s)
	for !ia.Done() {
		len++
		ia.Next()
	}
	return len
}

// max returns the larger of x or y.
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// joinCols join the columns of a table line with
// the pipe caracter as separator.
func joinCols(cols []string) string {
	return fmt.Sprintf("| %s |", strings.Join(cols, " | "))
}

// padSpaces pads the string s with spaces
// for the given pad length and alignement.
func padSpaces(s string, length int, align TableAlignment) string {
	sl := charLen(s)

	// Ensure that padding size is equal or greater
	// than the length of the string to pad to avoid
	// a panic later due to a negative repeat count.
	if sl > length {
		return s
	}
	switch align {
	case AlignRight:
		return strings.Repeat(" ", length-sl) + s
	case AlignCenter:
		len := float64(length-sl) / float64(2)
		pad := strings.Repeat(" ", int(math.Ceil(len/float64(1))))
		return pad[:int(math.Floor(float64(len)))] + s + pad[:int(math.Ceil(float64(len)))]
	default:
		// AlignLeft.
		return s + strings.Repeat(" ", length-sl)
	}
}
