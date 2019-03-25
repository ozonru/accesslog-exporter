package net

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestIP(t *testing.T) { TestingT(t) }

type IPSuite struct{}

var _ = Suite(&IPSuite{})

func (s IPSuite) TestIsSubnetContainsIP(c *C) {
	isSubnet, err := IsSubnetContainsIP("30.2.2.1", []string{"30.0.0.0/8", "31.0.0.0/8"})
	c.Assert(err, IsNil)
	c.Assert(isSubnet, Equals, true)

	isSubnet, err = IsSubnetContainsIP("32.2.2.1", []string{"30.0.0.0/8", "31.0.0.0/8"})
	c.Assert(err, IsNil)
	c.Assert(isSubnet, Equals, false)

	isSubnet, err = IsSubnetContainsIP("32.2.2.1", []string{})
	c.Assert(err, IsNil)
	c.Assert(isSubnet, Equals, false)

	isSubnet, err = IsSubnetContainsIP("30.3.0.1", []string{"30.0.0.t/8", "31.0.0.0/8"})
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "invalid CIDR address: 30.0.0.t/8")
	c.Assert(isSubnet, Equals, false)
}
