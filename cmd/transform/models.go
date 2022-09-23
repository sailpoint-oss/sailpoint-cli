// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

type transform struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type operation struct {
}

func (t transform) transformToColumns() []string {
	return []string{t.ID, t.Name}
}

var transformColumns = []string{"ID", "Name"}

type attributesOfAccount struct {
	AttributeName string `json:"attributeName"`
	SourceName    string `json:"sourceName"`
}

type accountAttribute struct {
	Type       string              `json:"type"`
	Attributes attributesOfAccount `json:"attributes"`
}

type attributesOfReference struct {
	Id    string           `json:"id"`
	Input accountAttribute `json:"input"`
}

type reference struct {
	Type       string                `json:"type"`
	Attributes attributesOfReference `json:"attributes"`
}

type transformDefinition struct {
	Type       string      `json:"type"`
	Attributes interface{} `json:"attributes"`
}

type attributeTransform struct {
	IdentityAttributeName string              `json:"identityAttributeName"`
	TransformDefinition   transformDefinition `json:"transformDefinition"`
}

type attributeTransformPreview struct {
	AttributeName string                `json:"attributeName"`
	Attributes    attributesOfReference `json:"attributes"`
	Type          string                `json:"type"`
}

type previewBodyImplicit struct {
	AttributeTransforms []attributeTransformPreview `json:"attributeTransforms"`
}

type previewBodyExplicit struct {
	AttributeTransforms []map[string]interface{} `json:"attributeTransforms"`
}

type identityAttributeConfig struct {
	AttributeTransforms []attributeTransform `json:"attributeTransforms"`
}

type objectRef struct {
	Type string `json:"type"`
	Id   string `json:"id"`
	Name string `json:"name"`
}
type identityProfile struct {
	AuthoritativeSource     objectRef               `json:"authoritativeSource"`
	IdentityAttributeConfig identityAttributeConfig `json:"identityAttributeConfig"`
}

type previewAttribute struct {
	Name          string `json:"name"`
	PreviousValue string `json:"previousValue"`
	Value         string `json:"value"`
}
type previewResponse struct {
	PreviewAttributes []previewAttribute `json:"previewAttributes"`
}

type user struct {
	Id string `json:"id"`
}

func makeAttributesOfAccount(data interface{}) attributesOfAccount {
	m := data.(map[string]interface{})
	attribute := attributesOfAccount{}
	attribute.AttributeName = m["attributeName"].(string)
	attribute.SourceName = m["sourceName"].(string)
	return attribute
}

func makeAccountAttribute(data interface{}) accountAttribute {
	m := data.(map[string]interface{})
	account := accountAttribute{}
	account.Type = m["type"].(string)
	account.Attributes = makeAttributesOfAccount(m["attributes"])
	return account
}

func makeReference(data interface{}) attributesOfReference {
	m := data.(map[string]interface{})
	reference := attributesOfReference{}
	reference.Id = m["id"].(string)
	reference.Input = makeAccountAttribute(m["input"])
	return reference
}

func makePreviewBodyImplicit(identityAttribute string, transformName string, accountAttribute string, sourceName string) previewBodyImplicit {
	attributeTransform := attributeTransformPreview{}
	attributeTransform.AttributeName = identityAttribute
	attributeTransform.Attributes.Id = transformName
	attributeTransform.Attributes.Input.Type = "accountAttribute"
	attributeTransform.Attributes.Input.Attributes.AttributeName = accountAttribute
	attributeTransform.Attributes.Input.Attributes.SourceName = sourceName
	attributeTransform.Type = "reference"

	previewBody := previewBodyImplicit{}
	previewBody.AttributeTransforms = append(previewBody.AttributeTransforms, attributeTransform)

	return previewBody
}

func makePreviewBodyExplicit(identityAttribute string, transformData map[string]interface{}) previewBodyExplicit {
	transformData["attributeName"] = identityAttribute

	previewBody := previewBodyExplicit{}
	previewBody.AttributeTransforms = append(previewBody.AttributeTransforms, transformData)

	return previewBody
}
