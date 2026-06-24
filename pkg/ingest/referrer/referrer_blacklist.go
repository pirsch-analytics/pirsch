package referrer

var referrerBlacklist = []string{
	" and ",
	" or ",
	" xor ",
	"/bin/cat",
	"/etc/passwd",
	"content-type",
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
