# \ClientAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**FuzzySearchApp**](ClientAPI.md#FuzzySearchApp) | **Post** /api/v0/apps/fuzzysearchapp | 模糊查找App
[**GetRepo**](ClientAPI.md#GetRepo) | **Get** /api/v1/repos/{repo} | 查看仓库信息
[**NewUploadTaskID**](ClientAPI.md#NewUploadTaskID) | **Post** /api/v1/upload-tasks | generate a new upload task id
[**SignIn**](ClientAPI.md#SignIn) | **Post** /api/v1/sign-in | 登陆帐号
[**UploadTaskFile**](ClientAPI.md#UploadTaskFile) | **Put** /api/v1/upload-tasks/{task_id}/tar | upload tgz file to upload task
[**UploadTaskInfo**](ClientAPI.md#UploadTaskInfo) | **Get** /api/v1/upload-tasks/{task_id}/status | get upload task status
[**UploadTaskLayerFile**](ClientAPI.md#UploadTaskLayerFile) | **Put** /api/v1/upload-tasks/{task_id}/layer | upload layer file to upload task



## FuzzySearchApp

> FuzzySearchApp200Response FuzzySearchApp(ctx).Data(data).Execute()

模糊查找App

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
    data := *openapiclient.NewRequestFuzzySearchReq() // RequestFuzzySearchReq | app json数据

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClientAPI.FuzzySearchApp(context.Background()).Data(data).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClientAPI.FuzzySearchApp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `FuzzySearchApp`: FuzzySearchApp200Response
    fmt.Fprintf(os.Stdout, "Response from `ClientAPI.FuzzySearchApp`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiFuzzySearchAppRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **data** | [**RequestFuzzySearchReq**](RequestFuzzySearchReq.md) | app json数据 | 

### Return type

[**FuzzySearchApp200Response**](FuzzySearchApp200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRepo

> GetRepo200Response GetRepo(ctx, repo).Execute()

查看仓库信息



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
    repo := "repo_example" // string | 仓库名称

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClientAPI.GetRepo(context.Background(), repo).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClientAPI.GetRepo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetRepo`: GetRepo200Response
    fmt.Fprintf(os.Stdout, "Response from `ClientAPI.GetRepo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**repo** | **string** | 仓库名称 | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRepoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**GetRepo200Response**](GetRepo200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## NewUploadTaskID

> NewUploadTaskID200Response NewUploadTaskID(ctx).XToken(xToken).Req(req).Execute()

generate a new upload task id



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
    xToken := "xToken_example" // string | 31a165ba1be6dec616b1f8f3207b4273
    req := *openapiclient.NewSchemaNewUploadTaskReq() // SchemaNewUploadTaskReq | JSON数据

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClientAPI.NewUploadTaskID(context.Background()).XToken(xToken).Req(req).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClientAPI.NewUploadTaskID``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `NewUploadTaskID`: NewUploadTaskID200Response
    fmt.Fprintf(os.Stdout, "Response from `ClientAPI.NewUploadTaskID`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiNewUploadTaskIDRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xToken** | **string** | 31a165ba1be6dec616b1f8f3207b4273 | 
 **req** | [**SchemaNewUploadTaskReq**](SchemaNewUploadTaskReq.md) | JSON数据 | 

### Return type

[**NewUploadTaskID200Response**](NewUploadTaskID200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SignIn

> SignIn200Response SignIn(ctx).Data(data).Execute()

登陆帐号

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
    data := *openapiclient.NewRequestAuth() // RequestAuth | auth json数据

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClientAPI.SignIn(context.Background()).Data(data).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClientAPI.SignIn``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SignIn`: SignIn200Response
    fmt.Fprintf(os.Stdout, "Response from `ClientAPI.SignIn`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSignInRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **data** | [**RequestAuth**](RequestAuth.md) | auth json数据 | 

### Return type

[**SignIn200Response**](SignIn200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UploadTaskFile

> UploadTaskLayerFile200Response UploadTaskFile(ctx, taskId).XToken(xToken).File(file).Execute()

upload tgz file to upload task



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
    xToken := "xToken_example" // string | 31a165ba1be6dec616b1f8f3207b4273
    taskId := "taskId_example" // string | task id
    file := os.NewFile(1234, "some_file") // *os.File | 文件路径

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClientAPI.UploadTaskFile(context.Background(), taskId).XToken(xToken).File(file).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClientAPI.UploadTaskFile``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UploadTaskFile`: UploadTaskLayerFile200Response
    fmt.Fprintf(os.Stdout, "Response from `ClientAPI.UploadTaskFile`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**taskId** | **string** | task id | 

### Other Parameters

Other parameters are passed through a pointer to a apiUploadTaskFileRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xToken** | **string** | 31a165ba1be6dec616b1f8f3207b4273 | 

 **file** | ***os.File** | 文件路径 | 

### Return type

[**UploadTaskLayerFile200Response**](UploadTaskLayerFile200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: multipart/form-data
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UploadTaskInfo

> UploadTaskInfo200Response UploadTaskInfo(ctx, taskId).XToken(xToken).Execute()

get upload task status



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
    xToken := "xToken_example" // string | 31a165ba1be6dec616b1f8f3207b4273
    taskId := "taskId_example" // string | task id

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClientAPI.UploadTaskInfo(context.Background(), taskId).XToken(xToken).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClientAPI.UploadTaskInfo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UploadTaskInfo`: UploadTaskInfo200Response
    fmt.Fprintf(os.Stdout, "Response from `ClientAPI.UploadTaskInfo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**taskId** | **string** | task id | 

### Other Parameters

Other parameters are passed through a pointer to a apiUploadTaskInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xToken** | **string** | 31a165ba1be6dec616b1f8f3207b4273 | 


### Return type

[**UploadTaskInfo200Response**](UploadTaskInfo200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UploadTaskLayerFile

> UploadTaskLayerFile200Response UploadTaskLayerFile(ctx, taskId).XToken(xToken).File(file).Execute()

upload layer file to upload task

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
    xToken := "xToken_example" // string | 31a165ba1be6dec616b1f8f3207b4273
    taskId := "taskId_example" // string | task id
    file := os.NewFile(1234, "some_file") // *os.File | 文件路径

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ClientAPI.UploadTaskLayerFile(context.Background(), taskId).XToken(xToken).File(file).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ClientAPI.UploadTaskLayerFile``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UploadTaskLayerFile`: UploadTaskLayerFile200Response
    fmt.Fprintf(os.Stdout, "Response from `ClientAPI.UploadTaskLayerFile`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**taskId** | **string** | task id | 

### Other Parameters

Other parameters are passed through a pointer to a apiUploadTaskLayerFileRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xToken** | **string** | 31a165ba1be6dec616b1f8f3207b4273 | 

 **file** | ***os.File** | 文件路径 | 

### Return type

[**UploadTaskLayerFile200Response**](UploadTaskLayerFile200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: multipart/form-data
- **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

