  <small>*注：该文件为OpenNHP项目在Windows环境下的编译说明 ，仅供参考。*</small>

  # Step 1. 安装go语言环境

  - Go语言环境：Go 1.18 或以上 。官网安装包下载地址: https://go.dev/dl/

  
  - 根据你的操作系统版本选择对应的安装包 ，Windows 系统通常选择 .msi 文件。
  
   ![安装页面](/images/install_go.png)

  - 下载完成后 ，双击下载的文件 ，会出现安装向导 ，接受协议 ，点击“Next”进行下一步。
  
  - 选择安装路径 ，然后点击“Next”。
  
  - 最后点击“Install”开始安装 ，安装完成后点击“Finish”。
  
  - 编辑环境变量：根据 ***自己的安装路径*** ，找到 `C:\Program Files\Go\bin` 目录 ，将该目录添加到环境变量 `PATH` 中。
  例如：

   ![编辑go环境变量](/images/Edit_Enviroment.png)

  - 完成后 ，在cmd命令行中输入命令 ，查看是否安装成功。<br>
  
  `go version`
  

  # Step 2. 安装windows下gcc编译器

<small>*注：MSYS2 是一个基于 Cygwin 的独立软件 ，它提供了一个包含大量开发工具的终端环境 ，常用于 Windows 系统上进行类 Unix 的开发和构建工作 。它特别适合需要跨平台编译和开发的程序员 ，可以方便地在 Windows 上使用 GCC 编译器和其他 Unix 工具*</small>

  - 安装mingw64。mingw64可以通过msys2的包管理工具进行下载。安装msys2系统要求、下载与安装教程见：https://www.msys2.org/
  
   ![安装mingw64](/images/install_msys2.png)

  - 安装后启动 MSYS2 的 MSYS2 Shell，执行以下命令更新软件包：<br>
  
  `pacman -Syu`

  - 之后，可以安装需要的工具，比如 GCC：<br>
  
  `pacman -S mingw-w64-x86_64-gcc`

  - 编辑环境变量：根据 ***自己的安装路径*** ，找到 `C:\Program Files\Go\bin` 目录 ，将该目录添加到环境变量 `PATH` 中。
  例如：

   ![编辑gcc环境变量](/images/Edit_env_msys2.png)

  - 完成后 ，在cmd命令行中输入命令 ，查看是否安装成功。<br>
  
  `gcc --version`

  # Step 3. 安装lib工具

  <small>*注：如果 Step 1 和 Step 2 已完成，直接在项目目录下执行编译命令 .\build.bat 时，通常会遇到 "系统找不到指定的路径。'lib' 不是内部或外部命令，也不是可运行的程序或批处理文件。" 的错误。Step 3 提供了解决该问题的方法，供参考使用。*</small>

  - 在编译运行的命令中使用了 lib 工具，这是用于生成 .lib 文件的工具，通常用于链接静态库或导出符号表（在 Windows 中生成 .lib 文件以便与 .dll 文件配合使用）。遇到的错误提示 lib 不是内部或外部命令，表示系统找不到 lib 工具。
  
  ## 解决（'lib' 不是内部或外部命令，也不是可运行的程序或批处理文件）问题 ：安装 Visual Studio 和 Visual Studio 工具。

  - lib 工具是微软的库管理工具，通常随 Visual Studio 的 Microsoft Build Tools 安装。确保你已安装 Visual Studio，并且选择了 C++ 生成工具（C++ Build Tools）组件，其中包括 lib.exe。

  - 如果还没有安装 Visual Studio，可以从 Visual Studio 官方网站下载安装：https://visualstudio.microsoft.com/zh-hans/ 
安装时，选择“桌面开发(C++)”工作负载，它包含 lib.exe 及其他必要的工具。

  - 安装 Visual Studio 后，确保使用 Visual Studio 开发者命令行（Developer Command Prompt） 来运行包含 lib 命令的 build.bat 文件。这个命令行工具会自动加载构建工具的环境变量，如 lib.exe
  
  ## 解决（系统找不到指定路径的错误）问题 ：更改bulid.bat文件中的路径

  - 打开 build.bat 文件，找到<br> 
  
  `call "C:\Program Files (x86)\Microsoft Visual Studio\2019\Community\VC\Auxiliary\Build\vcvarsall.bat" x64`

  - 修改为你自己的 visual studiom目录下安装路径。比如：<br>

  `call "F:\develop\visualstu\VC\Auxiliary\Build\vcvarsall.bat" x64`
  
  # Step 4. 编译

   - 在Visual Studio的developer command prompt for VS命令窗口中，切换到项目目录，执行编译命令<br>
  
  `.\build.bat`
