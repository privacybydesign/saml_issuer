package samltest

import (
	"crypto/x509"
	"encoding/base64"
	"time"

	"github.com/beevik/etree"
)

// Signing selects which elements of a response are signed.
type Signing int

const (
	SignAssertion Signing = iota
	SignResponseAndAssertion
	SignResponseOnly
	SignNothing
)

// KeyInfo selects what the signature's KeyInfo carries.
type KeyInfo int

const (
	KeyInfoSigner KeyInfo = iota
	KeyInfoOmit
	KeyInfoOther
)

// Attribute is one SAML attribute with its values.
type Attribute struct {
	Name   string
	Values []string
}

// Options describe a response. NewOptions returns one that is valid for the
// service provider it was built for; tests break one field at a time.
type Options struct {
	Issuer string
	// AssertionIssuer defaults to Issuer.
	AssertionIssuer string
	Destination     string
	Recipient       string
	Audience        string
	InResponseTo    string
	Status          string

	NotBefore                time.Time
	NotOnOrAfter             time.Time
	ConfirmationNotOnOrAfter time.Time

	Attributes []Attribute

	Signing Signing
	// SignedBy signs the response instead of the IdP that builds it.
	SignedBy *IdP
	KeyInfo  KeyInfo
	// KeyInfoCert is the certificate carried when KeyInfo is KeyInfoOther.
	KeyInfoCert *x509.Certificate
}

// NewOptions returns options for a response from idp to the service provider
// with the given entity ID and ACS URL, answering requestID.
func NewOptions(idp *IdP, entityID, acsURL, requestID string) Options {
	return Options{
		Issuer:                   idp.EntityID,
		Destination:              acsURL,
		Recipient:                acsURL,
		Audience:                 entityID,
		InResponseTo:             requestID,
		Status:                   StatusSuccess,
		NotBefore:                Now.Add(-time.Minute),
		NotOnOrAfter:             Now.Add(validWindow),
		ConfirmationNotOnOrAfter: Now.Add(validWindow),
		Attributes: []Attribute{
			{Name: "urn:mace:dir:attribute-def:uid", Values: []string{"jdoe"}},
			{Name: "urn:mace:dir:attribute-def:mail", Values: []string{"j.doe@example.org"}},
		},
	}
}

// Message is a built response and the elements tests tamper with.
type Message struct {
	Doc       *etree.Document
	Response  *etree.Element
	Assertion *etree.Element
}

// Build returns the response described by o, issued by i.
func (i *IdP) Build(o Options) *Message {
	response := etree.NewElement("samlp:Response")
	response.CreateAttr("xmlns:samlp", protocolNS)
	response.CreateAttr("xmlns:saml", assertionNS)
	response.CreateAttr("ID", ResponseID)
	response.CreateAttr("Version", samlVersion)
	response.CreateAttr("IssueInstant", formatTime(Now))
	response.CreateAttr("Destination", o.Destination)
	response.CreateAttr("InResponseTo", o.InResponseTo)
	response.CreateElement("saml:Issuer").SetText(o.Issuer)

	status := response.CreateElement("samlp:Status")
	status.CreateElement("samlp:StatusCode").CreateAttr("Value", o.Status)

	assertion := buildAssertion(response, o)

	signer := i
	if o.SignedBy != nil {
		signer = o.SignedBy
	}
	keyInfoCert := keyInfoCertFor(signer, o)

	switch o.Signing {
	case SignAssertion:
		signer.signWith(assertion, keyInfoCert)
	case SignResponseAndAssertion:
		signer.signWith(assertion, keyInfoCert)
		signer.signWith(response, keyInfoCert)
	case SignResponseOnly:
		signer.signWith(response, keyInfoCert)
	case SignNothing:
	}

	doc := etree.NewDocument()
	doc.SetRoot(response)

	return &Message{Doc: doc, Response: response, Assertion: assertion}
}

func keyInfoCertFor(signer *IdP, o Options) *x509.Certificate {
	switch o.KeyInfo {
	case KeyInfoOmit:
		return nil
	case KeyInfoOther:
		return o.KeyInfoCert
	default:
		return signer.Cert
	}
}

func buildAssertion(response *etree.Element, o Options) *etree.Element {
	assertion := response.CreateElement("saml:Assertion")
	assertion.CreateAttr("xmlns:saml", assertionNS)
	assertion.CreateAttr("xmlns:xs", "http://www.w3.org/2001/XMLSchema")
	assertion.CreateAttr("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	assertion.CreateAttr("ID", AssertionID)
	assertion.CreateAttr("Version", samlVersion)
	assertion.CreateAttr("IssueInstant", formatTime(Now))
	assertion.CreateElement("saml:Issuer").SetText(cmpOr(o.AssertionIssuer, o.Issuer))

	subject := assertion.CreateElement("saml:Subject")
	nameID := subject.CreateElement("saml:NameID")
	nameID.CreateAttr("Format", "urn:oasis:names:tc:SAML:2.0:nameid-format:transient")
	nameID.SetText("_transient-name-id")
	confirmation := subject.CreateElement("saml:SubjectConfirmation")
	confirmation.CreateAttr("Method", "urn:oasis:names:tc:SAML:2.0:cm:bearer")
	data := confirmation.CreateElement("saml:SubjectConfirmationData")
	data.CreateAttr("NotOnOrAfter", formatTime(o.ConfirmationNotOnOrAfter))
	data.CreateAttr("Recipient", o.Recipient)
	data.CreateAttr("InResponseTo", o.InResponseTo)

	conditions := assertion.CreateElement("saml:Conditions")
	conditions.CreateAttr("NotBefore", formatTime(o.NotBefore))
	conditions.CreateAttr("NotOnOrAfter", formatTime(o.NotOnOrAfter))
	conditions.CreateElement("saml:AudienceRestriction").CreateElement("saml:Audience").SetText(o.Audience)

	authn := assertion.CreateElement("saml:AuthnStatement")
	authn.CreateAttr("AuthnInstant", formatTime(Now))
	authn.CreateElement("saml:AuthnContext").CreateElement("saml:AuthnContextClassRef").
		SetText("urn:oasis:names:tc:SAML:2.0:ac:classes:Password")

	statement := assertion.CreateElement("saml:AttributeStatement")
	for _, attribute := range o.Attributes {
		el := statement.CreateElement("saml:Attribute")
		el.CreateAttr("Name", attribute.Name)
		el.CreateAttr("NameFormat", attrNameForm)
		for _, value := range attribute.Values {
			v := el.CreateElement("saml:AttributeValue")
			v.CreateAttr("xsi:type", "xs:string")
			v.SetText(value)
		}
	}

	return assertion
}

func cmpOr(value, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

func formatTime(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05Z") }

// Bytes returns the serialised response.
func (m *Message) Bytes() []byte {
	out, err := m.Doc.WriteToBytes()
	if err != nil {
		panic(err)
	}

	return out
}

// Encode returns the response as an HTTP-POST SAMLResponse value.
func (m *Message) Encode() string { return base64.StdEncoding.EncodeToString(m.Bytes()) }

// RemoveSignature removes the signature that is a direct child of el.
func RemoveSignature(el *etree.Element) {
	for _, child := range el.ChildElements() {
		if child.Tag == "Signature" {
			el.RemoveChild(child)
		}
	}
}

// SetAttribute replaces the values of the named attribute with value.
func SetAttribute(assertion *etree.Element, name, value string) {
	for _, attribute := range assertion.FindElements("./AttributeStatement/Attribute") {
		if attribute.SelectAttrValue("Name", "") != name {
			continue
		}

		for _, v := range attribute.FindElements("./AttributeValue") {
			v.SetText(value)
		}
	}
}
