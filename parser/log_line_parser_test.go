package parser

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestLogLineParser(t *testing.T) { TestingT(t) }

type LogLineParserSuite struct{}

var _ = Suite(&LogLineParserSuite{})

func (s LogLineParserSuite) TestParseSpacedFormat_Success(c *C) {
	data, err := ParseSpacedFormat(
		`$var1 [$var2] "$var3" ($var4) $var5`,
		`10 [text] "tested text that has [parentheses] (and) spaces" (text in parentheses) simple`,
	)

	c.Assert(err, IsNil)
	c.Assert(len(data), Equals, 5)
	c.Assert(data["$var1"], Equals, "10")
	c.Assert(data["$var2"], Equals, "text")
	c.Assert(data["$var3"], Equals, "tested text that has [parentheses] (and) spaces")
	c.Assert(data["$var4"], Equals, "text in parentheses")
	c.Assert(data["$var5"], Equals, "simple")
}

func (s LogLineParserSuite) TestParseSpacedFormat_Fail(c *C) {
	data, err := ParseSpacedFormat(`$var1 $var3`, `1 2 3`)

	c.Assert(data, IsNil)
	c.Assert(err, NotNil)
}

func (s LogLineParserSuite) TestParsePipedFormat_Success(c *C) {
	data, err := ParsePipedFormat(
		`$var1 | [$var2] | "$var3" | ($var4) | $var5`,
		`10 | [text] | "tested text that has [parentheses] (and) spaces" | (text in parentheses) | simple`,
	)

	c.Assert(err, IsNil)
	c.Assert(len(data), Equals, 5)
	c.Assert(data["$var1"], Equals, "10")
	c.Assert(data["$var2"], Equals, "text")
	c.Assert(data["$var3"], Equals, "tested text that has [parentheses] (and) spaces")
	c.Assert(data["$var4"], Equals, "text in parentheses")
	c.Assert(data["$var5"], Equals, "simple")
}

func (s LogLineParserSuite) TestParsePipedFormat_Fail(c *C) {
	data, err := ParsePipedFormat(`$var1 | $var3`, `1 | 2 | 3`)

	c.Assert(data, IsNil)
	c.Assert(err, NotNil)
}
