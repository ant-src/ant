// Package htmldata lists the elements and attributes of HTML for completion in markup (section 5,
// HTML expressions). Ant does not validate HTML, so the lists are advisory: they follow the WHATWG
// HTML standard's elements and the attributes each takes, besides the global ones.
package htmldata

import (
	"sort"
	"strings"
)

// Element is an element with a one-line description and the attributes it takes besides the
// global ones.
type Element struct {
	Name  string
	Doc   string
	Attrs []string
}

// Global are the attributes every element takes.
var Global = strings.Fields(`accesskey autocapitalize autofocus class contenteditable dir draggable enterkeyhint
	hidden id inert inputmode is itemid itemprop itemref itemscope itemtype lang nonce popover slot spellcheck
	style tabindex title translate`)

// Elements are the elements, sorted by name.
var Elements = elements()

var byName = func() map[string]*Element {
	m := map[string]*Element{}
	for i := range Elements {
		m[Elements[i].Name] = &Elements[i]
	}
	return m
}()

// Lookup finds an element by name, matched without regard to case.
func Lookup(name string) *Element { return byName[strings.ToLower(name)] }

// Attributes lists the attributes an element takes, its own first and then the global ones;
// for an unknown element, the global ones. Each is listed once.
func Attributes(element string) []string {
	var out []string
	seen := map[string]bool{}
	if el := Lookup(element); el != nil {
		for _, a := range el.Attrs {
			if !seen[a] {
				seen[a] = true
				out = append(out, a)
			}
		}
	}
	for _, a := range Global {
		if !seen[a] {
			seen[a] = true
			out = append(out, a)
		}
	}
	return out
}

func elements() []Element {
	type e = Element
	a := func(s string) []string { return strings.Fields(s) }
	list := []Element{
		e{"a", "A hyperlink.", a("href target download rel hreflang type referrerpolicy ping")},
		e{"abbr", "An abbreviation; title holds the expansion.", nil},
		e{"address", "Contact information.", nil},
		e{"area", "A region of an image map.", a("alt coords shape href target download rel referrerpolicy ping")},
		e{"article", "A self-contained composition.", nil},
		e{"aside", "Content aside from the main content.", nil},
		e{"audio", "Sound content.", a("src crossorigin preload autoplay loop muted controls")},
		e{"b", "Text to draw attention to.", nil},
		e{"base", "The base URL of the document.", a("href target")},
		e{"bdi", "Text isolated for bidirectional formatting.", nil},
		e{"bdo", "Text with an explicit direction.", a("dir")},
		e{"blockquote", "A quotation.", a("cite")},
		e{"body", "The document's content.", nil},
		e{"br", "A line break.", nil},
		e{"button", "A button.", a("type name value disabled form formaction formenctype formmethod formnovalidate formtarget popovertarget popovertargetaction")},
		e{"canvas", "A drawing surface.", a("width height")},
		e{"caption", "A table's caption.", nil},
		e{"cite", "The title of a work.", nil},
		e{"code", "Computer code.", nil},
		e{"col", "A table column.", a("span")},
		e{"colgroup", "A group of table columns.", a("span")},
		e{"data", "Content with a machine-readable value.", a("value")},
		e{"datalist", "Options for an input.", nil},
		e{"dd", "A description in a description list.", nil},
		e{"del", "Removed text.", a("cite datetime")},
		e{"details", "A disclosure widget.", a("open name")},
		e{"dfn", "The defining instance of a term.", nil},
		e{"dialog", "A dialog box.", a("open")},
		e{"div", "A generic container.", nil},
		e{"dl", "A description list.", nil},
		e{"dt", "A term in a description list.", nil},
		e{"em", "Emphasized text.", nil},
		e{"embed", "An external resource.", a("src type width height")},
		e{"fieldset", "A set of form controls.", a("disabled form name")},
		e{"figcaption", "A figure's caption.", nil},
		e{"figure", "A figure with an optional caption.", nil},
		e{"footer", "A footer.", nil},
		e{"form", "A form.", a("action method enctype name target autocomplete novalidate accept-charset rel")},
		e{"h1", "A heading.", nil}, e{"h2", "A heading.", nil}, e{"h3", "A heading.", nil},
		e{"h4", "A heading.", nil}, e{"h5", "A heading.", nil}, e{"h6", "A heading.", nil},
		e{"head", "The document's metadata.", nil},
		e{"header", "A header.", nil},
		e{"hgroup", "A heading group.", nil},
		e{"hr", "A thematic break.", nil},
		e{"html", "The document's root.", a("lang")},
		e{"i", "Text in an alternate voice.", nil},
		e{"iframe", "A nested browsing context; written self-closing in Ant.", a("src srcdoc name sandbox allow allowfullscreen width height referrerpolicy loading")},
		e{"img", "An image.", a("src alt srcset sizes crossorigin usemap ismap width height referrerpolicy decoding loading fetchpriority")},
		e{"input", "A form control.", a("type name value disabled form readonly required placeholder autocomplete checked accept alt autofocus capture dirname formaction formenctype formmethod formnovalidate formtarget height width list max maxlength min minlength multiple pattern size src step popovertarget popovertargetaction")},
		e{"ins", "Inserted text.", a("cite datetime")},
		e{"kbd", "User input.", nil},
		e{"label", "A caption for a form control.", a("for form")},
		e{"legend", "A fieldset's caption.", nil},
		e{"li", "A list item.", a("value")},
		e{"link", "A link to a resource.", a("href rel as type media sizes crossorigin hreflang referrerpolicy integrity imagesrcset imagesizes blocking color disabled fetchpriority")},
		e{"main", "The main content.", nil},
		e{"map", "An image map.", a("name")},
		e{"mark", "Highlighted text.", nil},
		e{"math", "MathML content.", nil},
		e{"menu", "A toolbar or list of commands.", nil},
		e{"meta", "Metadata.", a("name content charset http-equiv media")},
		e{"meter", "A scalar measurement.", a("value min max low high optimum")},
		e{"nav", "Navigation links.", nil},
		e{"noscript", "Content for when scripting is off.", nil},
		e{"object", "An external resource.", a("data type name form width height")},
		e{"ol", "An ordered list.", a("reversed start type")},
		e{"optgroup", "A group of options.", a("disabled label")},
		e{"option", "An option in a select.", a("disabled label selected value")},
		e{"output", "The result of a calculation.", a("for form name")},
		e{"p", "A paragraph.", nil},
		e{"picture", "A set of image sources.", nil},
		e{"pre", "Preformatted text, kept as written.", nil},
		e{"progress", "The progress of a task.", a("value max")},
		e{"q", "An inline quotation.", a("cite")},
		e{"rp", "Fallback parentheses for ruby.", nil},
		e{"rt", "Ruby text.", nil},
		e{"ruby", "Ruby annotations.", nil},
		e{"s", "Text no longer accurate.", nil},
		e{"samp", "Sample output.", nil},
		e{"script", "A script; holds literal text up to its end tag.", a("src type async defer crossorigin integrity nomodule referrerpolicy blocking fetchpriority")},
		e{"search", "A search section.", nil},
		e{"section", "A section.", nil},
		e{"select", "A drop-down list.", a("name disabled form multiple required size autocomplete")},
		e{"slot", "A shadow tree slot.", a("name")},
		e{"small", "Side comments.", nil},
		e{"source", "A media or image source.", a("src type srcset sizes media width height")},
		e{"span", "A generic inline container.", nil},
		e{"strong", "Important text.", nil},
		e{"style", "A style sheet; holds literal text up to its end tag.", a("media blocking")},
		e{"sub", "A subscript.", nil},
		e{"summary", "A details element's summary.", nil},
		e{"sup", "A superscript.", nil},
		e{"svg", "SVG content.", a("width height viewBox xmlns fill stroke")},
		e{"table", "A table.", nil},
		e{"tbody", "A table body.", nil},
		e{"td", "A table cell.", a("colspan rowspan headers")},
		e{"template", "A fragment to clone.", a("shadowrootmode shadowrootdelegatesfocus")},
		e{"textarea", "A multi-line text control; its text is kept as written.", a("name rows cols disabled form readonly required placeholder autocomplete maxlength minlength wrap dirname")},
		e{"tfoot", "A table footer.", nil},
		e{"th", "A table header cell.", a("colspan rowspan headers scope abbr")},
		e{"thead", "A table header.", nil},
		e{"time", "A time.", a("datetime")},
		e{"title", "The document's title; holds text only.", nil},
		e{"tr", "A table row.", nil},
		e{"track", "A text track.", a("kind src srclang label default")},
		e{"u", "Unarticulated annotation.", nil},
		e{"ul", "An unordered list.", nil},
		e{"var", "A variable.", nil},
		e{"video", "Video content.", a("src crossorigin poster preload autoplay playsinline loop muted controls width height")},
		e{"wbr", "A line break opportunity.", nil},
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list
}
