# ModelApp

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AppId** | Pointer to **string** | AppID             e.g. org.deepin.test | [optional] 
**Arch** | Pointer to **string** | CPU Architecture, e.g. x86_64 | [optional] 
**Channel** | Pointer to **string** | e.g. stable | [optional] 
**CheckSum** | Pointer to **string** | checksum, commit id | [optional] 
**CreatedAt** | Pointer to **string** |  | [optional] 
**DeletedAt** | Pointer to [**GormDeletedAt**](GormDeletedAt.md) |  | [optional] 
**Description** | Pointer to **string** | description | [optional] 
**Id** | Pointer to **int32** |  | [optional] 
**Kind** | Pointer to **string** | kind | [optional] 
**Module** | Pointer to **string** | e.g. binary | [optional] 
**Name** | Pointer to **string** | App Name          e.g. Deepin Music | [optional] 
**RepoName** | Pointer to **string** | repo name | [optional] 
**Runtime** | Pointer to **string** | runtime | [optional] 
**Size** | Pointer to **int32** | size | [optional] 
**UabUrl** | Pointer to **string** | deprecated | [optional] 
**UpdatedAt** | Pointer to **string** |  | [optional] 
**Uuid** | Pointer to **string** | 为了兼容旧接口，商店需要uuid做主键 | [optional] 
**Version** | Pointer to **string** | App Version       e.g. 0.0.1 | [optional] 

## Methods

### NewModelApp

`func NewModelApp() *ModelApp`

NewModelApp instantiates a new ModelApp object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewModelAppWithDefaults

`func NewModelAppWithDefaults() *ModelApp`

NewModelAppWithDefaults instantiates a new ModelApp object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAppId

`func (o *ModelApp) GetAppId() string`

GetAppId returns the AppId field if non-nil, zero value otherwise.

### GetAppIdOk

`func (o *ModelApp) GetAppIdOk() (*string, bool)`

GetAppIdOk returns a tuple with the AppId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAppId

`func (o *ModelApp) SetAppId(v string)`

SetAppId sets AppId field to given value.

### HasAppId

`func (o *ModelApp) HasAppId() bool`

HasAppId returns a boolean if a field has been set.

### GetArch

`func (o *ModelApp) GetArch() string`

GetArch returns the Arch field if non-nil, zero value otherwise.

### GetArchOk

`func (o *ModelApp) GetArchOk() (*string, bool)`

GetArchOk returns a tuple with the Arch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArch

`func (o *ModelApp) SetArch(v string)`

SetArch sets Arch field to given value.

### HasArch

`func (o *ModelApp) HasArch() bool`

HasArch returns a boolean if a field has been set.

### GetChannel

`func (o *ModelApp) GetChannel() string`

GetChannel returns the Channel field if non-nil, zero value otherwise.

### GetChannelOk

`func (o *ModelApp) GetChannelOk() (*string, bool)`

GetChannelOk returns a tuple with the Channel field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChannel

`func (o *ModelApp) SetChannel(v string)`

SetChannel sets Channel field to given value.

### HasChannel

`func (o *ModelApp) HasChannel() bool`

HasChannel returns a boolean if a field has been set.

### GetCheckSum

`func (o *ModelApp) GetCheckSum() string`

GetCheckSum returns the CheckSum field if non-nil, zero value otherwise.

### GetCheckSumOk

`func (o *ModelApp) GetCheckSumOk() (*string, bool)`

GetCheckSumOk returns a tuple with the CheckSum field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCheckSum

`func (o *ModelApp) SetCheckSum(v string)`

SetCheckSum sets CheckSum field to given value.

### HasCheckSum

`func (o *ModelApp) HasCheckSum() bool`

HasCheckSum returns a boolean if a field has been set.

### GetCreatedAt

`func (o *ModelApp) GetCreatedAt() string`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *ModelApp) GetCreatedAtOk() (*string, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *ModelApp) SetCreatedAt(v string)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *ModelApp) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetDeletedAt

`func (o *ModelApp) GetDeletedAt() GormDeletedAt`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *ModelApp) GetDeletedAtOk() (*GormDeletedAt, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *ModelApp) SetDeletedAt(v GormDeletedAt)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *ModelApp) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### GetDescription

`func (o *ModelApp) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *ModelApp) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *ModelApp) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *ModelApp) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetId

`func (o *ModelApp) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ModelApp) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ModelApp) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *ModelApp) HasId() bool`

HasId returns a boolean if a field has been set.

### GetKind

`func (o *ModelApp) GetKind() string`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *ModelApp) GetKindOk() (*string, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *ModelApp) SetKind(v string)`

SetKind sets Kind field to given value.

### HasKind

`func (o *ModelApp) HasKind() bool`

HasKind returns a boolean if a field has been set.

### GetModule

`func (o *ModelApp) GetModule() string`

GetModule returns the Module field if non-nil, zero value otherwise.

### GetModuleOk

`func (o *ModelApp) GetModuleOk() (*string, bool)`

GetModuleOk returns a tuple with the Module field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModule

`func (o *ModelApp) SetModule(v string)`

SetModule sets Module field to given value.

### HasModule

`func (o *ModelApp) HasModule() bool`

HasModule returns a boolean if a field has been set.

### GetName

`func (o *ModelApp) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ModelApp) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ModelApp) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ModelApp) HasName() bool`

HasName returns a boolean if a field has been set.

### GetRepoName

`func (o *ModelApp) GetRepoName() string`

GetRepoName returns the RepoName field if non-nil, zero value otherwise.

### GetRepoNameOk

`func (o *ModelApp) GetRepoNameOk() (*string, bool)`

GetRepoNameOk returns a tuple with the RepoName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepoName

`func (o *ModelApp) SetRepoName(v string)`

SetRepoName sets RepoName field to given value.

### HasRepoName

`func (o *ModelApp) HasRepoName() bool`

HasRepoName returns a boolean if a field has been set.

### GetRuntime

`func (o *ModelApp) GetRuntime() string`

GetRuntime returns the Runtime field if non-nil, zero value otherwise.

### GetRuntimeOk

`func (o *ModelApp) GetRuntimeOk() (*string, bool)`

GetRuntimeOk returns a tuple with the Runtime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRuntime

`func (o *ModelApp) SetRuntime(v string)`

SetRuntime sets Runtime field to given value.

### HasRuntime

`func (o *ModelApp) HasRuntime() bool`

HasRuntime returns a boolean if a field has been set.

### GetSize

`func (o *ModelApp) GetSize() int32`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *ModelApp) GetSizeOk() (*int32, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *ModelApp) SetSize(v int32)`

SetSize sets Size field to given value.

### HasSize

`func (o *ModelApp) HasSize() bool`

HasSize returns a boolean if a field has been set.

### GetUabUrl

`func (o *ModelApp) GetUabUrl() string`

GetUabUrl returns the UabUrl field if non-nil, zero value otherwise.

### GetUabUrlOk

`func (o *ModelApp) GetUabUrlOk() (*string, bool)`

GetUabUrlOk returns a tuple with the UabUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUabUrl

`func (o *ModelApp) SetUabUrl(v string)`

SetUabUrl sets UabUrl field to given value.

### HasUabUrl

`func (o *ModelApp) HasUabUrl() bool`

HasUabUrl returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *ModelApp) GetUpdatedAt() string`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *ModelApp) GetUpdatedAtOk() (*string, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *ModelApp) SetUpdatedAt(v string)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *ModelApp) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetUuid

`func (o *ModelApp) GetUuid() string`

GetUuid returns the Uuid field if non-nil, zero value otherwise.

### GetUuidOk

`func (o *ModelApp) GetUuidOk() (*string, bool)`

GetUuidOk returns a tuple with the Uuid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUuid

`func (o *ModelApp) SetUuid(v string)`

SetUuid sets Uuid field to given value.

### HasUuid

`func (o *ModelApp) HasUuid() bool`

HasUuid returns a boolean if a field has been set.

### GetVersion

`func (o *ModelApp) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *ModelApp) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *ModelApp) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *ModelApp) HasVersion() bool`

HasVersion returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


