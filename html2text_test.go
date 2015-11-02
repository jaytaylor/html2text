package html2text

import (
	"fmt"
	"regexp"
	"testing"
)

func TestStrippingWhitespace(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			"test text",
			"test text",
		},
		{
			"  \ttext\ntext\n",
			"text text",
		},
		{
			"  \na \n\t \n \n a \t",
			"a a",
		},
		{
			"test        text",
			"test text",
		},
		{
			"test&nbsp;&nbsp;&nbsp; text&nbsp;",
			"test    text",
		},
	}

	for _, testCase := range testCases {
		fmt.Printf("  testCase: <%s> <%s>\n", testCase.input, testCase.output)
		assertString(t, testCase.input, testCase.output)
		fmt.Printf("\n\n")
	}
}

func TestParagraphsAndBreaks(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			"Test text",
			"Test text",
		},
		{
			"Test text<br>",
			"Test text",
		},
		{
			"Test text<br>Test",
			"Test text\nTest",
		},
		{
			"<p>Test text</p>",
			"Test text",
		},
		{
			"<p>Test text</p><p>Test text</p>",
			"Test text\n\nTest text",
		},
		{
			"\n<p>Test text</p>\n\n\n\t<p>Test text</p>\n",
			"Test text\n\nTest text",
		},
		{
			"\n<p>Test text<br/>Test text</p>\n",
			"Test text\nTest text",
		},
		{
			"\n<p>Test text<br> \tTest text<br></p>\n",
			"Test text\nTest text",
		},
		{
			"Test text<br><BR />Test text",
			"Test text\n\nTest text",
		},
	}

	for _, testCase := range testCases {
		fmt.Printf("  testCase: <%s>\n", testCase.input)
		assertString(t, testCase.input, testCase.output)
		fmt.Printf("\n\n")
	}
}

func TestStrippingLists(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			"<ul></ul>",
			"",
		},
		{
			"<ul><li>item</li></ul>_",
			"* item\n\n_",
		},
		{
			"<li class='123'>item 1</li> <li>item 2</li>\n_",
			"* item 1\n* item 2\n_",
		},
		{
			"<li>item 1</li> \t\n <li>item 2</li> <li> item 3</li>\n_",
			"* item 1\n* item 2\n* item 3\n_",
		},
	}

	for _, testCase := range testCases {
		fmt.Printf("  testCase: <%s> <%s>\n", testCase.input, testCase.output)
		assertString(t, testCase.input, testCase.output)
		fmt.Printf("\n\n")
	}
}

func TestLinks(t *testing.T) {
	testCases := []struct {
		input  string
		output string
	}{
		{
			`<a></a>`,
			``,
		},
		{
			`<a href=""></a>`,
			``,
		},
		{
			`<a href="http://example.com/"></a>`,
			`( http://example.com/ )`,
		},
		{
			`<a href="">Link</a>`,
			`Link`,
		},
		{
			`<a href="http://example.com/">Link</a>`,
			`Link ( http://example.com/ )`,
		},
		{
			`<a href="http://example.com/"><span class="a">Link</span></a>`,
			`Link ( http://example.com/ )`,
		},
		{
			"<a href='http://example.com/'>\n\t<span class='a'>Link</span>\n\t</a>",
			`Link ( http://example.com/ )`,
		},
		{
			"<a href='mailto:contact@example.org'>Contact Us</a>",
			`Contact Us ( contact@example.org )`,
		},
		{
			"<a href=\"http://example.com:80/~user?aaa=bb&amp;c=d,e,f#foo\">Link</a>",
			`Link ( http://example.com:80/~user?aaa=bb&c=d,e,f#foo )`,
		},
		{
			"<a title='title' href=\"http://example.com/\">Link</a>",
			`Link ( http://example.com/ )`,
		},
		{
			"<a href=\"   http://example.com/ \"> Link </a>",
			`Link ( http://example.com/ )`,
		},
		{
			"<a href=\"http://example.com/a/\">Link A</a> <a href=\"http://example.com/b/\">Link B</a>",
			`Link A ( http://example.com/a/ ) Link B ( http://example.com/b/ )`,
		},
		{
			"<a href=\"%%LINK%%\">Link</a>",
			`Link ( %%LINK%% )`,
		},
		{
			"<a href=\"[LINK]\">Link</a>",
			`Link ( [LINK] )`,
		},
		{
			"<a href=\"{LINK}\">Link</a>",
			`Link ( {LINK} )`,
		},
		{
			"<a href=\"[[!unsubscribe]]\">Link</a>",
			`Link ( [[!unsubscribe]] )`,
		},
		{
			"<p>This is <a href=\"http://www.google.com\" >link1</a> and <a href=\"http://www.google.com\" >link2 </a> is next.</p>",
			`This is link1 ( http://www.google.com ) and link2 ( http://www.google.com ) is next.`,
		},
	}

	for _, testCase := range testCases {
		fmt.Printf("  testCase: <%s> <%s>\n", testCase.input, testCase.output)
		assertString(t, testCase.input, testCase.output)
		fmt.Printf("\n\n")
	}
}

func TestText(t *testing.T) {
	testCases := []struct {
		input string
		expr  string
	}{
		{
			`<li>
		  <a href="/new" data-ga-click="Header, create new repository, icon:repo"><span class="octicon octicon-repo"></span> New repository</a>
		</li>`,
			`\* New repository \( /new \)`,
		},
		{
			`hi

			<br>
	
	hello <a href="https://google.com">google</a>
	<br><br>
	test<p>List:</p>

	<ul>
		<li><a href="foo">Foo</a></li>
		<li><a href="http://www.microshwhat.com/bar/soapy">Barsoap</a></li>
        <li>Baz</li>
	</ul>
`,
			`hi
hello google \( https://google.com \)

test

List:

\* Foo \( foo \)
\* Barsoap \( http://www.microshwhat.com/bar/soapy \)
\* Baz`,
		},
		// Malformed input html.
		{
			`hi

			hello <a href="https://google.com">google</a>

			test<p>List:</p>

			<ul>
				<li><a href="foo">Foo</a>
				<li><a href="/
		                bar/baz">Bar</a>
		        <li>Baz</li>
			</ul>
		`,
			`hi hello google \( https://google.com \) test

List:

\* Foo \( foo \)
\* Bar \( /\n[ \t]+bar/baz \)
\* Baz`,
		},
	}

	for _, testCase := range testCases {
		assertRegexp(t, testCase.input, testCase.expr)
	}
}

type StringMatcher interface {
	MatchString(string) bool
	String() string
}

type RegexpStringMatcher string

func (m RegexpStringMatcher) MatchString(str string) bool {
	return regexp.MustCompile(string(m)).MatchString(str)
}
func (m RegexpStringMatcher) String() string {
	return string(m)
}

type ExactStringMatcher string

func (m ExactStringMatcher) MatchString(str string) bool {
	return string(m) == str
}
func (m ExactStringMatcher) String() string {
	return string(m)
}

func assertRegexp(t *testing.T, input string, outputRE string) {
	assertPlaintext(t, input, RegexpStringMatcher(outputRE))
}

func assertString(t *testing.T, input string, output string) {
	assertPlaintext(t, input, ExactStringMatcher(output))
}

func assertPlaintext(t *testing.T, input string, matcher StringMatcher) {
	text, err := FromString(input)
	if err != nil {
		t.Error(err)
	}

	if !matcher.MatchString(text) {
		t.Errorf("Input did not match expression\n"+
			"Input:\n>>>>\n%s\n<<<<\n\n"+
			"Output:\n>>>>\n%s\n<<<<\n\n"+
			"Expected output:\n>>>>\n%s\n<<<<\n\n",
			input, text, matcher.String())
	} else {
		t.Logf("input:\n\n%s\n\n\n\noutput:\n\n%s\n", input, text)
	}
}
