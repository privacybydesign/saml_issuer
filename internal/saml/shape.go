package saml

import (
	"fmt"

	"github.com/beevik/etree"
	saml2 "github.com/russellhaering/gosaml2"
)

const (
	dsigNS   = "http://www.w3.org/2000/09/xmldsig#"
	maxDepth = 32

	tagResponse           = "Response"
	tagAssertion          = "Assertion"
	tagEncryptedAssertion = "EncryptedAssertion"
	tagSignature          = "Signature"
	tagSignedInfo         = "SignedInfo"
	tagSignatureValue     = "SignatureValue"
	tagKeyInfo            = "KeyInfo"
	tagObject             = "Object"
	tagReference          = "Reference"
	attrID                = "ID"
	attrURI               = "URI"
)

// shape is what checkShape learns from the unverified document.
type shape struct {
	assertionID string
}

// inventory lists the elements checkShape cares about, in document order.
type inventory struct {
	ids        map[string]struct{}
	assertions []*etree.Element
	signatures []*etree.Element
	encrypted  int
}

// checkShape rejects documents that do not have the one layout the signature
// checks assume: a Response holding exactly one Assertion that carries its own
// enveloped signature, referencing that Assertion's ID. Every XML signature
// wrapping variant needs a second Assertion, a second Signature, a duplicate
// ID or a ds:Object, so none of them gets as far as signature verification.
//
// The library decides what is verified. This only decides whether the
// document is worth handing to it.
func checkShape(raw []byte) (*shape, error) {
	doc := etree.NewDocument()
	doc.ReadSettings.MaxDepth = maxDepth
	if err := doc.ReadFromBytes(raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMalformed, err)
	}

	for _, token := range doc.Child {
		if _, ok := token.(*etree.Directive); ok {
			return nil, fmt.Errorf("%w: DTD or other directive", ErrMalformed)
		}
	}

	root := doc.Root()
	if !isElement(root, saml2.SAMLProtocolNamespace, tagResponse) {
		return nil, fmt.Errorf("%w: root is not a SAML Response", ErrMalformed)
	}

	inv := &inventory{ids: map[string]struct{}{}}
	if err := inv.collect(root); err != nil {
		return nil, err
	}

	return inv.validate(root)
}

func (inv *inventory) collect(el *etree.Element) error {
	if id := el.SelectAttrValue(attrID, ""); id != "" {
		if _, dup := inv.ids[id]; dup {
			return fmt.Errorf("%w: duplicate ID %q", ErrMalformed, id)
		}
		inv.ids[id] = struct{}{}
	}

	switch {
	case isElement(el, saml2.SAMLAssertionNamespace, tagAssertion):
		inv.assertions = append(inv.assertions, el)
	case isElement(el, saml2.SAMLAssertionNamespace, tagEncryptedAssertion):
		inv.encrypted++
	case isElement(el, dsigNS, tagSignature):
		inv.signatures = append(inv.signatures, el)
	case isElement(el, dsigNS, tagObject):
		return fmt.Errorf("%w: ds:Object", ErrMalformed)
	}

	for _, token := range el.Child {
		switch t := token.(type) {
		case *etree.Comment:
			// Parsers disagree on where a comment splits a text node.
			return fmt.Errorf("%w: XML comment", ErrMalformed)
		case *etree.Element:
			if err := inv.collect(t); err != nil {
				return err
			}
		}
	}

	return nil
}

func (inv *inventory) validate(root *etree.Element) (*shape, error) {
	if inv.encrypted > 0 {
		return nil, fmt.Errorf("%w: encrypted assertions are not supported", ErrMalformed)
	}
	if len(inv.assertions) != 1 {
		return nil, fmt.Errorf("%w: %d assertions, want 1", ErrMalformed, len(inv.assertions))
	}

	assertion := inv.assertions[0]
	if assertion.Parent() != root {
		return nil, fmt.Errorf("%w: assertion is not a child of the response", ErrMalformed)
	}

	assertionID := assertion.SelectAttrValue(attrID, "")
	if assertionID == "" {
		return nil, fmt.Errorf("%w: assertion has no ID", ErrMalformed)
	}

	var assertionSigs, responseSigs []*etree.Element
	for _, sig := range inv.signatures {
		switch sig.Parent() {
		case assertion:
			assertionSigs = append(assertionSigs, sig)
		case root:
			responseSigs = append(responseSigs, sig)
		default:
			return nil, fmt.Errorf("%w: signature in an unexpected place", ErrMalformed)
		}
	}

	if len(assertionSigs) == 0 {
		return nil, ErrUnsigned
	}
	if len(assertionSigs) > 1 || len(responseSigs) > 1 {
		return nil, fmt.Errorf("%w: more than one signature on an element", ErrMalformed)
	}

	if err := checkSignature(assertionSigs[0], assertionID); err != nil {
		return nil, err
	}

	if len(responseSigs) == 1 {
		if err := checkSignature(responseSigs[0], root.SelectAttrValue(attrID, "")); err != nil {
			return nil, err
		}
	}

	return &shape{assertionID: assertionID}, nil
}

// checkSignature requires a plain enveloped signature with a single Reference
// to the element it sits in.
func checkSignature(sig *etree.Element, signedID string) error {
	if signedID == "" {
		return fmt.Errorf("%w: signed element has no ID", ErrMalformed)
	}

	counts := map[string]int{}
	var signedInfo *etree.Element
	for _, child := range sig.ChildElements() {
		if child.NamespaceURI() != dsigNS {
			return fmt.Errorf("%w: foreign element in signature", ErrMalformed)
		}

		switch child.Tag {
		case tagSignedInfo:
			signedInfo = child
		case tagSignatureValue, tagKeyInfo:
		default:
			return fmt.Errorf("%w: %s in signature", ErrMalformed, child.Tag)
		}
		counts[child.Tag]++
	}

	if counts[tagSignedInfo] != 1 || counts[tagSignatureValue] != 1 || counts[tagKeyInfo] > 1 {
		return fmt.Errorf("%w: unexpected signature layout", ErrMalformed)
	}

	var refs []*etree.Element
	for _, child := range signedInfo.ChildElements() {
		if isElement(child, dsigNS, tagReference) {
			refs = append(refs, child)
		}
	}
	if len(refs) != 1 || refs[0].SelectAttrValue(attrURI, "") != "#"+signedID {
		return fmt.Errorf("%w: signature does not reference %s", ErrMalformed, signedID)
	}

	return nil
}

func isElement(el *etree.Element, namespace, tag string) bool {
	return el != nil && el.Tag == tag && el.NamespaceURI() == namespace
}
