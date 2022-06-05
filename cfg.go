package cfg

import (
	"errors"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/jayacarlson/dbg"
	"github.com/jayacarlson/txt"
)

type (
	ConfigType int
)

const (
	ConfigBlock ConfigType = iota
	ConfigLines
	ConfigItems
	ConfigValue
)

var (
	ErrIllegalDataBlock = errors.New("Illegal ConfigData() -- no leading TAB")

	// 1: label  2: ,  3: <|[|{|(  4: .*
	findConfigStRex = regexp.MustCompile(`(?ms).*?^(\w+) *(,)* *(<|\[|{|\() *\n(.*)`)
	// 1: -contents-  2: >|]|}|)  3: .*
	findConfigEnRex = regexp.MustCompile(`(?ms)(.*?)\n^(>|\]|}|\))((\n|$).*)`)

	matching = map[string]string{
		"<": ">", // data block
		"[": "]", // data lines
		"{": "}", // data items
		"(": ")", // data : wraps (), <>, [], {}
	}

	// label := value
	// 1: label  2: value  3: remaining
	findConfigValueRex = regexp.MustCompile(`(?ms).*?^(\w+) *:= *(.*?) *\n(.*)`)

	// label < ... >
	// 1: label  2: -blockData-  3: remaining
	findConfigBlockRex = regexp.MustCompile(`(?ms).*?^(\w+) *<\n(.*?)\n>((\n|$).*)`)

	// label [ ... ]
	// 1: label  2: -lineData-  3: remaining
	findConfigLinesRex = regexp.MustCompile(`(?ms).*?^(\w+) *\[\n(.*?)\n\]((\n|$).*)`)

	// label , { ... }
	// 1: label  2: ,  3: -listData-  4: remaining  -- #2 may be empty or a comma
	findConfigItemsRex = regexp.MustCompile(`(?ms).*?^(\w+) *(,)* *{\n(.*?)\n}((\n|$).*)`)

	// 1: label 2: remaining
	dictRex = regexp.MustCompile(`^(\w+) *: *(.*)`)
)

func StringListToDict(l []string) map[string]string {
	result := make(map[string]string)
	for _, v := range l {
		if x := dictRex.FindStringSubmatch(v); x != nil {
			result[x[1]] = x[2]
		}
	}
	return result
}

/*
	Scan config data looking for 'label := value' pairs

	The 'label' can be made of any word chars (letter, number or underscore)

	The 'label' must start the line and be followed by ':='

	The 'value' is the rest of the line with leading / trailing whitespace
	removed.  The 'value' could be empty, e.g. "label :="
*/
func HandleConfigValues(str string, f func(label, value string)) {
	for x := findConfigValueRex.FindStringSubmatch(str); x != nil; x = findConfigValueRex.FindStringSubmatch(x[3]) {
		f(x[1], x[2])
	}
}

/*
	Scan configuration information looking for a 'label' and an area of text
	 surrounded by the < & > characters

	The 'label' can be made of any word chars (letter, number or underscore)

	The 'label' must start the line with the < the only text on the line

	The ending > must be 1st and only character on the closing line

	On finding a config block, the supplied function is called with the label
	 and all the block data

	All text between the < & > characters is captured, including leading/trailing
	whitespace; note the last \n before the ending > is consumed

	e.g. given some sample text of:

		blockData1<
		blah blah
		blah blah
		>

		blockData2 <
			# block data
			more stuff...
		>

	Callback func would be called with ("blockData1", "blah blah\nblah blah")
	Then called with ("blockData2", "\t# block data\n\tmore stuff...")
*/
func HandleConfigBlocks(str string, f func(label, block string)) {
	for x := findConfigBlockRex.FindStringSubmatch(str); x != nil; x = findConfigBlockRex.FindStringSubmatch(x[3]) {
		f(x[1], x[2])
	}
}

/*
	Scan configuration information looking for a 'label' and an area of text
	 surrounded by [brackets]

	The 'label' can be made of any word chars (letter, number or underscore)

	The 'label' must start the line with the [ the only text on the line

	The ending ] must be 1st and only character on the closing line

	On finding config lines, the supplied function is called with the label
	 and line data

	Whitespace around the text is stripped
	Empty lines are ignored
	Any lines starting with '#' inside the config area are also removed

	e.g. given some sample text of:

		lineData1[
			# line data...
			data1.1

			data1.2
		]

		lineData2 [
			foo: bar, boo
			goo, faz: gar
		]

	Callback func would be called with ("lineData1", []string{"data1.1","data1.2"})
	Then called with ("lineData2", []string{"foo: bar, boo","goo, faz: gar"})
*/
func HandleConfigLines(str string, f func(label string, lines []string)) {
	for x := findConfigLinesRex.FindStringSubmatch(str); x != nil; x = findConfigLinesRex.FindStringSubmatch(x[3]) {
		f(x[1], txt.ListToStringSlice(x[2]))
	}
}

/*
	Scan configuration information looking for a 'label' and an area of text
	 surrounded by {braces}

	The 'label' can be made of any word chars (letter, number or underscore)

	The 'label' must start the line with the { the only text on the line;
	 although an optional comma (,) can follow the label to be used as
	 the list item separator

	The ending ] must be 1st and only character on the closing line

	The list data contained inside the braces will be seperated into a slice of
	 strings by either the optional comma or by spaces

	On finding config items, the supplied function is called with the label
	 and list data

	Whitespace around the text is stripped
	Empty lines are ignored
	Any lines starting with '#' inside the config area are also removed

	e.g. given some sample text of:

		listData1{
			# list data
			item1.1

			item1.2
		}

		listData2 , {
			item2.1, item2.2
		}

	Callback func would be called with ("listData1", []string{"item1.1","item1.2"})
	Then called with ("listData2", []string{"item2.1","item2.2"})
*/
func HandleConfigItems(str string, f func(label string, list []string)) {
	for x := findConfigItemsRex.FindStringSubmatch(str); x != nil; x = findConfigItemsRex.FindStringSubmatch(x[4]) {
		if "" == x[2] {
			x[2] = " "
		}
		f(x[1], txt.SepListToStringSlice(x[3], x[2]))
	}
}

/*
	Config Data allows grouping of data...

	Scan configuration information, first looking for any ConfigValues; then
	 looking for a 'label' and an area of text surrounded by (parens),
	 {braces}, [brackets] or < & >

	For config data wrapped by (parens), all lines MUST HAVE a leading \t or be a
	 single \n which is then removed.  The 'label' is added to a labelPath; this
	 data is then fed back into the HandleConfigData.

	The labelPath generated is the list of data groups separated with a ':' with
	 the final data info name appended. e.g.

	 	data(
	 		label := value
	 		subdata (
	 			listData , {
	 				alpha, beta, delta
	 			}
	 		)
	 	)

	On finding config information, the supplied function is called with the
	 labelPath and the final blockdata, linedata or listdata.  For the example
	 above, the callback would be called with:
	   f( ConfigValue, "data:label", []string{"value"} )
	    then
	   f( ConfigItems, "data:subdata:listData", []string{"alpha", "beta", "delta"} )
*/
func HandleConfigData(str string, f func(t ConfigType, label string, data []string)) error {
	return handleConfigData("", str, f)
}

// ------------------------------------------------------------------------- //

/*
	Reads the config file and passes returned name := value pairs to handler
*/
func LoadConfigValues(flPath string, f func(label, value string)) {
	data, err := ioutil.ReadFile(flPath)
	dbg.ChkErrX(err, "Failed to read config file: %s (%v)", flPath, err)
	HandleConfigValues(string(data), f)
}

/*
	Reads the config file and passes returned block data to handler
*/
func LoadConfigBlocks(flPath string, f func(label, data string)) {
	data, err := ioutil.ReadFile(flPath)
	dbg.ChkErrX(err, "Failed to read config file: %s (%v)", flPath, err)
	HandleConfigBlocks(string(data), f)
}

/*
	Reads the config file and passes returned line data to handler
*/
func LoadConfigLines(flPath string, f func(label string, data []string)) {
	data, err := ioutil.ReadFile(flPath)
	dbg.ChkErrX(err, "Failed to read config file: %s (%v)", flPath, err)
	HandleConfigLines(string(data), f)
}

/*
	Reads the config file and passes returned list data to handler
*/
func LoadConfigItems(flPath string, f func(label string, data []string)) {
	data, err := ioutil.ReadFile(flPath)
	dbg.ChkErrX(err, "Failed to read config file: %s (%v)", flPath, err)
	HandleConfigItems(string(data), f)
}

/*
	Reads the config file and passes returned data to handler
*/
func LoadConfigData(flPath string, f func(t ConfigType, label string, data []string)) error {
	data, err := ioutil.ReadFile(flPath)
	dbg.ChkErrX(err, "Failed to read config file: %s (%v)", flPath, err)
	return handleConfigData("", string(data), f)
}

// ------------------------------------------------------------------------- //

func removeLeadingTabs(src string) (string, error) {
	result := ""
	if len(src) == 0 {
		return "", nil
	}
	for len(src) > 0 {
		c, i := src[0], strings.Index(src, "\n")
		if i > 0 && c == '\t' {
			result += src[1 : i+1]
		} else if i == 0 && c == '\n' {
			// because blank lines can be inside <blockdata> retain the blank line
			result += "\n"
		} else {
			return "", ErrIllegalDataBlock
		}
		src = src[i+1:]
	}
	return result, nil
}

func handleConfigData(lp, str string, f func(t ConfigType, label string, data []string)) error {
	for "" != str {
		// find any ConfigValues first
		HandleConfigValues(str, func(l, v string) {
			if lp != "" {
				l = lp + ":" + l
			}
			f(ConfigValue, l, []string{v})
		})
		s := findConfigStRex.FindStringSubmatch(str)
		if nil != s {
			e := findConfigEnRex.FindStringSubmatch(s[4])
			if nil == e {
				dbg.Error("Missing end char for config data: %s %s", s[1], s[3])
				break
			}
			if e[2] != matching[s[3]] {
				dbg.Error("Invalid end char for config data: %s %s ... %s", s[1], s[3], e[2])
				break
			}
			if "{" != s[3] && "" != s[2] {
				dbg.Error("Illegal config data: %s %s -- comma", s[1], s[3])
				break
			}
			lblPath := s[1]
			if lp != "" {
				lblPath = lp + ":" + lblPath
			}
			switch s[3] {
			case "(":
				st, err := removeLeadingTabs(e[1] + "\n")
				if nil == err {
					err = handleConfigData(lblPath, st, f)
				}
				if nil != err {
					return err
				}
			case "<":
				f(ConfigBlock, lblPath, []string{e[1]})
			case "[":
				f(ConfigLines, lblPath, txt.ListToStringSlice(e[1]))
			default: //case "{":
				if "" == s[2] {
					s[2] = " "
				}
				f(ConfigItems, lblPath, txt.SepListToStringSlice(e[1], s[2]))
			}
			str = e[3]
		} else {
			str = ""
		}
	}
	return nil
}
