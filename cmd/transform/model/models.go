// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transmodel

type Transform struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func (t Transform) TransformToColumns() []string {
	return []string{t.ID, t.Name}
}

var TransformColumns = []string{"ID", "Name"}

type AttributesOfAccount struct {
	AttributeName string `json:"attributeName"`
	SourceName    string `json:"sourceName"`
}

type AccountAttribute struct {
	Type       string              `json:"type"`
	Attributes AttributesOfAccount `json:"attributes"`
}

type AttributesOfReference struct {
	Id    string           `json:"id"`
	Input AccountAttribute `json:"input"`
}

type Reference struct {
	Type       string                `json:"type"`
	Attributes AttributesOfReference `json:"attributes"`
}

type TransformDefinition struct {
	Type       string      `json:"type"`
	Attributes interface{} `json:"attributes"`
}

type AttributeTransform struct {
	IdentityAttributeName string              `json:"identityAttributeName"`
	TransformDefinition   TransformDefinition `json:"transformDefinition"`
}

type AttributeTransformPreview struct {
	AttributeName string                `json:"attributeName"`
	Attributes    AttributesOfReference `json:"attributes"`
	Type          string                `json:"type"`
}

type PreviewBodyImplicit struct {
	AttributeTransforms []AttributeTransformPreview `json:"attributeTransforms"`
}

type PreviewBodyExplicit struct {
	AttributeTransforms []map[string]interface{} `json:"attributeTransforms"`
}

type IdentityAttributeConfig struct {
	AttributeTransforms []AttributeTransform `json:"attributeTransforms"`
}

type ObjectRef struct {
	Type string `json:"type"`
	Id   string `json:"id"`
	Name string `json:"name"`
}
type IdentityProfile struct {
	AuthoritativeSource     ObjectRef               `json:"authoritativeSource"`
	IdentityAttributeConfig IdentityAttributeConfig `json:"identityAttributeConfig"`
}

type PreviewAttribute struct {
	Name          string `json:"name"`
	PreviousValue string `json:"previousValue"`
	Value         string `json:"value"`
}
type PreviewResponse struct {
	PreviewAttributes []PreviewAttribute `json:"previewAttributes"`
}

type User struct {
	Id string `json:"id"`
}

func MakeAttributesOfAccount(data interface{}) AttributesOfAccount {
	m := data.(map[string]interface{})
	attribute := AttributesOfAccount{}
	attribute.AttributeName = m["attributeName"].(string)
	attribute.SourceName = m["sourceName"].(string)
	return attribute
}

func MakeAccountAttribute(data interface{}) AccountAttribute {
	m := data.(map[string]interface{})
	account := AccountAttribute{}
	account.Type = m["type"].(string)
	account.Attributes = MakeAttributesOfAccount(m["attributes"])
	return account
}

func MakeReference(data interface{}) AttributesOfReference {
	m := data.(map[string]interface{})
	reference := AttributesOfReference{}
	reference.Id = m["id"].(string)
	reference.Input = MakeAccountAttribute(m["input"])
	return reference
}

func MakePreviewBodyImplicit(identityAttribute string, transformName string, accountAttribute string, sourceName string) PreviewBodyImplicit {
	attributeTransform := AttributeTransformPreview{}
	attributeTransform.AttributeName = identityAttribute
	attributeTransform.Attributes.Id = transformName
	attributeTransform.Attributes.Input.Type = "accountAttribute"
	attributeTransform.Attributes.Input.Attributes.AttributeName = accountAttribute
	attributeTransform.Attributes.Input.Attributes.SourceName = sourceName
	attributeTransform.Type = "reference"

	previewBody := PreviewBodyImplicit{}
	previewBody.AttributeTransforms = append(previewBody.AttributeTransforms, attributeTransform)

	return previewBody
}

func MakePreviewBodyExplicit(identityAttribute string, transformData map[string]interface{}) PreviewBodyExplicit {
	transformData["attributeName"] = identityAttribute

	previewBody := PreviewBodyExplicit{}
	previewBody.AttributeTransforms = append(previewBody.AttributeTransforms, transformData)

	return previewBody
}
