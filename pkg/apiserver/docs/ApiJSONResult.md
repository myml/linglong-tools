# ApiJSONResult

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | Pointer to **int32** |  | [optional] 
**Data** | Pointer to **map[string]interface{}** |  | [optional] 
**Msg** | Pointer to **string** |  | [optional] 
**TraceId** | Pointer to **string** |  | [optional] 

## Methods

### NewApiJSONResult

`func NewApiJSONResult() *ApiJSONResult`

NewApiJSONResult instantiates a new ApiJSONResult object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewApiJSONResultWithDefaults

`func NewApiJSONResultWithDefaults() *ApiJSONResult`

NewApiJSONResultWithDefaults instantiates a new ApiJSONResult object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *ApiJSONResult) GetCode() int32`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *ApiJSONResult) GetCodeOk() (*int32, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *ApiJSONResult) SetCode(v int32)`

SetCode sets Code field to given value.

### HasCode

`func (o *ApiJSONResult) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetData

`func (o *ApiJSONResult) GetData() map[string]interface{}`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *ApiJSONResult) GetDataOk() (*map[string]interface{}, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *ApiJSONResult) SetData(v map[string]interface{})`

SetData sets Data field to given value.

### HasData

`func (o *ApiJSONResult) HasData() bool`

HasData returns a boolean if a field has been set.

### GetMsg

`func (o *ApiJSONResult) GetMsg() string`

GetMsg returns the Msg field if non-nil, zero value otherwise.

### GetMsgOk

`func (o *ApiJSONResult) GetMsgOk() (*string, bool)`

GetMsgOk returns a tuple with the Msg field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMsg

`func (o *ApiJSONResult) SetMsg(v string)`

SetMsg sets Msg field to given value.

### HasMsg

`func (o *ApiJSONResult) HasMsg() bool`

HasMsg returns a boolean if a field has been set.

### GetTraceId

`func (o *ApiJSONResult) GetTraceId() string`

GetTraceId returns the TraceId field if non-nil, zero value otherwise.

### GetTraceIdOk

`func (o *ApiJSONResult) GetTraceIdOk() (*string, bool)`

GetTraceIdOk returns a tuple with the TraceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraceId

`func (o *ApiJSONResult) SetTraceId(v string)`

SetTraceId sets TraceId field to given value.

### HasTraceId

`func (o *ApiJSONResult) HasTraceId() bool`

HasTraceId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


