# Table
---

Please pay attention to the completeness of HTML structure of `<table>`.

**Relative JS plugins: **
- [jQuery DataTables](https://github.com/amazeui/datatables)

## Default Style

Use `.am-table`。

`````html
<table class="am-table">
	<thead>
		<tr>
			<th>Website</th>
			<th>URL</th>
			<th>Created Date</th>
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
      <th>Website</th>
      <th>URL</th>
      <th>Created Date</th>
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
```

## Default Border

Use the `.am-table-bordered` class to create a table with default border.

`````html
<table class="am-table am-table-bordered">
	<thead>
		<tr>
			<th>Website</th>
			<th>URL</th>
			<th>Created Date</th>
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

## Round Cornor

Add `.am-table-bordered` and `.am-table-radius` to create a table with round border. 

`````html
<table class="am-table am-table-bordered am-table-radius am-table-striped">
	<thead>
		<tr>
      <th>Website</th>
      <th>URL</th>
      <th>Created Date</th>
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

## Contextual classes

Add following classes to `tr` to color table rows, and to `td` to color individual cells.

- `.am-active` Activated;
- `.am-disabled` Disabled;
- `.am-primary` Blue Highlight;
- `.am-success` Green Highlight;
- `.am-warning` Orange Highlight;
- `.am-danger` Red Highlight;

`````html
<table class="am-table am-table-bordered am-table-radius">
  <thead>
  <tr>
    <th>Class</th>
    <th>Description</th>
    <th>Target Element</th>
  </tr>
  </thead>
  <tbody>
  <tr>
    <td>.am-active</td>
    <td class="am-active">Activated</td>
    <td>td</td>
  </tr>
  <tr class="am-active">
    <td>.am-active</td>
    <td>Activated</td>
    <td>tr</td>
  </tr>
  <tr>
    <td>.am-disabled</td>
    <td class="am-disabled">Disabled</td>
    <td>td</td>
  </tr>
  <tr class="am-disabled">
    <td>.am-disabled</td>
    <td>Disabled</td>
    <td>tr</td>
  </tr>
  <tr>
    <td class="am-primary">.am-primary</td>
    <td>Blue Highlight</td>
    <td>td</td>
  </tr>
  <tr class="am-primary">
    <td>.am-primary</td>
    <td>Blue Highlight</td>
    <td>tr</td>
  </tr>
  <tr>
    <td class="am-success">.am-success</td>
    <td>Green Highlight</td>
    <td>td</td>
  </tr>
  <tr class="am-success">
    <td>.am-success</td>
    <td>Green Highlight</td>
    <td>tr</td>
  </tr>
  <tr>
    <td class="am-warning">.am-warning</td>
    <td>Orange Highlight</td>
    <td>td</td>
  </tr>
  <tr class="am-warning">
    <td>.am-warning</td>
    <td>Orange Highlight</td>
    <td>tr</td>
  </tr>
  <tr>
    <td class="am-danger">.am-danger</td>
    <td>Red Highlight</td>
    <td>td</td>
  </tr>
  <tr class="am-danger">
    <td>.am-danger</td>
    <td>Red Highlight</td>
    <td>tr</td>
  </tr>
  </tbody>
</table>
`````

## Other styles

- `.am-table-striped` Striped
- `.am-table-hover` hover

`````html
<table class="am-table am-table-striped am-table-hover">
	<thead>
		<tr>
			<th>Website</th>
      <th>URL</th>
      <th>Created Date</th>
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

## Condensed table

Add the `.am-table-compact` class to make tables more compact by adjusting `padding`.

`````html
<table class="am-table am-table-bordered am-table-striped am-table-compact">
  <thead>
  <tr>
    <th>Website</th>
    <th>URL</th>
    <th>Created Date</th>
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
    <th>Website</th>
    <th>URL</th>
    <th>Created Date</th>
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

## Responsive

- `.am-text-nowrap`: Disable wrap.
- `.am-scrollable-horizontal`: Use horizontal scroll when contents overflow.

These two classes is defined in **Utility**.

`````html
<div class="am-scrollable-horizontal">
  <table class="am-table am-table-bordered am-table-striped am-text-nowrap">
    <thead>
    <tr>
      <th>-= Head =-</th>
      <th>-= Head =-</th>
      <th>-= Head =-</th>
      <th>-= Head =-</th>
      <th>-= Head =-</th>
      <th>-= Head =-</th>
      <th>-= Head =-</th>
      <th>-= Head =-</th>
    </tr>
    </thead>
    <tbody>
    <tr>
      <td>Data</td>
      <td>Data</td>
      <td>Data</td>
      <td>Data</td>
      <td>Data</td>
      <td>Data</td>
      <td>Data</td>
      <td>Data</td>
    </tr>
    <tr>
      <td>Data</td>
      <td>Data</td>
      <td>Data</td>
      <td>Data</td>
      <td>Data</td>
      <td>Data</td>
      <td>Data</td>
      <td>Data</td>
    </tr>
    <tr>
      <td>Data</td>
      <td>Data</td>
      <td>Data</td>
      <td>Data</td>
      <td>Data</td>
      <td>Data</td>
      <td>Data</td>
      <td>Data</td>
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

## Update

### New features in 2.4.x

- Add `.am-table-centered` to `<table>` to apply center align to all the cells.
- Add `.am-text-middle` to make the content display in the middle of cell. More detail in [Utilily](http://amazeui.org/css/utility?_ver=2.x#wen-ben-zuo-you-dui-qi)）

`````html
<table class="am-table am-table-bordered am-table-centered">
  <tr>
    <th>Savings for holiday!</th>
    <th>Month</th>
    <th>Savings</th>
  </tr>
  <tr>
    <td rowspan="2" class="am-text-middle">$50</td>
    <td>January</td>
    <td>$100</td>
  </tr>
  <tr>
    <td>February</td>
    <td>$80</td>
  </tr>
</table>
`````
```html
<table class="am-table am-table-bordered am-table-centered">
  <tr>
    <th>Savings for holiday!</th>
    <th>Month</th>
    <th>Savings</th>
  </tr>
  <tr>
    <td rowspan="2" class="am-text-middle">$50</td>
    <td>January</td>
    <td>$100</td>
  </tr>
  <tr>
    <td>February</td>
    <td>$80</td>
  </tr>
</table>
```

## Combination of All Styles

`````html
<table class="am-table am-table-bordered am-table-striped am-table-hover">
	<thead>
		<tr>
			<th>Website</th>
      <th>URL</th>
      <th>Created Date</th>
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
			<th>Website</th>
      <th>URL</th>
      <th>Created Date</th>
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

## Reference

- [Table Sort - jQuery Table Sort] (https://github.com/kylefox/jquery-tablesort)
- [Tablesaw - A set of jQuery plugins for responsive tables](https://github.com/filamentgroup/tablesaw)
- [FooTable - jQuery plugin to make HTML tables responsive](http://fooplugins.com/plugins/footable-jquery/)
