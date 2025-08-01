# ApiV1ReposGet200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | Pointer to **int32** |  | [optional] 
**Data** | Pointer to [**[]SchemaRepoInfo**](SchemaRepoInfo.md) |  | [optional] 
**Msg** | Pointer to **string** |  | [optional] 
**TraceId** | Pointer to **string** |  | [optional] 

## Methods

### NewApiV1ReposGet200Response

`func NewApiV1ReposGet200Response() *ApiV1ReposGet200Response`

NewApiV1ReposGet200Response instantiates a new ApiV1ReposGet200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewApiV1ReposGet200ResponseWithDefaults

`func NewApiV1ReposGet200ResponseWithDefaults() *ApiV1ReposGet200Response`

NewApiV1ReposGet200ResponseWithDefaults instantiates a new ApiV1ReposGet200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *ApiV1ReposGet200Response) GetCode() int32`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *ApiV1ReposGet200Response) GetCodeOk() (*int32, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *ApiV1ReposGet200Response) SetCode(v int32)`

SetCode sets Code field to given value.

### HasCode

`func (o *ApiV1ReposGet200Response) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetData

`func (o *ApiV1ReposGet200Response) GetData() []SchemaRepoInfo`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *ApiV1ReposGet200Response) GetDataOk() (*[]SchemaRepoInfo, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *ApiV1ReposGet200Response) SetData(v []SchemaRepoInfo)`

SetData sets Data field to given value.

### HasData

`func (o *ApiV1ReposGet200Response) HasData() bool`

HasData returns a boolean if a field has been set.

### GetMsg

`func (o *ApiV1ReposGet200Response) GetMsg() string`

GetMsg returns the Msg field if non-nil, zero value otherwise.

### GetMsgOk

`func (o *ApiV1ReposGet200Response) GetMsgOk() (*string, bool)`

GetMsgOk returns a tuple with the Msg field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMsg

`func (o *ApiV1ReposGet200Response) SetMsg(v string)`

SetMsg sets Msg field to given value.

### HasMsg

`func (o *ApiV1ReposGet200Response) HasMsg() bool`

HasMsg returns a boolean if a field has been set.

### GetTraceId

`func (o *ApiV1ReposGet200Response) GetTraceId() string`

GetTraceId returns the TraceId field if non-nil, zero value otherwise.

### GetTraceIdOk

`func (o *ApiV1ReposGet200Response) GetTraceIdOk() (*string, bool)`

GetTraceIdOk returns a tuple with the TraceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraceId

`func (o *ApiV1ReposGet200Response) SetTraceId(v string)`

SetTraceId sets TraceId field to given value.

### HasTraceId

`func (o *ApiV1ReposGet200Response) HasTraceId() bool`

HasTraceId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


