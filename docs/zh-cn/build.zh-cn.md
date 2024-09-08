---
layout: page
title: 编译源代码
parent: 中文版
nav_order: 4
permalink: /zh-cn/build/
---

# 编译OpenNHP

**提示：** Windows 10/11下可以通过`WSL`子系统来运行Linux，详细请见WSL官方文档：<https://learn.microsoft.com/zh-cn/windows/wsl/install>

- **【开启WSL功能】** 在Win10上，需要首先开启WSL才能使用WSL安装Linux，设置界面请见下图。
   ![Win10上WSL设置](./docs/images/win10wsl.png)
- **【WSL上安装Linux】** 推荐在WSL上安装Ubuntu Linux，通过PowerShell运行以下命令安装：

   ```bat
   wsl --update
   wsl --install -d Ubuntu
   ```

   如果遇到以下问题，参考：<https://blog.csdn.net/weixin_44293949/article/details/121863559>

   ```text
   无法从 'https://raw.githubusercontent.com/microsoft/WSL/master/distributions/DistributionInfo.json’提取列表分发。无法解析服务器的名称或地址
   Error code: Wsl/WININET_E_NAME_NOT_RESOLVED
   ```

- **【WSL环境的IP地址】** 在WSL的Linux环境中，运行以下命令获取IP地址：
   | 主机 | 查看IP地址的命令  |
   |:--:|:--:|
   | WSL中Linux主机 | `hostname -I \| awk '{print $1}'` |  
   | WSL宿主Windows主机 | `ip route show \| grep -i default \| awk '{ print $3}'` |  

## 系统需求

- 3.1.1 `Go语言`环境：**Go 1.18** 或以上。安装包下载地址: <https://go.dev/dl/>
  - **Windows与macOS**环境下，通过下载的安装程序来安装Go。
  - **Linux**环境下可以直接通过管理工具安装： ：`sudo apt install golang ` 
  - 安装成功后，运行命令`go version` 来查看Go版本号。
  - **Windows与macOS**环境下，通过下载的安装程序来安装Go。
  - **Linux**环境下可以直接通过管理工具安装：`sudo apt install golang` 或者通过以下命令手动安装：

   ```bash
      1. sudo apt-get update
      2. wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
      3. sudo tar -xvf go1.21.0.linux-amd64.tar.gz
      4. sudo mv go /usr/local
      5. export GOROOT=/usr/local/go
      6. export GOPATH=$HOME/go
      7. export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
      8. source ~/.profile
   ```

  - 安装成功后，运行命令`go version` 来查看Go版本号。
- `GCC`环境：
  - **Linux与macOS**：**GCC 8.0**或以上。
    - 查看GCC版本的命令：`gcc -v`
    - 安装GCC： `sudo apt install build-essential`
  - **Windows**:
    1. 第一步：**安装mingw64**。mingw64可以通过msys2的包管理工具进行下载。安装msys2系统要求、下载与安装教程见：<https://www.msys2.org/>。
    ![install_msys2](./docs/images/install_msys2.png)

    2. 第二步：**安装GCC**。在msys2的控制台输入命令：

       ```bash
       pacman -S mingw-w64-ucrt-x86_64-gcc
       ```

    3. 第三步：**配置GCC**。将GCC工具路径加入Windows的 *%PATH%* 环境变量。例如：mingw-w64-gcc的安装路径为`C:\Program Files\MSYS2\`， 则需要运行命令

       ```bat
       setx PATH "%PATH%;C:\Program Files\MSYS2\ucrt64\bin
       ```
       执行成功之后，打开新的命令行窗口，检查*gcc*的版本号
       ```bat
       gcc --version
       ```

  - **提示：** Windows下可以通过`WSL`子系统来运行Linux，详细请见WSL官方文档：<https://learn.microsoft.com/zh-cn/windows/wsl/install> 
    - 推荐在WSL上运行Ubuntu最新版v22，在Windows上的PowerShell运行以下命令安装：
      ```bat
      wsl --install --distribution Ubuntu-22.04
      ```
## 构建

1. 拉取代码仓库

   ```bash
   git clone https://github.com/OpenNHP/opennhp.git
   ```

2. Go环境设置

   ```bash
   go env -w GOPROXY="https://goproxy.cn,direct"
   ```

3. 编译构建
   - **Linux与macOS**：运行代码根目录下脚本
   `make`
   - **Windows**：运行代码根目录下*BAT*文件
   `build.bat`

