package referrer

// Blacklist is a list of referrer keywords to ignore.
var Blacklist = []string{
	" and ",
	" or ",
	" xor ",
	"for (",
	"for(",
	"from (",
	"from(",
	"pg_sleep",
	"select (",
	"select(",
	"sleep (",
	"sleep(",
	"sysdate",
	"while (",
	"while(",
}
