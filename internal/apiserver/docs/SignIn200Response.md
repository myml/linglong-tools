# SignIn200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | Pointer to **int32** |  | [optional] 
**Data** | Pointer to [**ResponseSignIn**](ResponseSignIn.md) |  | [optional] 
**Msg** | Pointer to **string** |  | [optional] 
**TraceId** | Pointer to **string** |  | [optional] 

## Methods

### NewSignIn200Response

`func NewSignIn200Response() *SignIn200Response`

NewSignIn200Response instantiates a new SignIn200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSignIn200ResponseWithDefaults

`func NewSignIn200ResponseWithDefaults() *SignIn200Response`

NewSignIn200ResponseWithDefaults instantiates a new SignIn200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *SignIn200Response) GetCode() int32`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *SignIn200Response) GetCodeOk() (*int32, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *SignIn200Response) SetCode(v int32)`

SetCode sets Code field to given value.

### HasCode

`func (o *SignIn200Response) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetData

`func (o *SignIn200Response) GetData() ResponseSignIn`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *SignIn200Response) GetDataOk() (*ResponseSignIn, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *SignIn200Response) SetData(v ResponseSignIn)`

SetData sets Data field to given value.

### HasData

`func (o *SignIn200Response) HasData() bool`

HasData returns a boolean if a field has been set.

### GetMsg

`func (o *SignIn200Response) GetMsg() string`

GetMsg returns the Msg field if non-nil, zero value otherwise.

### GetMsgOk

`func (o *SignIn200Response) GetMsgOk() (*string, bool)`

GetMsgOk returns a tuple with the Msg field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMsg

`func (o *SignIn200Response) SetMsg(v string)`

SetMsg sets Msg field to given value.

### HasMsg

`func (o *SignIn200Response) HasMsg() bool`

HasMsg returns a boolean if a field has been set.

### GetTraceId

`func (o *SignIn200Response) GetTraceId() string`

GetTraceId returns the TraceId field if non-nil, zero value otherwise.

### GetTraceIdOk

`func (o *SignIn200Response) GetTraceIdOk() (*string, bool)`

GetTraceIdOk returns a tuple with the TraceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraceId

`func (o *SignIn200Response) SetTraceId(v string)`

SetTraceId sets TraceId field to given value.

### HasTraceId

`func (o *SignIn200Response) HasTraceId() bool`

HasTraceId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


