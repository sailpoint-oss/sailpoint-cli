/*
IdentityNow cc (private) APIs

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: 1.0.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package sailpointccsdk

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)


// SourcesAccountSchemaApiService SourcesAccountSchemaApi service
type SourcesAccountSchemaApiService service

type ApiCreateAccountSchemaAttributeRequest struct {
	ctx context.Context
	ApiService *SourcesAccountSchemaApiService
	id string
	objectType *string
	entitlement *bool
	multi *bool
	names *string
	type_ *string
	description *string
}

func (r ApiCreateAccountSchemaAttributeRequest) ObjectType(objectType string) ApiCreateAccountSchemaAttributeRequest {
	r.objectType = &objectType
	return r
}

func (r ApiCreateAccountSchemaAttributeRequest) Entitlement(entitlement bool) ApiCreateAccountSchemaAttributeRequest {
	r.entitlement = &entitlement
	return r
}

func (r ApiCreateAccountSchemaAttributeRequest) Multi(multi bool) ApiCreateAccountSchemaAttributeRequest {
	r.multi = &multi
	return r
}

func (r ApiCreateAccountSchemaAttributeRequest) Names(names string) ApiCreateAccountSchemaAttributeRequest {
	r.names = &names
	return r
}

func (r ApiCreateAccountSchemaAttributeRequest) Type_(type_ string) ApiCreateAccountSchemaAttributeRequest {
	r.type_ = &type_
	return r
}

func (r ApiCreateAccountSchemaAttributeRequest) Description(description string) ApiCreateAccountSchemaAttributeRequest {
	r.description = &description
	return r
}

func (r ApiCreateAccountSchemaAttributeRequest) Execute() (*http.Response, error) {
	return r.ApiService.CreateAccountSchemaAttributeExecute(r)
}

/*
CreateAccountSchemaAttribute Create Account Schema Attribute

Add an attribute to a source schema.
@param id of the source.
@return JSON string of the created attribute.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id
 @return ApiCreateAccountSchemaAttributeRequest
*/
func (a *SourcesAccountSchemaApiService) CreateAccountSchemaAttribute(ctx context.Context, id string) ApiCreateAccountSchemaAttributeRequest {
	return ApiCreateAccountSchemaAttributeRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
func (a *SourcesAccountSchemaApiService) CreateAccountSchemaAttributeExecute(r ApiCreateAccountSchemaAttributeRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "SourcesAccountSchemaApiService.CreateAccountSchemaAttribute")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/cc/api/source/createSchemaAttribute/{id}"
	localVarPath = strings.Replace(localVarPath, "{"+"id"+"}", url.PathEscape(parameterToString(r.id, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"multipart/form-data"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.objectType != nil {
		localVarFormParams.Add("objectType", parameterToString(*r.objectType, ""))
	}
	if r.entitlement != nil {
		localVarFormParams.Add("entitlement", parameterToString(*r.entitlement, ""))
	}
	if r.multi != nil {
		localVarFormParams.Add("multi", parameterToString(*r.multi, ""))
	}
	if r.names != nil {
		localVarFormParams.Add("names", parameterToString(*r.names, ""))
	}
	if r.type_ != nil {
		localVarFormParams.Add("type", parameterToString(*r.type_, ""))
	}
	if r.description != nil {
		localVarFormParams.Add("description", parameterToString(*r.description, ""))
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiDeleteAccountSchemaAttributeRequest struct {
	ctx context.Context
	ApiService *SourcesAccountSchemaApiService
	id string
	objectType *string
	names *string
}

func (r ApiDeleteAccountSchemaAttributeRequest) ObjectType(objectType string) ApiDeleteAccountSchemaAttributeRequest {
	r.objectType = &objectType
	return r
}

func (r ApiDeleteAccountSchemaAttributeRequest) Names(names string) ApiDeleteAccountSchemaAttributeRequest {
	r.names = &names
	return r
}

func (r ApiDeleteAccountSchemaAttributeRequest) Execute() (*http.Response, error) {
	return r.ApiService.DeleteAccountSchemaAttributeExecute(r)
}

/*
DeleteAccountSchemaAttribute Delete Account Schema Attribute

Delete an attribute from a source schema.
@param ID of the source.
@return JSON status of OK.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id
 @return ApiDeleteAccountSchemaAttributeRequest
*/
func (a *SourcesAccountSchemaApiService) DeleteAccountSchemaAttribute(ctx context.Context, id string) ApiDeleteAccountSchemaAttributeRequest {
	return ApiDeleteAccountSchemaAttributeRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
func (a *SourcesAccountSchemaApiService) DeleteAccountSchemaAttributeExecute(r ApiDeleteAccountSchemaAttributeRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "SourcesAccountSchemaApiService.DeleteAccountSchemaAttribute")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/cc/api/source/deleteSchemaAttribute/{id}"
	localVarPath = strings.Replace(localVarPath, "{"+"id"+"}", url.PathEscape(parameterToString(r.id, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"multipart/form-data"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.objectType != nil {
		localVarFormParams.Add("objectType", parameterToString(*r.objectType, ""))
	}
	if r.names != nil {
		localVarFormParams.Add("names", parameterToString(*r.names, ""))
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiGetSourceAccountSchemaRequest struct {
	ctx context.Context
	ApiService *SourcesAccountSchemaApiService
	id string
	page *int32
	start *int32
	limit *int32
}

func (r ApiGetSourceAccountSchemaRequest) Page(page int32) ApiGetSourceAccountSchemaRequest {
	r.page = &page
	return r
}

func (r ApiGetSourceAccountSchemaRequest) Start(start int32) ApiGetSourceAccountSchemaRequest {
	r.start = &start
	return r
}

func (r ApiGetSourceAccountSchemaRequest) Limit(limit int32) ApiGetSourceAccountSchemaRequest {
	r.limit = &limit
	return r
}

func (r ApiGetSourceAccountSchemaRequest) Execute() (*http.Response, error) {
	return r.ApiService.GetSourceAccountSchemaExecute(r)
}

/*
GetSourceAccountSchema Get Account Schema

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id
 @return ApiGetSourceAccountSchemaRequest
*/
func (a *SourcesAccountSchemaApiService) GetSourceAccountSchema(ctx context.Context, id string) ApiGetSourceAccountSchemaRequest {
	return ApiGetSourceAccountSchemaRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
func (a *SourcesAccountSchemaApiService) GetSourceAccountSchemaExecute(r ApiGetSourceAccountSchemaRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "SourcesAccountSchemaApiService.GetSourceAccountSchema")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/cc/api/source/getAccountSchema/{id}"
	localVarPath = strings.Replace(localVarPath, "{"+"id"+"}", url.PathEscape(parameterToString(r.id, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.page != nil {
		localVarQueryParams.Add("page", parameterToString(*r.page, ""))
	}
	if r.start != nil {
		localVarQueryParams.Add("start", parameterToString(*r.start, ""))
	}
	if r.limit != nil {
		localVarQueryParams.Add("limit", parameterToString(*r.limit, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiUpdateAccountSchemaAttributeRequest struct {
	ctx context.Context
	ApiService *SourcesAccountSchemaApiService
	id string
	objectType *string
	fieldName *string
	fieldValue *bool
	names *string
}

func (r ApiUpdateAccountSchemaAttributeRequest) ObjectType(objectType string) ApiUpdateAccountSchemaAttributeRequest {
	r.objectType = &objectType
	return r
}

func (r ApiUpdateAccountSchemaAttributeRequest) FieldName(fieldName string) ApiUpdateAccountSchemaAttributeRequest {
	r.fieldName = &fieldName
	return r
}

func (r ApiUpdateAccountSchemaAttributeRequest) FieldValue(fieldValue bool) ApiUpdateAccountSchemaAttributeRequest {
	r.fieldValue = &fieldValue
	return r
}

func (r ApiUpdateAccountSchemaAttributeRequest) Names(names string) ApiUpdateAccountSchemaAttributeRequest {
	r.names = &names
	return r
}

func (r ApiUpdateAccountSchemaAttributeRequest) Execute() (string, *http.Response, error) {
	return r.ApiService.UpdateAccountSchemaAttributeExecute(r)
}

/*
UpdateAccountSchemaAttribute Update Account Schema Attribute

Update an attribute in the source's schema.
@param ID of the source.
@return JSON string of the created attribute.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id
 @return ApiUpdateAccountSchemaAttributeRequest
*/
func (a *SourcesAccountSchemaApiService) UpdateAccountSchemaAttribute(ctx context.Context, id string) ApiUpdateAccountSchemaAttributeRequest {
	return ApiUpdateAccountSchemaAttributeRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
//  @return string
func (a *SourcesAccountSchemaApiService) UpdateAccountSchemaAttributeExecute(r ApiUpdateAccountSchemaAttributeRequest) (string, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPost
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  string
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "SourcesAccountSchemaApiService.UpdateAccountSchemaAttribute")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/cc/api/source/updateSchemaAttributes/{id}"
	localVarPath = strings.Replace(localVarPath, "{"+"id"+"}", url.PathEscape(parameterToString(r.id, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"multipart/form-data"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"text/plain"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.objectType != nil {
		localVarFormParams.Add("objectType", parameterToString(*r.objectType, ""))
	}
	if r.fieldName != nil {
		localVarFormParams.Add("fieldName", parameterToString(*r.fieldName, ""))
	}
	if r.fieldValue != nil {
		localVarFormParams.Add("fieldValue", parameterToString(*r.fieldValue, ""))
	}
	if r.names != nil {
		localVarFormParams.Add("names", parameterToString(*r.names, ""))
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}