package referrer

var referrerBlacklist = []string{
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
