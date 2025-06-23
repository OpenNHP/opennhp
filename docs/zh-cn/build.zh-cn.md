---
layout: page
title: 编译源代码
parent: 中文版
nav_order: 6
permalink: /zh-cn/build/
---

# 编译OpenNHP
{: .fs-9 }

[English](/build/){: .label .fs-4 }

---

## 1. WSL环境准备

**提示：** Windows 10/11下可以通过`WSL`子系统来运行Linux，详细请见WSL官方文档：<https://learn.microsoft.com/zh-cn/windows/wsl/install>

- **【开启WSL功能】** 在Win10上，需要首先开启WSL才能使用WSL安装Linux，设置界面请见下图。
   ![Win10上WSL设置](/images/win10wsl.png)
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

|        主机        |                    查看IP地址的命令                     |
| :----------------: | :-----------------------------------------------------: |
|   WSL中Linux主机   |            `hostname -I \| awk '{print $1}'`            |
| WSL宿主Windows主机 | `ip route show \| grep -i default \| awk '{ print $3}'` |

## 2. 系统需求

- 2.1 `Go语言`环境：**Go 1.23** 。安装包下载地址: <https://go.dev/dl/>
  - **Windows与macOS**环境下，通过下载的安装程序来安装Go。
  - **Linux**环境下可以直接通过管理工具安装： `sudo apt install golang `
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
- 2.2 `GCC`环境：
  - **Linux与macOS**：**GCC 8.0**或以上。
    - 查看GCC版本的命令：`gcc -v`
    - 安装GCC： `sudo apt install build-essential`
  - **Windows**:
    1. 第一步：**安装mingw64**。mingw64可以通过msys2的包管理工具进行下载。安装msys2系统要求、下载与安装教程见：<https://www.msys2.org/>。
    ![install_msys2](/images/install_msys2.png)

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

<small>*注：如果 2.1 和 2.2 已完成，直接在项目目录下执行编译命令 `.\build.bat` 时，通常会遇到 `系统找不到指定的路径`或 ` 'lib' 不是内部或外部命令，也不是可运行的程序或批处理文件。` 的错误。2.3 提供了解决该问题的方法，供参考使用。*</small>

- 2.3 `lib`环境：


  - 在编译运行的命令中使用了 lib 工具，这是用于生成 .lib 文件的工具，通常用于链接静态库或导出符号表（在 Windows 中生成 .lib 文件以便与 .dll 文件配合使用）。遇到的错误提示 lib 不是内部或外部命令，表示系统找不到 lib 工具。

  - **解决（'lib' 不是内部或外部命令，也不是可运行的程序或批处理文件）问题 ：** 安装 Visual Studio 和 Visual Studio tools。

    - lib 工具是微软的库管理工具，通常随 Visual Studio 的 Microsoft Build Tools 安装。确保你已安装 Visual Studio，并且选择了 C++ 生成工具（C++ Build Tools）组件，其中包括 lib.exe。

    - 如果还没有安装 Visual Studio，可以从 Visual Studio 官方网站下载安装：https://visualstudiomicrosoft.com/zh-hans/ 安装时，选择“桌面开发(C++)”工作负载，它包含 lib.exe 及其他必要的工具。

    - 安装 Visual Studio 后，确保使用 Visual Studio 开发者命令行（Developer Command Prompt） 来运行包含 lib 命令的 `build.bat `文件。这个命令行工具会自动加载构建工具的环境变量，如 lib.exe

   - **解决（系统找不到指定路径的错误）问题 ：** 更改`bulid.bat`文件中的路径

     - 打开 `build.bat` 文件，找到
     ```bat
     call "C:\Program Files (x86)\Microsoft Visual Studio\2019\Community\VC\Auxiliary\Build\vcvarsall.bat" x64
     ```

     - 修改为你自己的 visual studio目录下安装路径。比如：
     ```bat
     call "F:\develop\visualstu\VC\Auxiliary\Build\vcvarsall.bat" x64
     ```

- 2.4 `clang`编译环境(可选):

  - **提示：**
    - 关于clang编译工具，clang 只支持Linux，不支持windows，windows下无需安装clang。
    - 关于eBPF模块编译，eBPF不支持windows，eBPF只支持Linux及内核5.6版本以上。
  - 查看clang版本的命令：`clang --version`
  - **Linux Ubuntu**:
    - 安装clang llvm libbpf-dev：`sudo apt install clang llvm libbpf-dev`
  - **Linux Centos**:
    - 安装clang llvm libbpf-dev：`sudo yum install clang llvm libbpf-dev -y`


## 3. 编译

1. 拉取代码仓库

   ```bash
   git clone https://github.com/OpenNHP/opennhp.git
   ```

2. Go环境

   ```bash
   go env -w GOPROXY="https://goproxy.cn,direct"
   ```

3. 编译构建
   - **Linux与macOS**：运行代码根目录下脚本
   `make`
   - **Windows**：运行代码根目录下*BAT*文件
   `build.bat`<br>
   <small>*（注：如果在windows下编译过程中出现错误，请尝试此编译方法：在Visual Studio的developer command prompt for VS命令窗口中，切换到项目目录，执行`./build.bat`命令）*</small>
   - **Linux下编译eBPF**: 运行代码根目录下脚本
   `make ebpf`<br>
   <small>*（注：命令 `make ebpf`，会连带编译ebpf模块）*</small>

## 4. 结果

编译出来的二进制文件都在代码目录下的`release`子目录下。

- **NHP-Server**的可执行文件和配置文件： `release\nhp-server` 子目录下
- **NHP-AC**的可执行文件和配置文件： `release\nhp-ac` 子目录下
- **NHP-Agent**的可执行文件和配置文件： `release\nhp-agent` 子目录下
- **NHP-DB**的可执行文件和配置文件： `release\nhp-db` 子目录下
- **NHP-KGC**的可执行文件和配置文件： `release\nhp-kgc` 子目录下
- 所有二进制文件打包成一个`tar`文件:  `release\archive` 子目录下

---
