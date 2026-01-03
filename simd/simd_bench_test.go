package main

import (
	"encoding/json"
	"mcompiler/arena"
	"testing"
)

var jsonBytes = []byte(`{
	"statuses": [
		{
			"metadata": { "result_type": "recent", "iso_language_code": "ja" },
			"created_at": "Sun Aug 31 00:29:15 +0000 2014",
			"id": 505874924095815681,
			"id_str": "505874924095815681",
			"text": "Slack is awesome. Using arena allocator makes it faster.",
			"source": "<a href=\"http://twitter.com/download/iphone\" rel=\"nofollow\">Twitter for iPhone</a>",
			"truncated": false,
			"user": {
				"id": 2244994945,
				"name": "Gemini User",
				"screen_name": "gemini_dev",
				"followers_count": 142,
				"friends_count": 1833,
				"favourites_count": 10245,
				"is_translator": false
			},
			"retweet_count": 0,
			"favorite_count": 0,
			"favorited": false,
			"retweeted": false,
			"lang": "en",
			"coordinates": null
		}
	]
}`)

var largeJsonBytes []byte

func init() {
	// Construct a large JSON (~1MB) by repeating the tweet object
	tweetStr := `{
			"metadata": { "result_type": "recent", "iso_language_code": "ja" },
			"created_at": "Sun Aug 31 00:29:15 +0000 2014",
			"id": 505874924095815681,
			"id_str": "505874924095815681",
			"text": "Slack is awesome. Using arena allocator makes it faster.",
			"source": "<a href=\"http://twitter.com/download/iphone\" rel=\"nofollow\">Twitter for iPhone</a>",
			"truncated": false,
			"user": {
				"id": 2244994945,
				"name": "Gemini User",
				"screen_name": "gemini_dev",
				"followers_count": 142,
				"friends_count": 1833,
				"favourites_count": 10245,
				"is_translator": false
			},
			"retweet_count": 0,
			"favorite_count": 0,
			"favorited": false,
			"retweeted": false,
			"lang": "en",
			"coordinates": null
		}`

	// Create ~1000 items -> ~500KB - 1MB range
	var buf []byte
	buf = append(buf, []byte(`{"statuses": [`)...)
	for i := 0; i < 2000; i++ {
		buf = append(buf, []byte(tweetStr)...)
		if i < 1999 {
			buf = append(buf, ',')
		}
	}
	buf = append(buf, []byte(`]}`)...)
	largeJsonBytes = buf
}

func BenchmarkStdJSON_Map(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var m map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &m); err != nil {
			b.Fatal(err)
		}
	}
}

type Tweet struct {
	Statuses []struct {
		Text string `json:"text"`
		User struct {
			ID int64 `json:"id"`
		} `json:"user"`
	} `json:"statuses"`
}

func BenchmarkStdJSON_Struct(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var t Tweet
		if err := json.Unmarshal(jsonBytes, &t); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFastParser(b *testing.B) {
	a := arena.NewBestArena()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Reset()

		p := NewParser(jsonBytes, a)
		_ = p.ParseAny()
	}
}

func BenchmarkStdJSON_Map_Large(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var m map[string]interface{}
		if err := json.Unmarshal(largeJsonBytes, &m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStdJSON_Struct_Large(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var t Tweet
		if err := json.Unmarshal(largeJsonBytes, &t); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFastParser_Large(b *testing.B) {
	a := arena.NewBestArena()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Reset()
		p := NewParser(largeJsonBytes, a)
		_ = p.ParseAny()
	}
}

// --------------------------------------------------------------------------
// String Scanning Micro-Benchmarks
// --------------------------------------------------------------------------

// 1. String with escapes (Worst case for naive scanning, best for SIMD single-pass)
// 100 chars, escape at the end
var stringBenchData = []byte(`"This is a relatively long string that has an escaped quote \" right here to test the scanning logic."`)

func BenchmarkFindClosingQuoteLength(b *testing.B) {
	// Setup parser with just this string
	a := arena.NewBestArena()
	p := NewParser(stringBenchData, a)
	// Point cursor to start of quote + 1 (as if ParseString did p.cursor++)
	p.cursor = 1
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset cursor simulation
		p.cursor = 1
		_ = p.findClosingQuoteLength()
	}
}

func BenchmarkScanStringBoundary(b *testing.B) {
	a := arena.NewBestArena()
	p := NewParser(stringBenchData, a)
	p.cursor = 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.cursor = 1
		_, _ = p.scanStringBoundary()
	}
}
