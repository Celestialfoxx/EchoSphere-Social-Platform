package backend

import (
    "context"
    "fmt"
    "io"

    "around/constants"

    "cloud.google.com/go/storage"
)

var (
	//创建一个GCSBackend作为全局变量
    GCSBackend *GoogleCloudStorageBackend
)

/* 
定义包装这个连接的包装类

benefits
1.抽象化和封装: 通过创建一个包装类，你可以将所有与GCS相关的操作封装在一个单独的类中。
这样，你的应用代码就不需要直接与storage.Client交互，而是通过这个包装类来进行。这增加了代码的可读性和维护性，因为所有与GCS交互的逻辑都在一个地方。

2.易于测试: 使用包装类可以更容易地对你的代码进行单元测试。你可以通过接口和依赖注入来模拟GoogleCloudStorageBackend类，
而不是尝试模拟底层的storage.Client。这样，你可以在不依赖于外部GCS服务的情况下测试你的应用逻辑。

3.灵活性和扩展性: 如果将来你需要切换到另一个存储解决方案（如Amazon S3）或者需要添加额外的功能（如缓存、监控或者额外的日志记录），
使用包装类会更加灵活。你只需要修改这个包装类的实现，而不是寻找和替换项目中所有直接使用storage.Client的地方。

4.隐藏实现细节: 包装类允许你隐藏与GCS交互的实现细节。例如，你可能需要处理认证、错误处理、重试逻辑或者性能优化等。
将这些逻辑放在包装类中，调用者就不需要关心这些底层细节。

5.统一配置和管理: 通过使用包装类，你可以在一个地方统一管理与存储相关的配置（如桶名称、权限、区域等）。
这比在每个需要访问存储的地方单独配置要简单和容易维护得多
*/
type GoogleCloudStorageBackend struct {
    client *storage.Client
    bucket string
}

func InitGCSBackend() {
	//创建一个连接，即client，用来连接GCS
    client, err := storage.NewClient(context.Background())
    if err != nil {
        panic(err)
    }

    GCSBackend = &GoogleCloudStorageBackend{
        client: client,
        bucket: constants.GCS_BUCKET,
    }
}

//读取文件(r), 文件名(objectName); 返回String(URL)和err
//io.Reader是一个接口， 是multipart的父类
func (backend *GoogleCloudStorageBackend) SaveToGCS(r io.Reader, objectName string) (string, error) {
    ctx := context.Background()
	//通过backend *GoogleCloudStorageBackend这个连接在GCS上创建一个储存的空间
    object := backend.client.Bucket(backend.bucket).Object(objectName)
    wc := object.NewWriter(ctx)
	//把文件r写到wc中，即存储到GCS中
    if _, err := io.Copy(wc, r); err != nil {
        return "", err
    }

	//关闭通道
    if err := wc.Close(); err != nil {
        return "", err
    }

	//ACL ：Access Control List, 即谁能访问这个文件。 此处设置为所有人均可读
	//因为客户端和后端在这个system中都会访问GCS，因此需要扩大权限
    if err := object.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
        return "", err
    }

	//拿到这个文件的attributes，其中包含url
    attrs, err := object.Attrs(ctx)
    if err != nil {
        return "", err
    }

    fmt.Printf("File is saved to GCS: %s\n", attrs.MediaLink)
    return attrs.MediaLink, nil
}