package pirsch

// Contains all substrings used by bots in the User-Agent header.
// Note that user agents containing a link or the keywords "bot", "crawler", "spider" don't need to be included,
// as they are filtered out.
var userAgentBotList = []string{
	"zgrab",
}
