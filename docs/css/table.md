# Table
---

使用时注意 `<table>` HTML 结构的完整性。

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

添加 `.am-table-bd` 类。

`````html
<table class="am-table am-table-bd">
	<thead>
		<tr>
			<th>网站名称</th>
			<th>网址</th>
			<th>创建时间</th>
		</tr>
	</thead>
	<tdody>
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
	</tody>
</table>
`````

```html
<table class="am-table am-table-bd">
	...
</table>
```

## 圆角边框

同时添加 `.am-table-bd` 、 `.am-table-bdrs`，外层圆角边框通过 `box-shadow` 实现。

`````html
<table class="am-table am-table-bd am-table-bdrs am-table-striped">
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
<table class="am-table am-table-bd am-table-bdrs am-table-striped">
	...
</table>
```

## 其他效果

- `.am-table-striped` 斑马纹效果
- `.am-table-hover` hover 状态
- `.am-active` 添加到 `tr`，激活状态

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

## 所有样式叠加

`````html
<table class="am-table am-table-bd am-table-striped am-table-hover">
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
<table class="am-table am-table-bd am-table-striped am-table-hover">
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