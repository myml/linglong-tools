# Apiv2JSONError

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | Pointer to **int32** |  | [optional] 
**Fields** | Pointer to **map[string]string** |  | [optional] 
**Msg** | Pointer to **string** |  | [optional] 
**TraceId** | Pointer to **string** |  | [optional] 

## Methods

### NewApiv2JSONError

`func NewApiv2JSONError() *Apiv2JSONError`

NewApiv2JSONError instantiates a new Apiv2JSONError object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewApiv2JSONErrorWithDefaults

`func NewApiv2JSONErrorWithDefaults() *Apiv2JSONError`

NewApiv2JSONErrorWithDefaults instantiates a new Apiv2JSONError object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *Apiv2JSONError) GetCode() int32`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *Apiv2JSONError) GetCodeOk() (*int32, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *Apiv2JSONError) SetCode(v int32)`

SetCode sets Code field to given value.

### HasCode

`func (o *Apiv2JSONError) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetFields

`func (o *Apiv2JSONError) GetFields() map[string]string`

GetFields returns the Fields field if non-nil, zero value otherwise.

### GetFieldsOk

`func (o *Apiv2JSONError) GetFieldsOk() (*map[string]string, bool)`

GetFieldsOk returns a tuple with the Fields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFields

`func (o *Apiv2JSONError) SetFields(v map[string]string)`

SetFields sets Fields field to given value.

### HasFields

`func (o *Apiv2JSONError) HasFields() bool`

HasFields returns a boolean if a field has been set.

### GetMsg

`func (o *Apiv2JSONError) GetMsg() string`

GetMsg returns the Msg field if non-nil, zero value otherwise.

### GetMsgOk

`func (o *Apiv2JSONError) GetMsgOk() (*string, bool)`

GetMsgOk returns a tuple with the Msg field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMsg

`func (o *Apiv2JSONError) SetMsg(v string)`

SetMsg sets Msg field to given value.

### HasMsg

`func (o *Apiv2JSONError) HasMsg() bool`

HasMsg returns a boolean if a field has been set.

### GetTraceId

`func (o *Apiv2JSONError) GetTraceId() string`

GetTraceId returns the TraceId field if non-nil, zero value otherwise.

### GetTraceIdOk

`func (o *Apiv2JSONError) GetTraceIdOk() (*string, bool)`

GetTraceIdOk returns a tuple with the TraceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraceId

`func (o *Apiv2JSONError) SetTraceId(v string)`

SetTraceId sets TraceId field to given value.

### HasTraceId

`func (o *Apiv2JSONError) HasTraceId() bool`

HasTraceId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


