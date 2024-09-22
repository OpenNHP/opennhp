---
layout: page
title: How to Build
nav_order: 5
permalink: /build/
---

# Build OpenNHP Source Code
{: .fs-9 }
## 1. WSL Environment Setup

**Note：** You can run Linux through the WSL subsystem on Windows 10/11. For details, see the official WSL documentation: https://learn.microsoft.com/en-us/windows/wsl/install

- **【Enable the WSL function】** On Win10, you need to enable WSL first to use it for installing Linux. See the settings interface in the image below.
  
   ![Windows 10 on WSL Settings](/images/win10wsl.png)

- **【Install Linux on WSL】** It is recommended to install Ubuntu Linux on WSL by running the following command through PowerShell:

   ```bat
   wsl --update
   wsl --install -d Ubuntu
   ```

   If you encounter the following problems, refer to：<https://blog.csdn.net/weixin_44293949/article/details/121863559>

   ```text
   From 'https://raw.githubusercontent.com/microsoft/WSL/master/distributions/DistributionInfo.json' to extract the distribution list. The server name or address could not be resolved
   Error code: Wsl/WININET_E_NAME_NOT_RESOLVED
   ```

- **【IP address of the WSL environment】** In the Linux environment of WSL, run the following command to get the IP address:

|        Host machine        |             Command to view the IP address              |
| :------------------------: | :-----------------------------------------------------: |
|     Linux hosts in WSL     |            `hostname -I \| awk '{print $1}'`            |
| WSL hosts the Windows host | `ip route show \| grep -i default \| awk '{ print $3}'` |

## 2. System requirement

- 2.1 'Go Language' environment: **Go 1.18** or above. Installation package download: <https://go.dev/dl/>
  - **Windows and macOS**Environment, install Go through the downloaded installer.
  - **Linux** environment can be installed directly through the management tool: `sudo apt install golang`
  - After the installation is successful, run the command `go version`to see the Go version number.
  - **Windows and macOS**environment，Install Go through the downloaded installer.
  - **Linux**Environment can be installed directly through the management tool:`sudo apt install golang` Or install it manually with the following command:

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

  - After the installation is successful, run the command `go version` to see the Go version number.
- 2.2 `GCC`environment：
  - **Linux and macOS**：**GCC 8.0**or above。
    - To view the GCC version of the command:`gcc -v`
    - To install GCC: `sudo apt install build-essential`
  - **Windows**:
    1. Step 1: **Install mingw64**. mingw64 can be downloaded from msys2's package management tool. Installation requirements, downloads, and installation tutorials for msys2 are available at <https://www.msys2.org/>.
   
    ![install_msys2](/images/install_msys2.png)

    2. Step 2: **Install GCC**. Enter the command in msys2's console:

       ```bash
       pacman -S mingw-w64-ucrt-x86_64-gcc
       ```

    3. Step 3: **Configure GCC**. Add the GCC tool PATH to the Windows *%PATH%* environment variable. For example, if the installation path of mingw-w64-gcc is`C:\Program Files\MSYS2\ `, run the command

       ```bat
       setx PATH "%PATH%;C:\Program Files\MSYS2\ucrt64\bin
       ```
       After successful execution, open a new command line window and check the version number of *gcc*
       ```bat
       gcc --version
       ```

  - **Tip:** Under Windows can be ` WSL ` subsystem to run Linux, details please see WSL official document: < https://learn.microsoft.com/zh-cn/windows/wsl/install >
    - It is recommended to run the latest version of Ubuntu v22 on WSL and install it by running the following command from PowerShell on Windows:
      ```bat
      wsl --install --distribution Ubuntu-22.04
      ```

<small>*Note: If 2.1 and 2.2 are complete, when executing the compile command `.\build.bat `directly in the project directory, you will usually encounter` the system cannot find the specified path `or` 'lib' is not an internal or external command, nor is it a runnable program or batch file`The mistake. 2.3 Provides a solution to this problem for reference.*</small>

- 2.3 `lib`environment：


  - The lib utility is used in the compile run command, which is a tool for generating.lib files, usually for linking static libraries or exporting symbol tables (the.lib file is generated in Windows to work with the.dll file). The error message lib is not an internal or external command, indicating that the system cannot find the lib utility.
  
  - **To solve the problem ('lib' is not an internal or external command, nor is it a runnable program or batch file) :** Install Visual Studio and Visual Studio tools.

    - The lib tool is Microsoft's library management tool and is usually installed with Microsoft Build Tools for Visual Studio. Make sure you have Visual Studio installed and have selected the C++ Build Tools components, including lib.exe.

    - If you do not have Visual Studio installed, you can download and install it from the official Visual Studio website: https://visualstudiomicrosoft.com/zh-hans/ when installation, select the desktop development (c + +) "the workload, it contains the lib. Exe and other necessary tools.

    - After installing Visual Studio, make sure to use the Visual Studio Developer Command Prompt to run the `build.bat` file that contains the lib command. This command line tool automatically loads environment variables for the build tool, such as lib.exe
  
   - **To resolve the problem (the system cannot find the specified path) :** Change the path in the `build.bat` file

     - Open the `build.bat` file and find it
     ```bat
     call "C:\Program Files (x86)\Microsoft Visual Studio\2019\Community\VC\Auxiliary\Build\vcvarsall.bat" x64
     ```


     - Change the installation path to your own visual studio directory. For example:
     ```bat
     call "F:\develop\visualstu\VC\Auxiliary\Build\vcvarsall.bat" x64
     ```

## 3. compile

1. Pull the code repository

   ```bash
   git clone https://github.com/OpenNHP/opennhp.git
   ```

2. Go environment Settings

   ```bash
   go env -w GOPROXY="https://goproxy.cn,direct"
   ```

3. Compile and build
   - **Linux and macOS**：Run the script in the code root directory
   `make`
   - **Windows**：Run the *BAT* file in the code root directory
   `build.bat`<br>
   <small>*（Note: If an error occurs during the compilation process under windows, try this compilation method: In the Visual Studio developer command prompt for VS command window, switch to the project directory and execute the `./build.bat `command）*</small>

## 4. result

Compiled binaries are in the code directory under the `release` subdirectory.

- **NHP-Server** executable and configuration files: `release\nhp-server` subdirectory
- **NHP-AC** executable and configuration files: `release\nhp-ac` subdirectory
- **NHP-Agent** executable and configuration files: `release\nhp-agent` subdirectory
- All binaries are packaged into a `tar` file: `release\archive` subdirectory
  
[中文版](/zh-cn/build/){: .label .fs-4 }

---

