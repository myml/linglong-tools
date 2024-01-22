# ApiUploadTaskFileResp

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | Pointer to **int32** |  | [optional] 
**Data** | Pointer to [**ResponseUploadTaskResp**](ResponseUploadTaskResp.md) |  | [optional] 
**Msg** | Pointer to **string** |  | [optional] 
**TraceId** | Pointer to **string** |  | [optional] 

## Methods

### NewApiUploadTaskFileResp

`func NewApiUploadTaskFileResp() *ApiUploadTaskFileResp`

NewApiUploadTaskFileResp instantiates a new ApiUploadTaskFileResp object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewApiUploadTaskFileRespWithDefaults

`func NewApiUploadTaskFileRespWithDefaults() *ApiUploadTaskFileResp`

NewApiUploadTaskFileRespWithDefaults instantiates a new ApiUploadTaskFileResp object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *ApiUploadTaskFileResp) GetCode() int32`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *ApiUploadTaskFileResp) GetCodeOk() (*int32, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *ApiUploadTaskFileResp) SetCode(v int32)`

SetCode sets Code field to given value.

### HasCode

`func (o *ApiUploadTaskFileResp) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetData

`func (o *ApiUploadTaskFileResp) GetData() ResponseUploadTaskResp`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *ApiUploadTaskFileResp) GetDataOk() (*ResponseUploadTaskResp, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *ApiUploadTaskFileResp) SetData(v ResponseUploadTaskResp)`

SetData sets Data field to given value.

### HasData

`func (o *ApiUploadTaskFileResp) HasData() bool`

HasData returns a boolean if a field has been set.

### GetMsg

`func (o *ApiUploadTaskFileResp) GetMsg() string`

GetMsg returns the Msg field if non-nil, zero value otherwise.

### GetMsgOk

`func (o *ApiUploadTaskFileResp) GetMsgOk() (*string, bool)`

GetMsgOk returns a tuple with the Msg field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMsg

`func (o *ApiUploadTaskFileResp) SetMsg(v string)`

SetMsg sets Msg field to given value.

### HasMsg

`func (o *ApiUploadTaskFileResp) HasMsg() bool`

HasMsg returns a boolean if a field has been set.

### GetTraceId

`func (o *ApiUploadTaskFileResp) GetTraceId() string`

GetTraceId returns the TraceId field if non-nil, zero value otherwise.

### GetTraceIdOk

`func (o *ApiUploadTaskFileResp) GetTraceIdOk() (*string, bool)`

GetTraceIdOk returns a tuple with the TraceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraceId

`func (o *ApiUploadTaskFileResp) SetTraceId(v string)`

SetTraceId sets TraceId field to given value.

### HasTraceId

`func (o *ApiUploadTaskFileResp) HasTraceId() bool`

HasTraceId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


