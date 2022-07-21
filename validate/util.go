package validate

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/sailpoint/sp-cli/client"
)

// entitlementAttr returns the attribute for entitlements
func entitlementAttr(spec *client.ConnSpec) string {
	for _, attr := range spec.AccountSchema.Attributes {
		if attr.Entitlement {
			return attr.Name
		}
	}
	return ""
}

// accountEntitlements returns all entitlements on the account
func accountEntitlements(account *client.Account, spec *client.ConnSpec) ([]string, error) {
	entitlementAttr := entitlementAttr(spec)
	if entitlementAttr == "" {
		return nil, fmt.Errorf("no entitlement attr found")
	}
	entitlements := []string{}
	for _, identity := range account.Attributes[entitlementAttr].([]interface{}) {
		entitlements = append(entitlements, identity.(string))
	}
	return entitlements, nil
}

// accountHasEntitlement returns whether or not an account has a specific entitlement
func accountHasEntitlement(account *client.Account, spec *client.ConnSpec, entitlementID string) bool {
	entitlements, err := accountEntitlements(account, spec)
	if err != nil {
		panic(err.Error())
	}
	for _, id := range entitlements {
		if id == entitlementID {
			return true
		}
	}
	return false
}

// Diff is a difference between two values
type diff struct {
	Field string
	A     interface{}
	B     interface{}
}

// compareIntersection compares two objects and returns any differences between
// the fields that are common to both objects.
func compareIntersection(a map[string]interface{}, b map[string]interface{}) (diffs []diff) {
	for key := range a {
		if _, found := b[key]; !found {
			continue
		}

		switch v := b[key].(type) {
		case []interface{}:
			var sliceB []string
			for _, val := range v {
				sliceB = append(sliceB, val.(string))
			}

			sliceA, ok := a[key].([]string)
			if !ok {
				log.Println("failed to convert to sliceA to slice of strings")
			}

			for i := range sliceA {
				if sliceA[i] != sliceB[i] {
					diffs = append(diffs, diff{
						Field: key,
						A:     a[key],
						B:     b[key],
					})
				}
			}
		default:
			if a[key] != b[key] {
				diffs = append(diffs, diff{
					Field: key,
					A:     a[key],
					B:     b[key],
				})
			}
		}
	}
	return diffs
}

const (
	fieldTypeStatic    = "static"
	fieldTypeGenerator = "generator"
	generatorPassword  = "Create Password"
	generatorAccountId = "Create Unique Account ID"
)

// genCreateField generates a value for the provided account create template field
func genCreateField(field client.AccountCreateTemplateField) interface{} {

	// Return typed based value if the field is in deprecated format
	// TODO: Once we move away from the old format, this should also be removed
	if field.Key == "" {
		return genValueByTypeAndName(field)
	}

	// Return default value if field is set to static
	if field.InitialValue.Type == fieldTypeStatic {
		return field.InitialValue.Attributes.Value
	}

	// Build value for generator field
	if field.InitialValue.Type == fieldTypeGenerator {
		if field.InitialValue.Attributes.Name == generatorPassword {
			return fmt.Sprintf("RandomPassword.%d", rand.Intn(65536))
		}

		if field.InitialValue.Attributes.Name == generatorAccountId {
			template := field.InitialValue.Attributes.Template

			counterRegex := regexp.MustCompile(`\$\(uniqueCounter\)`)
			template = counterRegex.ReplaceAllString(template, strconv.Itoa(rand.Intn(65536)))

			stringRegex := regexp.MustCompile(`\$\(.*?\)`)
			template = stringRegex.ReplaceAllString(template, fmt.Sprintf("string%d", rand.Intn(99)))

			return template
		}
	}

	// For other cases including identity attributes, use the default way to generate value by type.
	return genValueByTypeAndName(field)
}

// getFieldName returns the name of the field
// TODO: This is to support both key and name base field. Once the name based filds are gone, we can remove this helper method
func getFieldName(field client.AccountCreateTemplateField) string {
	if field.Key == "" {
		return field.Name
	}
	return field.Key
}

// genValueByTypeAndName generates attribute values base on field type and name
func genValueByTypeAndName(field client.AccountCreateTemplateField) interface{} {
	switch field.Type {
	case "string":
		if getFieldName(field) == "email" || getFieldName(field) == "name" {
			return fmt.Sprintf("test.%d@example.com", rand.Intn(65536))
		} else if getFieldName(field) == "siteRole" {
			return "Creator"
		} else {
			return fmt.Sprintf("string.%d", rand.Intn(65536))
		}
	case "boolean":
		// TODO: we want to eventually remove these. These fields needs only for Smartsheet connectors
		if getFieldName(field) == "admin" {
			return false
		}
		if getFieldName(field) == "groupAdmin" {
			return false
		}
		if getFieldName(field) == "licensedSheetCreator" {
			return false
		}
		if getFieldName(field) == "resourceViewer" {
			return false
		}
		return true
	case "array":
		// TODO: we need to avoid hardcoding any specific code in the validation suite.
		// Freshservice connector only
		if getFieldName(field) == "roles" {
			return []string{"27000245813:entire_helpdesk"}
		}
		return nil
	default:
		panic(fmt.Sprintf("unknown type: %q", field.Type))
	}
}

// testSchema verifies that value is of the expectedType
func testSchema(res *CheckResult, attrName string, value interface{}, expectedMulti bool, expectedType string) {
	// Check if it's a multi value (array) and unwrap if necessary
	// TODO should we check all values in the array?
	isMulti := false
	switch value.(type) {
	case []interface{}:
		if len(value.([]interface{})) > 0 {
			value = value.([]interface{})[0]
		} else {
			value = nil
		}
		isMulti = true
	}

	if expectedMulti != isMulti {
		res.errf("expected multi=%t but multi=%t", expectedMulti, isMulti)
	}

	switch value.(type) {
	case string:
		if expectedType == "int" {
			_, err := strconv.Atoi(value.(string))
			if err != nil {
				res.errf("failed to convert int to string on field %s", attrName)
			}
		}
		if expectedType != "string" && expectedType != "int" {
			res.errf("%s expected type %q but was 'string'", attrName, expectedType)
		}
	case bool:
		if expectedType != "boolean" {
			res.errf("expected type %q but was 'boolean'", expectedType)
		}
	case float64:
		if expectedType != "int" {
			res.errf("expected type %q but was 'int'", expectedType)
		}
	case nil:
		// If a value is nil we can't validate the type.
	default:
		res.errf("unknown type %T for %q", value, attrName)
	}
}

// attrChange generates an attribute change event for the provided account and
// attribute.
func attrChange(acct *client.Account, attr *client.AccountSchemaAttribute) client.AttributeChange {
	var op string
	switch attr.Multi {
	case true:
		op = "Add"
	case false:
		op = "Set"
	}

	var newValue interface{}
	switch attr.Type {
	case "string":
		if attr.Name == "email" {
			newValue = fmt.Sprintf("test.%d@example.com", rand.Intn(65536))
		} else {
			newValue = fmt.Sprintf("string.%x", rand.Intn(16777216))
		}
	case "int":
		if current, found := acct.Attributes[attr.Name]; found {
			newValue = current.(int) + 1
		} else {
			newValue = 42
		}
	case "boolean":
		// flip
		if current, found := acct.Attributes[attr.Name]; found {
			newValue = current.(bool)
		} else {
			newValue = true
		}
	}

	return client.AttributeChange{
		Op:        op,
		Attribute: attr.Name,
		Value:     newValue,
	}
}

func isAvailableForUpdating(entitlements []string, entitlementID string) bool {
	entID := strings.Split(entitlementID, ":")[0]

	for _, ent := range entitlements {
		if entID == strings.Split(ent, ":")[0] {
			return false
		}

	}

	return true
}

func getIdentity(input map[string]interface{}) string {
	_, ok := input["email"]
	if ok {
		return input["email"].(string)
	}

	_, ok = input["username"]
	if ok {
		return input["username"].(string)
	}

	_, ok = input["name"]
	if ok {
		return input["name"].(string)
	}

	return fmt.Sprintf("test.%d@example.com", rand.Intn(65536))
}

func canonicalizeAttributes(attrs map[string]interface{}) {
	for key, val := range attrs {
		switch val.(type) {
		case []interface{}:
			var arrayOfStrings []string

			for _, elem := range val.([]interface{}) {
				arrayOfStrings = append(arrayOfStrings, fmt.Sprintf("%v", elem))
			}

			sort.Strings(arrayOfStrings)

			attrs[key] = arrayOfStrings
		case []float64:
			var arrayOfFloats []float64

			for _, elem := range val.([]float64) {
				arrayOfFloats = append(arrayOfFloats, elem)
			}

			sort.Float64s(arrayOfFloats)

			attrs[key] = arrayOfFloats
		case []int:
			var arrayOfInts []int

			for _, elem := range val.([]int) {
				arrayOfInts = append(arrayOfInts, elem)
			}

			sort.Ints(arrayOfInts)

			attrs[key] = arrayOfInts
		case []string:
			var arrayOfStrings []string

			for _, elem := range val.([]string) {
				arrayOfStrings = append(arrayOfStrings, elem)
			}

			sort.Strings(arrayOfStrings)

			attrs[key] = arrayOfStrings
		}
	}
}

func isCheckPossible(commands, checkCommands []string) (bool, []string) {
	var result []string

	commandsMap := make(map[string]bool)

	for _, c := range commands {
		commandsMap[c] = true
	}

	for _, cc := range checkCommands {
		_, ok := commandsMap[cc]
		if !ok {
			result = append(result, cc)
		}
	}

	if len(result) != 0 {
		return false, result
	}

	return true, nil
}
