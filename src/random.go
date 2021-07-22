package src

import (
	"math/rand"

	"github.com/rogpeppe/fastuuid"
)

var UUIDGen = fastuuid.MustNewGenerator()

var SampleReferrers = []string{
	// List of search engines
	"http://www.google.com/",
	"http://www.bing.com/",
	"http://www.yahoo.com/",
	"http://www.baidu.com/",
	"http://www.aol.com/",
	"http://www.ask.com/",
	"http://www.altavista.com/",
	"http://www.live.com/",
	"http://www.msn.com/",

	// List of social networks
	"http://www.facebook.com/",
	"http://www.twitter.com/",
	"http://www.linkedin.com/",
	"http://www.pinterest.com/",
	"http://www.instagram.com/",
	"http://www.youtube.com/",

	// List of news sites
	"http://www.cnn.com/",
	"http://www.bbc.co.uk/",
	"http://www.nytimes.com/",
	"http://www.washingtonpost.com/",
	"http://www.reddit.com/",
	"http://www.huffingtonpost.com/",
	"http://www.theguardian.com/",
	"http://www.theverge.com/",
}

// return a random referrer
func RandomReferrer() string {
	return SampleReferrers[rand.Intn(len(SampleReferrers))]
}

func NextUserId() string {
	return UUIDGen.Hex128()
}
