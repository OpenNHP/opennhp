# 编译make教程

## 步骤一

基于 Ubuntu的 Linux 发行版本都可以使用 apt-get 命令来进行安装：

```pretty
sudo apt-get install golang
```

要查看当前系统安装的 Go 语言版本可以使用如下命令：

```pretty
go version
```

由于 Go 代码必需保存在 workspace(工作区)中，所以我们必需在 **Home** 目录(例如 ~/workspace)创建一个 **workspace** 目录并定义 **GOPATH** 环境变量指向该目录，这个目录将被 Go 工具用于保存和编辑二进制文件。

```pretty
mkdir ~/workspace
echo 'export GOPATH="$HOME/workspace"' >> ~/.bashrc
source ~/.bashrc
```

根据不同的需要，我们可以使用 apt-get 安装 Go tools：

```pretty
sudo apt-cache search golang
```

## 步骤二

go环境设置

```
go env -w GOPROXY="https://goproxy.cn,direct"
```

## 步骤三

**Linux与macOS**：运行代码根目录下脚本
`make`

## 编译出错（因版本问题）

在root用户下操作

```
sudo passwd
输入密码
su
```

需要先删除旧版本的go环境，忘记安装路径的可以通过环境变量查看

```
echo $GOROOT
```

通常是在/usr/local/go/目录下，删除

```
rm -rf $GOROOT
```


下载新版的go，这里的version是1.20

```
wget https://golang.google.cn/dl/go1.20.linux-amd64.tar.gz
```


或者到 Golang官网 手动下载压缩包

解压到usr/local下（官方推荐）

```
tar -C /usr/local/ -zxvf go1.20.linux-amd64.tar.gz
```


修改配置文件（系统配置为/etc/profile，用户配置为~/.profile），这里就修改系统配置
在文件最后加上两行（如果有旧版本的go配置就不用加，或者要修改路径）

```
export GOROOT=/usr/local/go
export PATH=$PATH:$GOROOT/bin  

```


或者

```
sed -i '$aexport GOROOT=/usr/local/go\nexport PATH=$PATH:$GOROOT/bin' /etc/profile
```


执行使配置文件生效

```
source /etc/profile
```


查看go版本

```
go version
```


这里有可能不是显示最新版本，很可能是因为旧版本的go可执行文件没有删除，只要用新安装的go可执行文件覆盖掉旧的好了

```
cp -f $GOROOT/bin/go* /usr/bin/
```

