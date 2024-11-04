---
layout: page
title: Overleaf 教程
parent: 中文版
nav_order: 11
permalink: /zh-cn/Overleaf/
---

# Overleaf 简明教程
{: .fs-9 }

[English](/Overleaf/){: .label .fs-4 }

---

## 1. LaTex简介

​	LaTeX（读作/ˈlɑːtɛx/或/ˈleɪtɛx/）是一种排版系统，专为使文档排版更专业而设计，区别于传统的文字处理软件。它非常适合用于篇幅较长、结构复杂的文档，尤其擅长处理数学公式。作为一款免费的软件，LaTeX 可以在大多数操作系统上使用。LaTeX 建立在 1978 年由 Donald Knuth 设计的 TeX 排版系统之上。TeX 是一种低级排版语言，尽管功能强大，但许多人认为它难以掌握。LaTeX 的设计初衷是为了简化 TeX 的使用。目前的 LaTeX 版本为 LaTeX 2e。

​	与 Word 相比，LaTeX 更适合于处理学术类、高格式要求的文档，特别是理工科领域。以下是LaTeX的优势所在。

- **精确的排版控制**：LaTeX 能够精确控制文档的版面布局，特别适合需要精密排版的学术论文、技术报告和书籍等长篇文档，尤其在公式、引用、图表等方面表现突出。
- **专业的数学公式处理**：LaTeX 在处理复杂的数学公式时具有极大的优势，提供了强大的公式编辑功能，能确保公式的格式和排版美观、专业。
- **自动化处理**：LaTeX 可以自动生成目录、引用、交叉引用、索引等，避免了手动更新的繁琐，特别是在处理多章节和大量引用时显著提高效率。
- **一致的格式规范**：通过模板和预定义的样式，LaTeX 保证了整个文档的格式统一。对于需要遵循严格排版要求的论文或出版物，这点尤为重要。
- **可移植性和免费开源**：LaTeX 是免费开源的，且兼容多种操作系统，可以自由使用和修改，同时文档文件（.tex 文件）轻量、跨平台易于编辑和共享。
- **文本文件格式**：LaTeX 的文档是纯文本文件（.tex），这使得其更易于版本控制，便于在协作中使用版本控制系统如 Git，同时也避免了二进制文件的兼容性问题。



## 2. LaTex编辑工具

​LaTeX 的源代码文件格式为 .tex，并以纯文本形式存储，因此可以用任意纯文本编辑器进行编辑。你可以使用系统自带的文本编辑器（如 Windows 的记事本），也可以选择专为 TeX 设计的编辑器（如 TeXworks、TeXmaker、TeXstudio、WinEdt 等），甚至是通用的文本编辑器（如 Sublime Text、Atom、Visual Studio Code 等）。无论选择哪种编辑器，都能编辑 .tex 文件。此外，TeX 的发行版通常会自带编辑器，例如 TeX Live 和 MiKTeX 都默认集成了 TeXworks 编辑器。

所谓 TeX 发行版，指的是一个包含 TeX 系统的完整软件包，内含各种可执行程序、辅助工具和宏包文档。这些发行版（如 TeX Live 和 MiKTeX）提供了 TeX 系统运行所需的所有资源。

本文只介绍 **Overleaf** 的使用，原因有以下这些：

- **在线协作功能**：Overleaf 是基于云的在线 LaTeX 编辑器，允许多个用户同时编辑同一份文档。实时同步和版本控制使得合作编写学术论文、项目报告等更加高效，适合团队协作和跨地区工作。

- **免安装环境**：用户无需安装任何本地软件或配置 TeX 发行版。只需登录 Overleaf，即可在浏览器中编写和编译 LaTeX 文档，特别方便那些不想花时间设置本地 LaTeX 环境的用户。

- **即时预览**：Overleaf 提供实时编译和预览功能，用户在编辑时可以即时看到排版效果，不必手动执行编译命令，大大提高了 LaTeX 初学者和有时限项目的工作效率。

- **模板库丰富**：Overleaf 内置了大量的模板库，包括论文、简历、书籍、演示文稿等多种类型的模板，尤其是许多**期刊和会议提供官方模板**，对方便用户直接使用标准格式。

- **版本控制和历史记录**：Overleaf 自动保存每次修改，并且允许用户回溯文档的历史版本，这对多人合作或大规模文档编辑时尤为有用，用户可以方便地恢复到之前的任何版本。

- **跨平台和无缝同步**：由于是基于云的平台，用户可以随时随地通过任何设备访问和编辑文档，保证了工作的连贯性和便捷性。并且，Overleaf 还支持与 GitHub 的集成，方便版本管理和代码协作。

- **共享和发布**：文档可以轻松生成共享链接，直接分享给合作作者或审稿人，还能一键发布到各类学术期刊或预印本平台（如 arXiv），加速论文发表过程。

  **Overleaf: https://cn.overleaf.com/**

  ![overleaf界面](/images/overleaf_1.png)

## 3. 常用编辑代码

### 3.1 标题设置

**`\documentclass{article/report/book...}`**:定义文档类型，常见类型包括文章,报告和较长文档,书籍等。

**`\title{}`**：设置文档的标题。

**`\author{}`**：设置作者姓名。

**`\date{}`**：设置日期。使用 `\today` 可自动生成当天日期。

**`\maketitle`**：在文档中插入并显示标题、作者和日期。该命令通常放在 `\begin{document}` 后。

**`\section{Introduction}`**:创建章节，自动编号。

**`\noindent`**:文本不缩进。

```
\documentclass{article}

\title{My First LaTeX Document}
\author{Author Name}
\date{\today}

\begin{document}
\maketitle

\section{Introduction}
\subsection{Background}
\subsubsection{Details}
\noindent This is the introduction section.
\end{document}
```

![标题设置](/images/overleaf_2.png)

### 3.2 插入公式

**行内公式**
用 `$...$` 包裹数学公式，使其在正文中显示。

```
$E = mc^2$
```

**独立公式**
用 `\[ ... \]` 将公式单独放在一行。

```
\[
E = mc^2
\]
```

**带编号公式**
用 `equation` 环境生成带自动编号的公式。

```
\begin{equation}
E = mc^2
\end{equation}
```

**对齐公式**
使用 `align` 环境对齐多行公式。

```
\begin{align}
a &= b + c \\换行
x &= y + z
\end{align}
```

![插入公式](/images/overleaf_3.png)

### 3.3 插入图像/表格

**插入图像**
使用 `graphicx` 宏包插入图片。

```
\usepackage{graphicx}
\begin{figure}[h]
  \centering
  \includegraphics[width=0.5\textwidth]{image.png}
  \caption{Sample Image}
  \label{fig:image1}
\end{figure}
```

**创建表格**
使用 `tabular` 环境创建表格。

```
\begin{tabular}{|c|c|c|}
\hline
Column 1 & Column 2 & Column 3 \\
\hline
Data 1 & Data 2 & Data 3 \\
\hline
\end{tabular}
```

![图像和表格](/images/overleaf_4.png)

### 3.4 引用与参考文献

**创建引用**
使用 `\label{}` 和 `\ref{}` 创建对章节、图表、公式的引用。

```
\label{sec:method}
As explained in Section \ref{sec:method}, we...
```

**插入参考文献**
使用 `biblatex` 或 `natbib` 宏包管理参考文献。

```
\usepackage{biblatex}
\addbibresource{references.bib}

\cite{key}  % 引用文献
\printbibliography  % 打印参考文献
```

### 3.5 超链接和脚注

**插入超链接**
使用 `hyperref` 宏包创建超链接。

```
\usepackage{hyperref}
For more information, visit \href{https://www.example.com}{this website}.
```

**插入脚注**
使用 `\footnote{}` 插入脚注。

```
This is a statement.\footnote{This is a footnote.}
```

![超链接和脚注](/images/overleaf_5.png)

### 3.6 插入列表

**无序列表**
使用 `itemize` 环境创建无序列表。

```
\begin{itemize}
  \item First item
  \item Second item
  \item Third item
\end{itemize}
```

**有序列表**
使用 `enumerate` 环境创建有序列表。

```
\begin{enumerate}
  \item First item
  \item Second item
  \item Third item
\end{enumerate}
```

![插入列表](/images/overleaf_6.png)

这些命令和功能能帮助您在 Overleaf 上高效编写 LaTeX 文档，利用 Overleaf 的实时预览和协作功能，可以更加直观和便捷地管理项目。

***注：以上内容均为快速入门内容，如想深入学习请访问Overleaf官方文档：***[Overleaf教程](https://cn.overleaf.com/learn)


