package html2text

import (
	"regexp"
	"testing"
)

func TestText(t *testing.T) {
	testCases := []struct {
		input string
		expr  string
	}{
		{
			`<li>
  <a href="/new" data-ga-click="Header, create new repository, icon:repo"><span class="octicon octicon-repo"></span> New repository</a>
</li>`,
			`/new`,
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

hello
https://google.com
test

List:

foo
http://www.microshwhat.com/bar/soapy

Baz`,
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
			`hi hello
https://google.com
test

List:

foo
/\n[ \t]+bar/baz

Baz`,
		},
	}

	for _, testCase := range testCases {
		text, err := FromString(testCase.input)
		if err != nil {
			t.Error(err)
		}
		if expr := regexp.MustCompile(testCase.expr); !expr.MatchString(text) {
			t.Errorf("Input did not match expression\nInput:\n>>>>\n%s\n<<<<\n\nOutput:\n>>>>\n%s\n<<<<\n\nExpression: %s\n", testCase.input, text, expr.String())
		} else {
			t.Logf("input:\n\n%s\n\n\n\noutput:\n\n%s\n", testCase.input, text)
		}
	}
}
