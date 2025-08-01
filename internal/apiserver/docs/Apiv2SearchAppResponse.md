# Apiv2SearchAppResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | Pointer to **int32** |  | [optional] 
**Data** | Pointer to [**[]RequestRegisterStruct**](RequestRegisterStruct.md) |  | [optional] 
**TraceId** | Pointer to **string** |  | [optional] 

## Methods

### NewApiv2SearchAppResponse

`func NewApiv2SearchAppResponse() *Apiv2SearchAppResponse`

NewApiv2SearchAppResponse instantiates a new Apiv2SearchAppResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewApiv2SearchAppResponseWithDefaults

`func NewApiv2SearchAppResponseWithDefaults() *Apiv2SearchAppResponse`

NewApiv2SearchAppResponseWithDefaults instantiates a new Apiv2SearchAppResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *Apiv2SearchAppResponse) GetCode() int32`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *Apiv2SearchAppResponse) GetCodeOk() (*int32, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *Apiv2SearchAppResponse) SetCode(v int32)`

SetCode sets Code field to given value.

### HasCode

`func (o *Apiv2SearchAppResponse) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetData

`func (o *Apiv2SearchAppResponse) GetData() []RequestRegisterStruct`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *Apiv2SearchAppResponse) GetDataOk() (*[]RequestRegisterStruct, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *Apiv2SearchAppResponse) SetData(v []RequestRegisterStruct)`

SetData sets Data field to given value.

### HasData

`func (o *Apiv2SearchAppResponse) HasData() bool`

HasData returns a boolean if a field has been set.

### GetTraceId

`func (o *Apiv2SearchAppResponse) GetTraceId() string`

GetTraceId returns the TraceId field if non-nil, zero value otherwise.

### GetTraceIdOk

`func (o *Apiv2SearchAppResponse) GetTraceIdOk() (*string, bool)`

GetTraceIdOk returns a tuple with the TraceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraceId

`func (o *Apiv2SearchAppResponse) SetTraceId(v string)`

SetTraceId sets TraceId field to given value.

### HasTraceId

`func (o *Apiv2SearchAppResponse) HasTraceId() bool`

HasTraceId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


