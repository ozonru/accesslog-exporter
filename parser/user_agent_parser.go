package parser

import "github.com/ua-parser/uap-go/uaparser"

// UserAgentParser is an interface of user agent parser
type UserAgentParser interface {
	Parse(line string) *uaparser.Client
}

// UAParser is a parser that uses uaparser package to parse user agent
type UAParser struct {
	parser *uaparser.Parser
}

func NewUAParser(regexPath string) (UserAgentParser, error) {
	parser, err := uaparser.NewWithOptions(
		regexPath,
		uaparser.EOsLookUpMode|uaparser.EDeviceLookUpMode|uaparser.EUserAgentLookUpMode,
		500000,
		20,
		false,
		false,
	)
	if err != nil {
		return nil, err
	}

	return &UAParser{parser: parser}, nil
}

func (p *UAParser) Parse(line string) *uaparser.Client {
	return p.parser.Parse(line)
}
