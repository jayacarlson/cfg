package cfg

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/jayacarlson/dbg"
	"github.com/jayacarlson/pth"
)

const (
	blk1 = `block1 blah blah

blah blah blah`
	blk2 = `block2 blah blah
	
blah blah blah`
	blk3 = `block3 blah blah
blah blah blah`
	blk4 = `block4 blah blah
# this is included in the block!
blah blah blah`
	dictTestList = `
dictTest [
	# comment
	alpha : apple sauce
	beta: banana bread

	delta :date soup
]`
	valueTests = `
apple := tree
banana := plant
blank :=
cherry := berry
cashew := nut
`
)

var (
	itm1 = []string{"apple", "banana", "cherry", "date", "{-}", "}-{", "fig", "grape"}
	itm2 = []string{"apple", "banana", "cherry", "date", "fig", "grape"}
	lst1 = []string{"apple", "banana cherry", "date", "[ ]", "] [", "fig grape"}
	lst2 = []string{"apple banana", "cherry", "date", "fig   grape"}

	conf []byte
)

func init() {
	cnf, err := ioutil.ReadFile(pth.AsRealPath("$/testdata/testBlocks.cfg"))
	dbg.ChkErrX(err, "Failed to read config file")
	conf = cnf
}

func compareEntries(a, b []string) bool {
	if len(a) != len(b) {
		dbg.Error("Mismatch in slice lengths")
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			dbg.Error("Mismatch in entries: a(%s)  b(%s)", a[i], b[i])
			return false
		}
	}

	return true
}

func TestConfigBlocks(t *testing.T) {
	HandleConfigBlocks(string(conf), func(l, d string) {
		switch l {
		case "block1":
			if d != blk1 {
				dbg.Error("block1<\n%s\n>", d)
				t.Fail()
			}
		case "block2":
			if d != blk2 {
				dbg.Error("block2<\n%s\n>", d)
				t.Fail()
			}
		case "block3":
			if d != blk3 {
				dbg.Error("block3<\n%s\n>", d)
				t.Fail()
			}
		case "block4":
			if d != blk4 {
				dbg.Error("block4<\n%s\n>", d)
				t.Fail()
			}
		default:
			if l != "unknownBlock" {
				dbg.Error("Unknown block: %s", l)
				t.Fail()
			}
		}
	})
}

func TestConfigLines(t *testing.T) {
	HandleConfigLines(string(conf), func(l string, d []string) {
		switch l {
		case "lines1":
			if !compareEntries(lst1, d) {
				dbg.Info("%v", d)
				t.Fail()
			}
		case "lines2":
			if !compareEntries(lst2, d) {
				dbg.Info("%v", d)
				t.Fail()
			}
		default:
			if l != "unknownBlock" {
				dbg.Error("Unknown block: %s", l)
				t.Fail()
			}
		}
	})
}

func TestConfigItems(t *testing.T) {
	HandleConfigItems(string(conf), func(l string, d []string) {
		switch l {
		case "items1":
			if !compareEntries(itm1, d) {
				dbg.Info("%v", d)
				t.Fail()
			}
		case "items2":
			if !compareEntries(itm2, d) {
				dbg.Info("%v", d)
				t.Fail()
			}
		default:
			if l != "unknownBlock" {
				dbg.Error("Unknown block: %s", l)
				t.Fail()
			}
		}
	})
}

func TestConfigData(t *testing.T) {
	err := HandleConfigData(string(conf), func(ctp ConfigType, l string, d []string) {
		switch l {
		case "testData:apple":
			if ctp != ConfigValue || d[0] != "tree" {
				t.Fail()
			}
		case "testData:blocks:banana":
			if ctp != ConfigValue || d[0] != "plant" {
				t.Fail()
			}
		case "testData:lists:cherry":
			if ctp != ConfigValue || d[0] != "berry" {
				t.Fail()
			}
		case "testData:lines:cashew":
			if ctp != ConfigValue || d[0] != "nut" {
				t.Fail()
			}

		case "testData:blocks:block1":
			if ctp != ConfigBlock || d[0] != blk1 {
				dbg.Error("block1<\n%s\n>", d[0])
				t.Fail()
			}
		case "testData:blocks:block2":
			if ctp != ConfigBlock || d[0] != blk2 {
				dbg.Error("block2<\n%s\n>", d[0])
				t.Fail()
			}
		case "testData:blocks:block3":
			if ctp != ConfigBlock || d[0] != blk3 {
				dbg.Error("block3<\n%s\n>", d[0])
				t.Fail()
			}

		case "testData:lines:lines1":
			if ctp != ConfigLines || !compareEntries(lst1, d) {
				dbg.Error("%s: %v", l, d)
				t.Fail()
			}
		case "testData:lines:lines2":
			if ctp != ConfigLines || !compareEntries(lst2, d) {
				dbg.Error("%s: %v", l, d)
				t.Fail()
			}

		case "testData:lists:items1":
			if ctp != ConfigItems || !compareEntries(itm1, d) {
				dbg.Error("%s: %v", l, d)
				t.Fail()
			}
		case "testData:lists:items2":
			if ctp != ConfigItems || !compareEntries(itm2, d) {
				dbg.Error("%s: %v", l, d)
				t.Fail()
			}

		default:
			i := strings.Index(l, ":")
			if i > 0 {
				if !strings.HasSuffix(l, ":unknownBlock") {
					dbg.Error("Unknown label %s", l)
					t.Fail()
				}
			}
		}
	})
	if nil != err {
		dbg.Error(err.Error())
		t.Fail()
	}
}

func TestConfigValues(t *testing.T) {
	HandleConfigValues(valueTests, func(l, v string) {
		switch l {
		case "apple":
			if "tree" != v {
				t.Fail()
			}
		case "banana":
			if "plant" != v {
				t.Fail()
			}
		case "blank":
			if "" != v {
				t.Fail()
			}
		case "cherry":
			if "berry" != v {
				t.Fail()
			}
		case "cashew":
			if "nut" != v {
				t.Fail()
			}
		}
	})
}

func TestStringListDict(t *testing.T) {
	HandleConfigLines(dictTestList, func(l string, d []string) {
		if l == "dictTest" {
			dict := StringListToDict(d)
			if len(dict) != 3 {
				dbg.Error("Miscount in dict: %d", len(dict))
				t.Fail()
			} else {
				if dict["alpha"] != "apple sauce" {
					dbg.Error("Invalid entry, alpha: %s", dict["alpha"])
					t.Fail()
				}
				if dict["beta"] != "banana bread" {
					dbg.Error("Invalid entry, beta: %s", dict["beta"])
					t.Fail()
				}
				if dict["delta"] != "date soup" {
					dbg.Error("Invalid entry, delta: %s", dict["delta"])
					t.Fail()
				}
			}
		} else {
			dbg.Error("Unknown label %s", l)
			t.Fail()
		}
	})
}
