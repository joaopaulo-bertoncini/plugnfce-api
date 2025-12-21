package signer

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/beevik/etree"
	"golang.org/x/crypto/pkcs12"
)

// KeyMaterial represents the PFX bundle and its password.
type KeyMaterial struct {
	PFXBase64 string
	Password  string
}

// Signer encapsulates XMLDSig enveloped signature logic.
type Signer interface {
	SignEnveloped(ctx context.Context, unsignedXML []byte, key KeyMaterial, referenceID string) ([]byte, error)
}

// signer implements Signer interface
type signer struct{}

// NewSigner creates a new XML signer
func NewSigner() Signer {
	return &signer{}
}

// SignEnveloped signs XML with enveloped signature
func (s *signer) SignEnveloped(ctx context.Context, unsignedXML []byte, key KeyMaterial, referenceID string) ([]byte, error) {
	// Parse XML
	doc := etree.NewDocument()
	err := doc.ReadFromBytes(unsignedXML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	// Find the element to sign (usually infNFe)
	elementToSign := s.findElementByID(doc.Root(), referenceID)
	if elementToSign == nil {
		return nil, fmt.Errorf("element with ID %s not found", referenceID)
	}

	// Load certificate and private key
	cert, privateKey, err := s.loadCertificateAndKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	// Create signature
	signature, err := s.createSignature(elementToSign, cert, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signature: %w", err)
	}

	// Add signature to document
	doc.Root().AddChild(signature)

	// Return signed XML
	signedXML, err := doc.WriteToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize signed XML: %w", err)
	}

	return signedXML, nil
}

// findElementByID finds an element by its ID attribute
func (s *signer) findElementByID(element *etree.Element, id string) *etree.Element {
	if element.SelectAttr("Id") != nil && element.SelectAttr("Id").Value == id {
		return element
	}

	for _, child := range element.ChildElements() {
		if found := s.findElementByID(child, id); found != nil {
			return found
		}
	}

	return nil
}

// loadCertificateAndKey loads certificate and private key from PFX
func (s *signer) loadCertificateAndKey(key KeyMaterial) (*x509.Certificate, *rsa.PrivateKey, error) {
	// Decode base64 PFX
	pfxData, err := base64.StdEncoding.DecodeString(key.PFXBase64)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode PFX base64: %w", err)
	}

	// Parse PFX/P12
	privateKey, cert, err := pkcs12.Decode(pfxData, key.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse PFX: %w", err)
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("private key is not RSA")
	}

	return cert, rsaKey, nil
}

// createSignature creates the XMLDSig signature element
func (s *signer) createSignature(elementToSign *etree.Element, cert *x509.Certificate, privateKey *rsa.PrivateKey) (*etree.Element, error) {
	// Canonicalize the element
	canonicalized, err := s.canonicalize(elementToSign)
	if err != nil {
		return nil, fmt.Errorf("failed to canonicalize: %w", err)
	}

	// Calculate digest (SHA-256)
	digest := sha256.Sum256(canonicalized)
	digestBase64 := base64.StdEncoding.EncodeToString(digest[:])

	// Create SignedInfo
	signedInfo := s.createSignedInfo(elementToSign, digestBase64)

	// Canonicalize SignedInfo
	signedInfoCanonicalized, err := s.canonicalizeElement(signedInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to canonicalize SignedInfo: %w", err)
	}

	// Sign SignedInfo
	signatureValue, err := s.signData(signedInfoCanonicalized, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	// Create signature element
	signature := etree.NewElement("Signature")
	signature.CreateAttr("xmlns", "http://www.w3.org/2000/09/xmldsig#")

	signature.AddChild(signedInfo)
	signature.AddChild(s.createSignatureValue(signatureValue))
	signature.AddChild(s.createKeyInfo(cert))

	return signature, nil
}

// canonicalize performs C14N canonicalization
func (s *signer) canonicalize(element *etree.Element) ([]byte, error) {
	// For simplicity, we'll use a basic canonicalization
	// In production, you should use a proper C14N implementation
	var buf strings.Builder
	element.WriteTo(&buf, &etree.WriteSettings{
		CanonicalText:    true,
		CanonicalAttrVal: true,
	})

	xmlStr := buf.String()
	// Remove extra whitespace between tags
	re := regexp.MustCompile(`>\s+<`)
	xmlStr = re.ReplaceAllString(xmlStr, "><")
	// Trim spaces
	xmlStr = strings.TrimSpace(xmlStr)
	return []byte(xmlStr), nil
}

// canonicalizeElement converts element to canonicalized bytes
func (s *signer) canonicalizeElement(element *etree.Element) ([]byte, error) {
	return s.canonicalize(element)
}

// createSignedInfo creates the SignedInfo element
func (s *signer) createSignedInfo(elementToSign *etree.Element, digestBase64 string) *etree.Element {
	signedInfo := etree.NewElement("SignedInfo")

	// CanonicalizationMethod
	canonicalizationMethod := etree.NewElement("CanonicalizationMethod")
	canonicalizationMethod.CreateAttr("Algorithm", "http://www.w3.org/TR/2001/REC-xml-c14n-20010315")
	signedInfo.AddChild(canonicalizationMethod)

	// SignatureMethod
	signatureMethod := etree.NewElement("SignatureMethod")
	signatureMethod.CreateAttr("Algorithm", "http://www.w3.org/2000/09/xmldsig#rsa-sha256")
	signedInfo.AddChild(signatureMethod)

	// Reference
	reference := etree.NewElement("Reference")
	reference.CreateAttr("URI", "#"+elementToSign.SelectAttr("Id").Value)

	// Transforms
	transforms := etree.NewElement("Transforms")
	envelopedTransform := etree.NewElement("Transform")
	envelopedTransform.CreateAttr("Algorithm", "http://www.w3.org/2000/09/xmldsig#enveloped-signature")
	transforms.AddChild(envelopedTransform)

	canonicalTransform := etree.NewElement("Transform")
	canonicalTransform.CreateAttr("Algorithm", "http://www.w3.org/TR/2001/REC-xml-c14n-20010315")
	transforms.AddChild(canonicalTransform)

	reference.AddChild(transforms)

	// DigestMethod
	digestMethod := etree.NewElement("DigestMethod")
	digestMethod.CreateAttr("Algorithm", "http://www.w3.org/2001/04/xmlenc#sha256")
	reference.AddChild(digestMethod)

	// DigestValue
	digestValue := etree.NewElement("DigestValue")
	digestValue.SetText(digestBase64)
	reference.AddChild(digestValue)

	signedInfo.AddChild(reference)

	return signedInfo
}

// signData signs the data with RSA-SHA256
func (s *signer) signData(data []byte, privateKey *rsa.PrivateKey) (string, error) {
	hashed := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// createSignatureValue creates the SignatureValue element
func (s *signer) createSignatureValue(signature string) *etree.Element {
	signatureValue := etree.NewElement("SignatureValue")
	signatureValue.SetText(signature)
	return signatureValue
}

// createKeyInfo creates the KeyInfo element
func (s *signer) createKeyInfo(cert *x509.Certificate) *etree.Element {
	keyInfo := etree.NewElement("KeyInfo")
	x509Data := etree.NewElement("X509Data")
	x509Certificate := etree.NewElement("X509Certificate")
	x509Certificate.SetText(base64.StdEncoding.EncodeToString(cert.Raw))
	x509Data.AddChild(x509Certificate)
	keyInfo.AddChild(x509Data)
	return keyInfo
}
