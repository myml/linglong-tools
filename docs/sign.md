# 签名技术细节

## layer 签名

layer 文件的原始结构是

```
<head(chars 40 byte)>
<json metainfo size(uint32 4 byte)>
<json metainfo(varchar)>
<erofs image>
```

为避免对 erofs 镜像做修改，layer 的签名数据放置到尾部

签名后的数据结构是

```
<head(chars 40 byte)>
<json metainfo size(uint32 4 byte)>
<json metainfo(json)>
<erofs data(image)>
<sign data(tar)>
```

签名时会在 metainfo 添加 erofs_size 和 sign_size 字段，便于后续读取和辨认。

## uab 签名

uab 是 elf 文件，签名数据会插入到 elf 的 section 中，section 名字是 'linglong.bundle.sign'

## 签名推仓

未签名的 layer 文件会直接推送，服务器支持 layer 文件直接推送。

已签名的 layer 文件和 uab 文件会先进行解压，然后打包成 tar.gz 再推送。解压会同时将签名数据也解压出来，所以签名数据也会被推送到服务器。
