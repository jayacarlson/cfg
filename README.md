# cfg
Utility to read in configuration information of different types

**cfg** allows you to create configuration files that can be read in at run-time with different types of configuration information being possible.

### Config Values:  Simple name:value pairs
```x
label1 := value1
label2 := value2
label3 := value3
```
### Config Blocks:  All text contained inside a block surrounded by < & >
```x
blockData <
all this is part
  of the text block

  including any leading
or trailing whitespace, 
the blank line above,
 and what would...
# normally be a comment
>
```
### Config Lines:  Individual lines of text contained inside a block surrounded by [ & ]
```x
lineData [
a line of text
  another line, note that leading/trailing whitespace is removed
    # a comment line like this, and the blank line below

are removed and are not a part of the line data
]
```
### Config Items:  Individual tems contained inside a block surrounded by { & }
```x
itemData {
item1 item2 item3
  item4   item5   item6
# comment lines are ignored
# items have leading/trailing whitespace removed
}

itemData , {
# a comma can be used to change the separator to allow
# whitespace as part of the item, note the change to the block identifier
item 1, item 2, item 3
  item 4,   item 5,   item 6,
# comment lines are ignored
# items have leading/trailing whitespace removed
}
```
### Config Data:  A hierarchical container of all of the above for data grouping, contained inside a block surrounded by ( & )
```x
dataContainer (
	# note that all lines inside a data container must have a leading TAB
	# which is removed before recursing back through the data looking for:
	
	name := value pairs

	subBlock <
	a block of text
	>

	subLines [
		lines of text...
	]

	subItems {
		list of items...
	}

	subData (
		name := value pairs
		# more blocks/lines/items/data
	)
)
```
