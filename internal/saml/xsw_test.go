package saml_test

import (
	"testing"

	"github.com/beevik/etree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/privacybydesign/saml_issuer/internal/saml"
	"github.com/privacybydesign/saml_issuer/internal/samltest"
)

const (
	evilID   = "_evil"
	evilUser = "admin"
)

// xswCase tampers with a valid response the way the XSW1 to XSW8 attacks do:
// the attacker's own Assertion is meant to be read, a copy of the signed one
// is there to satisfy signature verification.
type xswCase struct {
	name    string
	signing samltest.Signing
	attack  func(m *samltest.Message)
}

func TestSignatureWrappingAttacks(t *testing.T) {
	cases := []xswCase{
		{"XSW1 clone of the signed Response after its signature", samltest.SignResponseAndAssertion, xsw1},
		{"XSW2 clone of the signed Response before its signature", samltest.SignResponseAndAssertion, xsw2},
		{"XSW3 unsigned Assertion before the signed one", samltest.SignAssertion, xsw3},
		{"XSW4 signed Assertion inside an unsigned one", samltest.SignAssertion, xsw4},
		{"XSW5 unsigned original after the modified signed Assertion", samltest.SignAssertion, xsw5},
		{"XSW6 unsigned original inside the modified signed Assertion", samltest.SignAssertion, xsw6},
		{"XSW7 unsigned Assertion in Extensions", samltest.SignAssertion, xsw7},
		{"XSW8 unsigned original in a ds:Object of the signature", samltest.SignAssertion, xsw8},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			opts := f.options()
			opts.Signing = tt.signing
			m := f.build(opts)

			_, err := f.validate(m)
			require.NoError(t, err, "the untouched response must be valid")

			m = f.build(opts)
			tt.attack(m)
			attributes, err := f.validate(m)

			require.ErrorIs(t, err, saml.ErrMalformed)
			assert.Nil(t, attributes)
		})
	}
}

// unsignedEvilCopy is an Assertion with the attacker's values and no signature.
func unsignedEvilCopy(m *samltest.Message) *etree.Element {
	evil := m.Assertion.Copy()
	samltest.RemoveSignature(evil)
	evil.CreateAttr("ID", evilID)
	samltest.SetAttribute(evil, uidAttr, evilUser)

	return evil
}

// unsignedOriginal is the signed Assertion as it was, without its signature.
func unsignedOriginal(m *samltest.Message) *etree.Element {
	original := m.Assertion.Copy()
	samltest.RemoveSignature(original)

	return original
}

func childIndex(parent, child *etree.Element) int {
	for i, token := range parent.Child {
		if token == child {
			return i
		}
	}

	panic("not a child")
}

func signatureOf(el *etree.Element) *etree.Element { return el.FindElement("./Signature") }

func xsw1(m *samltest.Message) {
	clone := m.Response.Copy()
	samltest.RemoveSignature(clone)

	m.Response.CreateAttr("ID", evilID)
	samltest.SetAttribute(m.Assertion, uidAttr, evilUser)
	m.Response.InsertChildAt(childIndex(m.Response, signatureOf(m.Response))+1, clone)
}

func xsw2(m *samltest.Message) {
	clone := m.Response.Copy()
	samltest.RemoveSignature(clone)

	m.Response.CreateAttr("ID", evilID)
	samltest.SetAttribute(m.Assertion, uidAttr, evilUser)
	m.Response.InsertChildAt(childIndex(m.Response, signatureOf(m.Response)), clone)
}

func xsw3(m *samltest.Message) {
	m.Response.InsertChildAt(childIndex(m.Response, m.Assertion), unsignedEvilCopy(m))
}

func xsw4(m *samltest.Message) {
	evil := unsignedEvilCopy(m)

	m.Response.RemoveChild(m.Assertion)
	evil.AddChild(m.Assertion)
	m.Response.AddChild(evil)
}

func xsw5(m *samltest.Message) {
	original := unsignedOriginal(m)

	samltest.SetAttribute(m.Assertion, uidAttr, evilUser)
	m.Response.AddChild(original)
}

func xsw6(m *samltest.Message) {
	original := unsignedOriginal(m)

	samltest.SetAttribute(m.Assertion, uidAttr, evilUser)
	m.Assertion.InsertChildAt(childIndex(m.Assertion, signatureOf(m.Assertion))+1, original)
}

func xsw7(m *samltest.Message) {
	extensions := etree.NewElement("samlp:Extensions")
	extensions.AddChild(unsignedEvilCopy(m))

	m.Response.InsertChildAt(1, extensions)
}

func xsw8(m *samltest.Message) {
	original := unsignedOriginal(m)

	m.Assertion.CreateAttr("ID", evilID)
	samltest.SetAttribute(m.Assertion, uidAttr, evilUser)
	signatureOf(m.Assertion).CreateElement("ds:Object").AddChild(original)
}
