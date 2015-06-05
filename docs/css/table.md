# Table
---

使用时注意 `<table>` HTML 结构的完整性。

**表格相关 JS 插件：**
- [jQuery DataTables](https://github.com/amazeui/datatables)

## 基本样式

添加 `.am-table`。

`````html
<table class="am-table">
	<thead>
		<tr>
			<th>网站名称</th>
			<th>网址</th>
			<th>创建时间</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr class="am-active">
			<td>Amaze UI(Active)</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
    <tr>
      <td>Amaze UI</td>
      <td>http://amazeui.org</td>
      <td>2012-10-01</td>
    </tr>
	</tbody>
  <tbody>
  <tr>
    <td>Amaze UI</td>
    <td>http://amazeui.org</td>
    <td>2012-10-01</td>
  </tr>
  </tbody>
</table>
`````

```html
<table class="am-table">
	<thead>
		<tr>
			<th>网站名称</th>
			<th>网址</th>
			<th>创建时间</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr class="am-active">
			<td>Amaze UI(Active)</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
	</tbody>
</table>
```

## 基本边框

添加 `.am-table-bordered` 类。

`````html
<table class="am-table am-table-bordered">
	<thead>
		<tr>
			<th>网站名称</th>
			<th>网址</th>
			<th>创建时间</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
	</tbody>
</table>
`````

```html
<table class="am-table am-table-bordered">
	...
</table>
```

## 圆角边框

同时添加 `.am-table-bordered` 、 `.am-table-radius`，外层圆角边框通过 `box-shadow` 实现。

`````html
<table class="am-table am-table-bordered am-table-radius am-table-striped">
	<thead>
		<tr>
			<th>网站名称</th>
			<th>网址</th>
			<th>创建时间</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
	</tbody>
</table>
`````

```html
<table class="am-table am-table-bordered am-table-radius am-table-striped">
	...
</table>
```

## 单元格状态

表示表格状态的 class 添加到 `tr` 整行整行，添加到 `td` 高亮单元格。

- `.am-active` 激活；
- `.am-disabled` 禁用；
- `.am-primary` 蓝色高亮；
- `.am-success` 绿色高亮；
- `.am-warning` 橙色高亮；
- `.am-danger` 红色高亮。

`````html
<table class="am-table am-table-bordered am-table-radius">
  <thead>
  <tr>
    <th>Class</th>
    <th>状态描述</th>
    <th>目标元素</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td>.am-active</td>
    <td class="am-active">激活</td>
    <td>td</td>
  </tr>
  <tr class="am-active">
    <td>.am-active</td>
    <td>激活</td>
    <td>tr</td>
  </tr>
  <tr>
    <td>.am-disabled</td>
    <td class="am-disabled">禁用</td>
    <td>td</td>
  </tr>
  <tr class="am-disabled">
    <td>.am-disabled</td>
    <td>禁用</td>
    <td>tr</td>
  </tr>
  <tr>
    <td class="am-primary">.am-primary</td>
    <td>蓝色高亮</td>
    <td>td</td>
  </tr>
  <tr class="am-primary">
    <td>.am-primary</td>
    <td>蓝色高亮</td>
    <td>tr</td>
  </tr>
  <tr>
    <td class="am-success">.am-success</td>
    <td>绿色高亮</td>
    <td>td</td>
  </tr>
  <tr class="am-success">
    <td>.am-success</td>
    <td>绿色高亮</td>
    <td>tr</td>
  </tr>
  <tr>
    <td class="am-warning">.am-warning</td>
    <td>橙色高亮</td>
    <td>td</td>
  </tr>
  <tr class="am-warning">
    <td>.am-warning</td>
    <td>橙色高亮</td>
    <td>tr</td>
  </tr>
  <tr>
    <td class="am-danger">.am-danger</td>
    <td>橙色高亮</td>
    <td>td</td>
  </tr>
  <tr class="am-danger">
    <td>.am-danger</td>
    <td>橙色高亮</td>
    <td>tr</td>
  </tr>
  </tbody>
</table>
`````

## 其他效果

- `.am-table-striped` 斑马纹效果
- `.am-table-hover` hover 状态

`````html
<table class="am-table am-table-striped am-table-hover">
	<thead>
		<tr>
			<th>网站名称</th>
			<th>网址</th>
			<th>创建时间</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
	</tbody>
</table>
`````

```html
<table class="am-table am-table-striped am-table-hover">
	...
</table>
```

## 紧凑型

添加 `.am-table-compact` class，调整 `padding` 显示更紧凑的单元格。

`````html
<table class="am-table am-table-bordered am-table-striped am-table-compact">
  <thead>
  <tr>
    <th>网站名称</th>
    <th>网址</th>
    <th>创建时间</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td>Amaze UI</td>
    <td>http://amazeui.org</td>
    <td>2012-10-01</td>
  </tr>
  <tr>
    <td>Amaze UI</td>
    <td>http://amazeui.org</td>
    <td>2012-10-01</td>
  </tr>
  <tr class="am-active">
    <td>Amaze UI(Active)</td>
    <td>http://amazeui.org</td>
    <td>2012-10-01</td>
  </tr>
  <tr>
    <td>Amaze UI</td>
    <td>http://amazeui.org</td>
    <td>2012-10-01</td>
  </tr>
  <tr>
    <td>Amaze UI</td>
    <td>http://amazeui.org</td>
    <td>2012-10-01</td>
  </tr>
  </tbody>
</table>
`````

```html
<table class="am-table am-table-bordered am-table-striped am-table-compact">
  <thead>
  <tr>
    <th>网站名称</th>
    <th>网址</th>
    <th>创建时间</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td>Amaze UI</td>
    <td>http://amazeui.org</td>
    <td>2012-10-01</td>
  </tr>
  <tr>
    <td>Amaze UI</td>
    <td>http://amazeui.org</td>
    <td>2012-10-01</td>
  </tr>
  <tr class="am-active">
    <td>Amaze UI(Active)</td>
    <td>http://amazeui.org</td>
    <td>2012-10-01</td>
  </tr>
  <tr>
    <td>Amaze UI</td>
    <td>http://amazeui.org</td>
    <td>2012-10-01</td>
  </tr>
  <tr>
    <td>Amaze UI</td>
    <td>http://amazeui.org</td>
    <td>2012-10-01</td>
  </tr>
  </tbody>
</table>
```

## 响应式表格

- `.am-text-nowrap`: 禁止文字换行；
- `.am-scrollable-horizontal`: 内容超出容器宽度时显示水平滚动条。

以上两个 class 在「**辅助类**」中定义。

`````html
<div class="am-scrollable-horizontal">
  <table class="am-table am-table-bordered am-table-striped am-text-nowrap">
    <thead>
    <tr>
      <th>-= 表格标题 =-</th>
      <th>-= 表格标题 =-</th>
      <th>-= 表格标题 =-</th>
      <th>-= 表格标题 =-</th>
      <th>-= 表格标题 =-</th>
      <th>-= 表格标题 =-</th>
      <th>-= 表格标题 =-</th>
      <th>-= 表格标题 =-</th>
    </tr>
    </thead>
    <tbody>
    <tr>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
    </tr>
    <tr>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
    </tr>
    <tr>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
      <td>表格数据</td>
    </tr>
    </tbody>
  </table>
</div>
`````
```html
<div class="am-scrollable-horizontal">
  <table class="am-table am-table-bordered am-table-striped am-text-nowrap">
    ...
  </table>
</div>
```

## 所有样式叠加

`````html
<table class="am-table am-table-bordered am-table-striped am-table-hover">
	<thead>
		<tr>
			<th>网站名称</th>
			<th>网址</th>
			<th>创建时间</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr class="am-active">
			<td>Amaze UI(Active)</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		<tr>
			<td>Amaze UI</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
	</tbody>
</table>
`````

```html
<table class="am-table am-table-bordered am-table-striped am-table-hover">
	<thead>
		<tr>
			<th>网站名称</th>
			<th>网址</th>
			<th>创建时间</th>
		</tr>
	</thead>
	<tbody>
		...
		<tr class="am-active">
			<td>Amaze UI(Active)</td>
			<td>http://amazeui.org</td>
			<td>2012-10-01</td>
		</tr>
		...
	</tbody>
</table>
```

## 参考资源

- [表格排序 jQuery Table Sort] (https://github.com/kylefox/jquery-tablesort)
- [Tablesaw - A set of jQuery plugins for responsive tables](https://github.com/filamentgroup/tablesaw)
- [FooTable - jQuery plugin to make HTML tables responsive](http://fooplugins.com/plugins/footable-jquery/)
