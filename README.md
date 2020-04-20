## 模块划分
* gtp模块为gtp-lib包
* pfcp模块为pfcp-lib包
* src模块包含服务目录
    * upf为upf服务

### 环境配置及依赖包下载

The following packages should be installed before starting.  

```shell-session
go get -u github.com/pkg/errors
go get -u github.com/google/go-cmp/cmp
go get -u github.com/pascaldekloe/goe/verify
go get -u github.com/vishvananda/netlink
go get -u github.com/vishvananda/netns
go get -u github.com/pascaldekloe/goe

```

